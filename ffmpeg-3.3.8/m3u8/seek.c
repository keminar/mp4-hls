/*
 * Copyright (c) 2013 Stefano Sabatini
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

/**
 * @file
 * libavformat/libavcodec demuxing and muxing API example.
 *
 * Remux streams from one container format to another.
 * @example remuxing.c
 */

#include <libavutil/timestamp.h>
#include <libavformat/avformat.h>
#include "libavutil/avstring.h"

struct options_t {
    const char *input_file;
    const char *output_prefix;
    const char *url_prefix;
};

int main(int argc, char **argv)
{
    AVOutputFormat *ofmt = NULL;
    AVFormatContext *ifmt_ctx = NULL, *ofmt_ctx = NULL;
    AVPacket pkt;
    int ret, i;
    int stream_index = 0;
    int *stream_mapping = NULL;
    int stream_mapping_size = 0;

    char *out_filename;
    double prev_segment_time = 0;
    double tmp_segment_time;
    unsigned int output_index = 1; //分片名
    struct options_t options;
    options.url_prefix = "";
    int video_first = 0;
    int first_segment = 0;
    int64_t segment_pts = 0; //快进到此位置

    if (argc < 5) {
        printf("usage: %s input output\n"
               "API example program to remux a media file with libavformat and libavcodec.\n"
               "The output format is guessed according to the file extension.\n"
               "\n", argv[0]);
        return 1;
    }

    options.input_file  = argv[1];
    options.output_prefix = argv[2];

    segment_pts = (int64_t)strtol(argv[3], NULL, 10);
    output_index = (unsigned int)strtol(argv[4], NULL, 10);

    out_filename = malloc(sizeof(char) * (strlen(options.output_prefix) + 20));
    if (!out_filename) {
        fprintf(stderr, "Could not allocate space for output filenames\n");
        exit(1);
    }

    av_register_all();
    av_log_set_level(AV_LOG_ERROR);

    if ((ret = avformat_open_input(&ifmt_ctx, options.input_file, 0, 0)) < 0) {
        fprintf(stderr, "Could not open input file '%s'", options.input_file);
        goto end;
    }

    if ((ret = avformat_find_stream_info(ifmt_ctx, 0)) < 0) {
        fprintf(stderr, "Failed to retrieve input stream information");
        goto end;
    }

    av_dump_format(ifmt_ctx, 0, options.input_file, 0);

    ofmt = av_guess_format("mpegts", NULL, NULL);
    if (!ofmt) {
       fprintf(stderr, "Could not find MPEG-TS muxer\n");
       goto end;
    }

    ofmt_ctx = avformat_alloc_context();
    if (!ofmt_ctx) {
       fprintf(stderr, "Could not allocated output context");
       goto end;
    }
    ofmt_ctx->oformat = ofmt;

    stream_mapping_size = ifmt_ctx->nb_streams;
    stream_mapping = av_mallocz_array(stream_mapping_size, sizeof(*stream_mapping));
    if (!stream_mapping) {
        ret = AVERROR(ENOMEM);
        goto end;
    }

    for (i = 0; i < ifmt_ctx->nb_streams; i++) {
        AVStream *out_stream;
        AVStream *in_stream = ifmt_ctx->streams[i];
        AVCodecParameters *in_codecpar = in_stream->codecpar;

        if (in_codecpar->codec_type != AVMEDIA_TYPE_AUDIO &&
            in_codecpar->codec_type != AVMEDIA_TYPE_VIDEO &&
            in_codecpar->codec_type != AVMEDIA_TYPE_SUBTITLE) {
            stream_mapping[i] = -1;
            continue;
        }

        stream_mapping[i] = stream_index++;

        out_stream = avformat_new_stream(ofmt_ctx, NULL);
        if (!out_stream) {
            fprintf(stderr, "Failed allocating output stream\n");
            ret = AVERROR_UNKNOWN;
            goto end;
        }

        ret = avcodec_parameters_copy(out_stream->codecpar, in_codecpar);
        if (ret < 0) {
            fprintf(stderr, "Failed to copy codec parameters\n");
            goto end;
        }
        out_stream->codecpar->codec_tag = 0;
    }

    snprintf(out_filename, strlen(options.output_prefix) + 20, "%s-%ld-%u.ts", options.output_prefix, segment_pts, output_index);

    if (!(ofmt->flags & AVFMT_NOFILE)) {
        ret = avio_open(&ofmt_ctx->pb, out_filename, AVIO_FLAG_WRITE);
        if (ret < 0) {
            fprintf(stderr, "Could not open output file '%s'", out_filename);
            goto end;
        }
    }

    ret = avformat_write_header(ofmt_ctx, NULL);

    if (ret < 0) {
        fprintf(stderr, "Error occurred when opening output file\n");
        goto end;
    }

    if (segment_pts != 0) {
        ret = av_seek_frame(ifmt_ctx, 0,  segment_pts, AVSEEK_FLAG_ANY);
        if (ret < 0) {
            fprintf(stderr, "Error could not seek to position\n");
            goto end;
        }
    }

    while (1) {
        AVStream *in_stream, *out_stream;

        ret = av_read_frame(ifmt_ctx, &pkt);
        if (ret < 0)
            break;

        //fprintf(stderr, "helo stream_index %d\n", pkt.stream_index);
        in_stream  = ifmt_ctx->streams[pkt.stream_index];
        if (pkt.stream_index >= stream_mapping_size ||
            stream_mapping[pkt.stream_index] < 0) {
            av_packet_unref(&pkt);
            continue;
        }

        pkt.stream_index = stream_mapping[pkt.stream_index];
        out_stream = ofmt_ctx->streams[pkt.stream_index];

        if (first_segment == 0) {
            first_segment = 1;
            prev_segment_time = pkt.pts * av_q2d(in_stream->time_base);
        }
        tmp_segment_time = pkt.pts * av_q2d(in_stream->time_base);
        if (in_stream->codecpar->codec_type == AVMEDIA_TYPE_VIDEO && (pkt.flags & AV_PKT_FLAG_KEY)) {
            if (video_first == 0) { //第一个片
                video_first = 1;
            } else {
                if (tmp_segment_time - prev_segment_time >= 2) { // 几秒一个分隔
                    //fprintf(stderr, "helo %d,  %f - %f = %f\n", output_index, segment_time, prev_segment_time, (segment_time - prev_segment_time));
                    goto finish;
                }
            }
        }

        /* copy packet */
        pkt.pts = av_rescale_q_rnd(pkt.pts, in_stream->time_base, out_stream->time_base, AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX);
        pkt.dts = av_rescale_q_rnd(pkt.dts, in_stream->time_base, out_stream->time_base, AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX);
        pkt.duration = av_rescale_q(pkt.duration, in_stream->time_base, out_stream->time_base);
        pkt.pos = -1;

        ret = av_interleaved_write_frame(ofmt_ctx, &pkt);
        if (ret < 0) {
            fprintf(stderr, "Error muxing packet\n");
            break;
        }
        av_packet_unref(&pkt);
    }
finish:
    av_write_trailer(ofmt_ctx);

end:
    free(out_filename);
    avformat_close_input(&ifmt_ctx);

    /* close output */
    if (ofmt_ctx && !(ofmt->flags & AVFMT_NOFILE)) {
        avio_closep(&ofmt_ctx->pb);
    }
    avformat_free_context(ofmt_ctx);
    av_freep(&stream_mapping);

    if (ret < 0 && ret != AVERROR_EOF) {
        fprintf(stderr, "Error occurred: %s\n", av_err2str(ret));
        return 1;
    }

    return 0;
}

