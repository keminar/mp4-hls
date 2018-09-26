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
    double segment_max_duration;
    char *m3u8_file;
    char *tmp_m3u8_file;
    const char *url_prefix;
    unsigned int write_m3u8;
    unsigned int write_index;
    unsigned int sequence;
};
int write_index_header(FILE *index_fp, const struct options_t options);
int write_index_segment(FILE *index_fp, const struct options_t options, const char* basename, unsigned int output_index, double duration);
int write_index_trailer(FILE *index_fp);
/*
static void log_packet(const AVFormatContext *fmt_ctx, const AVPacket *pkt, const char *tag)
{
    AVRational *time_base = &fmt_ctx->streams[pkt->stream_index]->time_base;

    printf("%s: pts:%s pts_time:%s dts:%s dts_time:%s duration:%s duration_time:%s stream_index:%d pos:%d, size:%d\n",
           tag,
           av_ts2str(pkt->pts), av_ts2timestr(pkt->pts, time_base),
           av_ts2str(pkt->dts), av_ts2timestr(pkt->dts, time_base),
           av_ts2str(pkt->duration), av_ts2timestr(pkt->duration, time_base),
           pkt->stream_index, pkt->pos, pkt->size);
}*/

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
    const char  *basename;
    double prev_segment_time = 0;
    double segment_time;
    double tmp_segment_time;
    double duration = 0;
    unsigned int output_index = 1; //分片名从1开始
    struct options_t options;
    options.segment_max_duration = 0; //每个分片TS的最大的时长
    options.url_prefix = "";
    int write_ret = 1;
    int video_first = 0;
    int has_write_ts = 0;
    FILE *index_fp;


    if (argc < 5) {
        printf("usage: %s input output\n"
               "API example program to remux a media file with libavformat and libavcodec.\n"
               "The output format is guessed according to the file extension.\n"
               "\n", argv[0]);
        return 1;
    }

    options.sequence = output_index;
    options.input_file  = argv[1];
    options.output_prefix = argv[2];
    options.write_m3u8 = (unsigned int)strtol(argv[3], NULL, 10);
    options.write_index = (unsigned int)strtol(argv[4], NULL, 10);
    options.m3u8_file = malloc(sizeof(char) * (strlen(options.output_prefix) + strlen(".m3u8") + 1));
    snprintf(options.m3u8_file, strlen(options.output_prefix) + 15, "%s.m3u8", options.output_prefix);

    options.tmp_m3u8_file = malloc(sizeof (char) * (strlen(options.output_prefix) + strlen(".m3u8.tmp") + 1));
    snprintf(options.tmp_m3u8_file, strlen(options.output_prefix) + 15, "%s.m3u8.tmp", options.output_prefix);

    out_filename = malloc(sizeof(char) * (strlen(options.output_prefix) + 15));
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

