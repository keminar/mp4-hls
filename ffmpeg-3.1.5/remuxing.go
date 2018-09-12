package main


//#include <libavutil/timestamp.h>
//#include <libavformat/avformat.h>
//#include <libavutil/dict.h>
//
//static void log_packet(const AVFormatContext *fmt_ctx, const AVPacket *pkt, const char *tag)
//{
//    AVRational *time_base = &fmt_ctx->streams[pkt->stream_index]->time_base;
//
//    printf("%s: pts:%s pts_time:%s dts:%s dts_time:%s duration:%s duration_time:%s stream_index:%d\n",
//           tag,
//           av_ts2str(pkt->pts), av_ts2timestr(pkt->pts, time_base),
//           av_ts2str(pkt->dts), av_ts2timestr(pkt->dts, time_base),
//           av_ts2str(pkt->duration), av_ts2timestr(pkt->duration, time_base),
//           pkt->stream_index);
//}
//
//static int remux(const char *in_filename,const char *out_filename)
//{
//    AVOutputFormat *ofmt = NULL;
//    AVFormatContext *ifmt_ctx = NULL, *ofmt_ctx = NULL;
//    AVPacket pkt;
//    int ret, i;
//
//    av_register_all();
//
//    if ((ret = avformat_open_input(&ifmt_ctx, in_filename, 0, 0)) < 0) {
//        fprintf(stderr, "Could not open input file '%s'", in_filename);
//        goto end;
//    }
//
//    if ((ret = avformat_find_stream_info(ifmt_ctx, 0)) < 0) {
//        fprintf(stderr, "Failed to retrieve input stream information");
//        goto end;
//    }
//
//    av_dump_format(ifmt_ctx, 0, in_filename, 0);
//
//    avformat_alloc_output_context2(&ofmt_ctx, NULL, NULL, out_filename);
//    if (!ofmt_ctx) {
//        fprintf(stderr, "Could not create output context\n");
//        ret = AVERROR_UNKNOWN;
//        goto end;
//    }
//
//    ofmt = ofmt_ctx->oformat;
//
//    for (i = 0; i < ifmt_ctx->nb_streams; i++) {
//        AVStream *in_stream = ifmt_ctx->streams[i];
//        AVStream *out_stream = avformat_new_stream(ofmt_ctx, in_stream->codec->codec);
//        if (!out_stream) {
//            fprintf(stderr, "Failed allocating output stream\n");
//            ret = AVERROR_UNKNOWN;
//            goto end;
//        }
//
//        ret = avcodec_copy_context(out_stream->codec, in_stream->codec);
//        if (ret < 0) {
//            fprintf(stderr, "Failed to copy context from input to output stream codec context\n");
//            goto end;
//        }
//        out_stream->codec->codec_tag = 0;
//        if (ofmt_ctx->oformat->flags & AVFMT_GLOBALHEADER)
//            out_stream->codec->flags |= AV_CODEC_FLAG_GLOBAL_HEADER;
//    }
//    av_dump_format(ofmt_ctx, 0, out_filename, 1);
//
//    if (!(ofmt->flags & AVFMT_NOFILE)) {
//        ret = avio_open(&ofmt_ctx->pb, out_filename, AVIO_FLAG_WRITE);
//        if (ret < 0) {
//            fprintf(stderr, "Could not open output file '%s'", out_filename);
//            goto end;
//        }
//    }
//
//    AVDictionary *opt = NULL;
//    av_dict_set_int(&opt, "hls_list_size", 0, 0);
//
//    ret = avformat_write_header(ofmt_ctx, &opt);
//    if (ret < 0) {
//        fprintf(stderr, "Error occurred when opening output file\n");
//        goto end;
//    }
//
//    AVBitStreamFilterContext *vbsf = av_bitstream_filter_init("h264_mp4toannexb");
//    while (1) {
//        AVStream *in_stream, *out_stream;
//
//        ret = av_read_frame(ifmt_ctx, &pkt);
//        if (ret < 0)
//            break;
//
//        in_stream  = ifmt_ctx->streams[pkt.stream_index];
//        out_stream = ofmt_ctx->streams[pkt.stream_index];
//
////       log_packet(ifmt_ctx, &pkt, "in");
//	if (pkt.stream_index == 0) {
//		AVPacket fpkt = pkt;
//		av_bitstream_filter_filter(vbsf, out_stream->codec, NULL, &fpkt.data, &fpkt.size, pkt.data, pkt.size, pkt.flags & AV_PKT_FLAG_KEY);
//		pkt.data = fpkt.data;
//		pkt.size = fpkt.size;
//	}
//        /* copy packet */
//        pkt.pts = av_rescale_q_rnd(pkt.pts, in_stream->time_base, out_stream->time_base, AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX);
//        pkt.dts = av_rescale_q_rnd(pkt.dts, in_stream->time_base, out_stream->time_base, AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX);
//        pkt.duration = av_rescale_q(pkt.duration, in_stream->time_base, out_stream->time_base);
//        pkt.pos = -1;
//  //      log_packet(ofmt_ctx, &pkt, "out");
//
//
//        ret = av_interleaved_write_frame(ofmt_ctx, &pkt);
//        if (ret < 0) {
//            fprintf(stderr, "Error muxing packet\n");
//            break;
//        }
//        av_packet_unref(&pkt);
//    }
//
//    av_write_trailer(ofmt_ctx);
//end:
//
//    avformat_close_input(&ifmt_ctx);
//
//    /* close output */
//    if (ofmt_ctx && !(ofmt->flags & AVFMT_NOFILE))
//        avio_closep(&ofmt_ctx->pb);
//    avformat_free_context(ofmt_ctx);
//    av_bitstream_filter_close(vbsf);
//    av_dict_free(&opt);
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

func main() {
	inFile := C.CString("/tmp2/big.mp4")
	defer C.free(unsafe.Pointer(inFile))

	outFile := C.CString("/tmp2/bigAA.m3u8")
	defer C.free(unsafe.Pointer(outFile))
	C.remux(inFile, outFile)
}
