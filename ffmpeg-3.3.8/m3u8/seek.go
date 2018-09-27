package remux

//#include <libavutil/timestamp.h>
//#include <libavformat/avformat.h>
//#include "libavutil/avstring.h"
//
//static int seek(const char *input_file, const char *output_prefix, int64_t segment_pts, unsigned int output_index)
//{
//    AVOutputFormat *ofmt = NULL;
//    AVFormatContext *ifmt_ctx = NULL, *ofmt_ctx = NULL;
//    AVPacket pkt;
//    int ret, i;
//    int stream_index = 0;
//    int *stream_mapping = NULL;
//    int stream_mapping_size = 0;
//
//    char *out_filename;
//    double prev_segment_time = 0;
//    double tmp_segment_time;
//    int video_first = 0;
//    int first_segment = 0;
//
//
//    out_filename = malloc(sizeof(char) * (strlen(output_prefix) + 20));
//    if (!out_filename) {
//        fprintf(stderr, "Could not allocate space for output filenames\n");
//        exit(1);
//    }
//
//    av_register_all();
//    av_log_set_level(AV_LOG_ERROR);
//
//    if ((ret = avformat_open_input(&ifmt_ctx, input_file, 0, 0)) < 0) {
//        fprintf(stderr, "Could not open input file '%s'", input_file);
//        ret = AVERROR(ENOMEM);
//        goto end;
//    }
//
//    if ((ret = avformat_find_stream_info(ifmt_ctx, 0)) < 0) {
//        fprintf(stderr, "Failed to retrieve input stream information");
//        ret = AVERROR(ENOMEM);
//        goto end;
//    }
//
//    av_dump_format(ifmt_ctx, 0, input_file, 0);
//
//    if (segment_pts != 0) {
//        ret = av_seek_frame(ifmt_ctx, 0,  segment_pts, AVSEEK_FLAG_ANY);
//        if (ret < 0) {
//            fprintf(stderr, "Error could not seek to position\n");
//            ret = AVERROR(ENOMEM);
//            goto end;
//        }
//    }
//
//    ofmt = av_guess_format("mpegts", NULL, NULL);
//    if (!ofmt) {
//       fprintf(stderr, "Could not find MPEG-TS muxer\n");
//       ret = AVERROR(ENOMEM);
//       goto end;
//    }
//
//    ofmt_ctx = avformat_alloc_context();
//    if (!ofmt_ctx) {
//       fprintf(stderr, "Could not allocated output context");
//       ret = AVERROR(ENOMEM);
//       goto end;
//    }
//    ofmt_ctx->oformat = ofmt;
//
//    stream_mapping_size = ifmt_ctx->nb_streams;
//    stream_mapping = av_mallocz_array(stream_mapping_size, sizeof(*stream_mapping));
//    if (!stream_mapping) {
//        ret = AVERROR(ENOMEM);
//        goto end;
//    }
//
//    for (i = 0; i < ifmt_ctx->nb_streams; i++) {
//        AVStream *out_stream;
//        AVStream *in_stream = ifmt_ctx->streams[i];
//        AVCodecParameters *in_codecpar = in_stream->codecpar;
//
//        if (in_codecpar->codec_type != AVMEDIA_TYPE_AUDIO &&
//            in_codecpar->codec_type != AVMEDIA_TYPE_VIDEO &&
//            in_codecpar->codec_type != AVMEDIA_TYPE_SUBTITLE) {
//            stream_mapping[i] = -1;
//            continue;
//        }
//
//        stream_mapping[i] = stream_index++;
//
//        out_stream = avformat_new_stream(ofmt_ctx, NULL);
//        if (!out_stream) {
//            fprintf(stderr, "Failed allocating output stream\n");
//            ret = AVERROR_UNKNOWN;
//            goto end;
//        }
//
//        ret = avcodec_parameters_copy(out_stream->codecpar, in_codecpar);
//        if (ret < 0) {
//            fprintf(stderr, "Failed to copy codec parameters\n");
//            ret = AVERROR(ENOMEM);
//            goto end;
//        }
//        out_stream->codecpar->codec_tag = 0;
//    }
//
//    snprintf(out_filename, strlen(output_prefix) + 20, "%s-%ld-%u.ts", output_prefix, segment_pts, output_index);
//    if (!(ofmt->flags & AVFMT_NOFILE)) {
//        ret = avio_open(&ofmt_ctx->pb, out_filename, AVIO_FLAG_WRITE);
//        if (ret < 0) {
//            fprintf(stderr, "Could not open output file '%s'", out_filename);
//            ret = AVERROR(ENOMEM);
//            goto end;
//        }
//    }
//
//    ret = avformat_write_header(ofmt_ctx, NULL);
//    if (ret < 0) {
//        fprintf(stderr, "Error occurred when write mpegts header to output file\n");
//        ret = AVERROR(ENOMEM);
//        goto end;
//    }
//
//    while (1) {
//        AVStream *in_stream, *out_stream;
//
//        ret = av_read_frame(ifmt_ctx, &pkt);
//        if (ret < 0)
//            break;
//
//        //fprintf(stderr, "helo stream_index %d\n", pkt.stream_index);
//        in_stream  = ifmt_ctx->streams[pkt.stream_index];
//        if (pkt.stream_index >= stream_mapping_size ||
//            stream_mapping[pkt.stream_index] < 0) {
//            av_packet_unref(&pkt);
//            continue;
//        }
//
//        pkt.stream_index = stream_mapping[pkt.stream_index];
//        out_stream = ofmt_ctx->streams[pkt.stream_index];
//
//        if (first_segment == 0) {
//            first_segment = 1;
//            prev_segment_time = pkt.pts * av_q2d(in_stream->time_base);
//        }
//        tmp_segment_time = pkt.pts * av_q2d(in_stream->time_base);
//        if (in_stream->codecpar->codec_type == AVMEDIA_TYPE_VIDEO && (pkt.flags & AV_PKT_FLAG_KEY)) {
//            if (video_first == 0) {
//                video_first = 1;
//            } else {
//                if (tmp_segment_time - prev_segment_time >= 2) {
//                    //fprintf(stderr, "helo %d,  %f - %f = %f\n", output_index, segment_time, prev_segment_time, (segment_time - prev_segment_time));
//                    av_packet_unref(&pkt);
//                    goto finish;
//                }
//            }
//        }
//        /* copy packet */
//        pkt.pts = av_rescale_q_rnd(pkt.pts, in_stream->time_base, out_stream->time_base, AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX);
//        pkt.dts = av_rescale_q_rnd(pkt.dts, in_stream->time_base, out_stream->time_base, AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX);
//        pkt.duration = av_rescale_q(pkt.duration, in_stream->time_base, out_stream->time_base);
//        pkt.pos = -1;
//
//        ret = av_interleaved_write_frame(ofmt_ctx, &pkt);
//        if (ret < 0) {
//            fprintf(stderr, "Error muxing packet\n");
//            av_packet_unref(&pkt);
//            break;
//        }
//        av_packet_unref(&pkt);
//    }
//
//finish:
//    av_write_trailer(ofmt_ctx);
//
//end:
//    free(out_filename);
//    avformat_close_input(&ifmt_ctx);
//
//    /* close output */
//    if (ofmt_ctx && !(ofmt->flags & AVFMT_NOFILE)) {
//        avio_closep(&ofmt_ctx->pb);
//    }
//    avformat_free_context(ofmt_ctx);
//    av_freep(&stream_mapping);
//
//    if (ret < 0 && ret != AVERROR_EOF) {
//        fprintf(stderr, "Error occurred: %s\n", av_err2str(ret));
//        return 1;
//    }
//
//    return 0;
//}
// #cgo pkg-config: libavformat libavutil
import "C"

import (
	"unsafe"
)

/**
 * inFile 源文件路径如/tmp/big.mp4
 * outFilePre 要输出的ts的前缀, 比如 /tmp/ccc
 * segmentPts 生成ts的开始pts , 比如 0
 * writeIndex 生成哪个ts文件， 比如 1
 */
func Seek(inFile string, outFilePre string, segmentPts int64, writeIndex int) {
	in_filename := C.CString(inFile)
	defer C.free(unsafe.Pointer(in_filename))

	out_prefix := C.CString(outFilePre)
	defer C.free(unsafe.Pointer(out_prefix))

	segment_pts := C.int64_t(segmentPts)
	write_index := C.uint(writeIndex)
	C.seek(in_filename, out_prefix, segment_pts, write_index)
}