/*    avformat_alloc_output_context2(&ofmt_ctx, NULL, NULL, out_filename);
    if (!ofmt_ctx) {
        fprintf(stderr, "Could not create output context\n");
        ret = AVERROR_UNKNOWN;
        goto end;
    }
    */

    ofmt = av_guess_format("mpegts", NULL, NULL);
    if (!ofmt) {
       fprintf(stderr, "Could not find MPEG-TS muxer\n");
       exit(1);
    }

    ofmt_ctx = avformat_alloc_context();
    if (!ofmt_ctx) {
       fprintf(stderr, "Could not allocated output context");
       exit(1);
    }
    ofmt_ctx->oformat = ofmt;

    stream_mapping_size = ifmt_ctx->nb_streams;
    stream_mapping = av_mallocz_array(stream_mapping_size, sizeof(*stream_mapping));
    if (!stream_mapping) {
        ret = AVERROR(ENOMEM);
        goto end;
    }

    ofmt = ofmt_ctx->oformat;

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
    av_dump_format(ofmt_ctx, 0, options.output_prefix, 1);

    //fprintf(stderr, "%d", options.write_index);
    if (options.write_index == 0 || options.write_index == output_index) {

        snprintf(out_filename, strlen(options.output_prefix) + 15, "%s-%u.ts", options.output_prefix, output_index);

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
        has_write_ts = 1;
    }

    if (options.write_m3u8 > 0) {
        index_fp = fopen(options.tmp_m3u8_file, "w");
        if (!index_fp) {
            fprintf(stderr, "Could not open temporary m3u8 index file (%s), no index file will be created\n", options.tmp_m3u8_file);
            goto end;
        }
        write_ret = write_index_header(index_fp, options);
        if (write_ret < 0) {
            goto end;
        }
    }

    basename = av_basename(options.output_prefix);

    while (1) {
        AVStream *in_stream, *out_stream;

        ret = av_read_frame(ifmt_ctx, &pkt);
        if (ret < 0)
            break;

        in_stream  = ifmt_ctx->streams[pkt.stream_index];
        if (pkt.stream_index >= stream_mapping_size ||
            stream_mapping[pkt.stream_index] < 0) {
            av_packet_unref(&pkt);
            continue;
        }

        pkt.stream_index = stream_mapping[pkt.stream_index];
        out_stream = ofmt_ctx->streams[pkt.stream_index];
        //log_packet(ifmt_ctx, &pkt, "in");

        tmp_segment_time = pkt.pts * av_q2d(in_stream->time_base);
        if (in_stream->codecpar->codec_type == AVMEDIA_TYPE_VIDEO && (pkt.flags & AV_PKT_FLAG_KEY)) {
            if (video_first == 0) { //第一个片
                video_first = 1;
            } else {
                if (tmp_segment_time - prev_segment_time >= 2) { // 几秒一个分隔
                    duration = segment_time - prev_segment_time; // 保证下一个片从关键帧开始
                    //fprintf(stderr, "helo -%f, %f,%f = %f\n", tmp_segment_time, segment_time, prev_segment_time, duration);
                    if (duration > options.segment_max_duration) {
                        options.segment_max_duration = duration;
                    }

                    if (options.write_index == 0) { // 全部分片
                        av_write_trailer(ofmt_ctx);
                        avio_flush(ofmt_ctx->pb);
                        avio_closep(&ofmt_ctx->pb);
                    }
                    if (options.write_m3u8 > 0) {
                        write_ret = write_index_segment(index_fp, options, basename, output_index, duration);
                        if (write_ret < 0) {
                            goto end;
                        }
                    }

                    prev_segment_time = segment_time;
                    output_index++;

                    if (options.write_index == 0 || options.write_index == output_index) {
                        // 下一片
                        snprintf(out_filename, strlen(options.output_prefix) + 15, "%s-%u.ts", options.output_prefix, output_index);
                        ret = avio_open(&ofmt_ctx->pb, out_filename, AVIO_FLAG_WRITE);
                        if (ret < 0) {
                            fprintf(stderr, "Could not open output file '%s'", out_filename);
                            goto end;
                        }
                        ret = avformat_write_header(ofmt_ctx, NULL);
                        if (ret < 0) {
                            fprintf(stderr, "Error occurred when opening output file\n");
                            goto end;
                        }
                        has_write_ts = 1;
                    }

                }
            }
        }
        segment_time = tmp_segment_time;


        if (options.write_index == 0 ||options.write_index == output_index) {
            /* copy packet */
            pkt.pts = av_rescale_q_rnd(pkt.pts, in_stream->time_base, out_stream->time_base, AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX);
            pkt.dts = av_rescale_q_rnd(pkt.dts, in_stream->time_base, out_stream->time_base, AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX);
            pkt.duration = av_rescale_q(pkt.duration, in_stream->time_base, out_stream->time_base);
            pkt.pos = -1;
            //log_packet(ofmt_ctx, &pkt, "out");

            ret = av_interleaved_write_frame(ofmt_ctx, &pkt);
            if (ret < 0) {
                fprintf(stderr, "Error muxing packet\n");
                break;
            }
        }
        av_packet_unref(&pkt);
    }

    if (has_write_ts == 1) {
        av_write_trailer(ofmt_ctx);
    }
    if (options.write_m3u8 > 0) {

        duration = segment_time - prev_segment_time;
        if (duration > options.segment_max_duration) {
            options.segment_max_duration = duration;
        }
        //fprintf(stderr, "helo - %f,%f = %f\n", segment_time, prev_segment_time, duration);
        write_ret = write_index_segment(index_fp, options, basename, output_index, duration);
        if (write_ret < 0) {
            goto end;
        }

        write_ret = write_index_trailer(index_fp);
        if (write_ret == 0) {
            fseek(index_fp, 0, SEEK_SET);
            write_index_header(index_fp, options); //修改TARGETDURATION值
            rename(options.tmp_m3u8_file, options.m3u8_file);
        }
    }

end:

    free(options.m3u8_file);
    free(options.tmp_m3u8_file);
    if (options.write_m3u8 > 0) {
        fclose(index_fp);
    }
    free(out_filename);
    avformat_close_input(&ifmt_ctx);

    /* close output */
    if (has_write_ts == 1) {
        if (ofmt_ctx && !(ofmt->flags & AVFMT_NOFILE)) {
            avio_closep(&ofmt_ctx->pb);
        }
    }
    avformat_free_context(ofmt_ctx);
    av_freep(&stream_mapping);

    if (ret < 0 && ret != AVERROR_EOF) {
        fprintf(stderr, "Error occurred: %s\n", av_err2str(ret));
        return 1;
    }

    return 0;
}

int write_index_header(FILE *index_fp, const struct options_t options) {
    char *write_buf;
    write_buf = malloc(sizeof(char) * 1024);
    if (!write_buf) {
        fprintf(stderr, "Could not allocate write buffer for index file, index file will be invalid\n");
        return -1;
    }

    snprintf(write_buf, 1024, "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:%-5lu\n#EXT-X-MEDIA-SEQUENCE:%d\n", (long)options.segment_max_duration, options.sequence);
    if (fwrite(write_buf, strlen(write_buf), 1, index_fp) != 1) {
        fprintf(stderr, "Could not write to m3u8 index file, will not continue writing to index file\n");
        free(write_buf);
        return -1;
    }
    free(write_buf);
    return 0;
}

int write_index_segment(FILE *index_fp, const struct options_t options, const char *basename, unsigned int output_index, double duration) {
    char *write_buf;
    write_buf = malloc(sizeof(char) * 1024);
    if (!write_buf) {
        fprintf(stderr, "Could not allocate write buffer for index file, index file will be invalid\n");
        return -1;
    }
    snprintf(write_buf, 1024, "#EXTINF:%f,\n%s%s-%u.ts\n", duration, options.url_prefix, basename, output_index);
    if (fwrite(write_buf, strlen(write_buf), 1, index_fp) != 1) {
        fprintf(stderr, "Could not write to m3u8 index file, will not continue writing to index file\n");
        free(write_buf);
        return -1;
    }
    free(write_buf);
    return 0;
}

int write_index_trailer(FILE *index_fp) {
    char *write_buf;
    write_buf = malloc(sizeof(char) * 1024);
    if (!write_buf) {
        fprintf(stderr, "Could not allocate write buffer for index file, index file will be invalid\n");
        return -1;
    }
    snprintf(write_buf, 1024, "#EXT-X-ENDLIST\n");
    if (fwrite(write_buf, strlen(write_buf), 1, index_fp) != 1) {
        fprintf(stderr, "Could not write last file and endlist tag to m3u8 index file\n");
        free(write_buf);
        return -1;
    }

    free(write_buf);
    return 0;
}


