package remux


//#include <libavutil/timestamp.h>
//#include <libavformat/avformat.h>
//#include "libavutil/avstring.h"
//
//struct options_m3u8 {
//    const char *input_file;
//    const char *ts_path_prefix;
//    double segment_max_duration;
//    const char *m3u8_file;
//    char *tmp_m3u8_file;
//    const char *ts_url_prefix;
//    unsigned int sequence;
//};
//int write_index_header(FILE *index_fp, char *write_buf, const struct options_m3u8 options);
//int write_index_segment(FILE *index_fp, char *write_buf, const struct options_m3u8 options, unsigned int output_index, double duration, int64_t segment_pts);
//int write_index_trailer(FILE *index_fp, char *write_buf);
//
//static int m3u8(const char *in_filename,const char *m3u8_file, const char *ts_path_prefix, const char *ts_url_prefix)
//{
//    AVFormatContext *ifmt_ctx = NULL;
//    AVPacket pkt;
//    int ret, i;
//    int stream_index = 0;
//    int *stream_mapping = NULL;
//    int stream_mapping_size = 0;
//
//    double prev_segment_time = 0;
//    double segment_time;
//    double tmp_segment_time;
//    double duration = 0;
//    unsigned int output_index = 1;
//    struct options_m3u8 options;
//    options.segment_max_duration = 0;
//
//    int write_ret = 1;
//    int video_first = 0;
//    FILE *index_fp;
//    int64_t segment_pts = 0;
//
//
//    memset(&options, 0, sizeof(options));
//    options.sequence = output_index;
//    options.input_file  = in_filename;
//    options.m3u8_file = m3u8_file;
//    options.ts_path_prefix = ts_path_prefix;
//    options.ts_url_prefix = ts_url_prefix;
//    options.tmp_m3u8_file = malloc(sizeof (char) * (strlen(options.m3u8_file) + strlen(".tmp") + 1));
//    snprintf(options.tmp_m3u8_file, strlen(options.m3u8_file) + 5, "%s.tmp", options.m3u8_file);
//
//
//    av_register_all();
//    av_log_set_level(AV_LOG_ERROR);
//
//    if ((ret = avformat_open_input(&ifmt_ctx, options.input_file, 0, 0)) < 0) {
//        fprintf(stderr, "Could not open input file '%s'", options.input_file);
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
//    av_dump_format(ifmt_ctx, 0, options.input_file, 0);
//
//
//    stream_mapping_size = ifmt_ctx->nb_streams;
//    stream_mapping = av_mallocz_array(stream_mapping_size, sizeof(*stream_mapping));
//    if (!stream_mapping) {
//        ret = AVERROR(ENOMEM);
//        goto end;
//    }
//    for (i = 0; i < ifmt_ctx->nb_streams; i++) {
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
//    }
//
//    char *write_buf;
//    write_buf = malloc(sizeof(char) * 1024);
//    if (!write_buf) {
//        fprintf(stderr, "Could not allocate write buffer for index file, index file will be invalid\n");
//        ret = AVERROR(ENOMEM);
//        goto free_buffer;
//    }
//
//    index_fp = fopen(options.tmp_m3u8_file, "w");
//    if (!index_fp) {
//        fprintf(stderr, "Could not open temporary m3u8 index file (%s), no index file will be created\n", options.tmp_m3u8_file);
//        ret = AVERROR(ENOMEM);
//        goto close_indexfp;
//    }
//    write_ret = write_index_header(index_fp, write_buf, options);
//    if (write_ret < 0) {
//        ret = AVERROR(ENOMEM);
//        goto close_indexfp;
//    }
//    while (1) {
//        AVStream *in_stream;
//
//        ret = av_read_frame(ifmt_ctx, &pkt);
//        if (ret < 0)
//            break;
//        //fprintf(stderr, "hi %ld\n", pkt.pts);
//        in_stream  = ifmt_ctx->streams[pkt.stream_index];
//        if (pkt.stream_index >= stream_mapping_size ||
//            stream_mapping[pkt.stream_index] < 0) {
//            av_packet_unref(&pkt);
//            continue;
//        }
//        if (in_stream->codecpar->codec_type != AVMEDIA_TYPE_VIDEO ) {
//            av_packet_unref(&pkt);
//            continue;
//        }
//        pkt.stream_index = stream_mapping[pkt.stream_index];
//        tmp_segment_time = pkt.pts * av_q2d(in_stream->time_base);
//        //printf("%d: pts %d video %d\n", pkt.pts, in_stream->codecpar->codec_type == AVMEDIA_TYPE_VIDEO, pkt.flags & AV_PKT_FLAG_KEY);
//        if (in_stream->codecpar->codec_type == AVMEDIA_TYPE_VIDEO && (pkt.flags & AV_PKT_FLAG_KEY)) {
//            //printf("%d , %d \n", pkt.pts,  in_stream->time_base.den);
//            if (video_first == 0) {
//                video_first = 1;
//            } else {
//                if (tmp_segment_time - prev_segment_time >= 1) {
//                    duration = segment_time - prev_segment_time;
//                    if (duration > options.segment_max_duration) {
//                        options.segment_max_duration = duration;
//                    }
//
//                    write_ret = write_index_segment(index_fp, write_buf, options, output_index, duration, segment_pts);
//                    if (write_ret < 0) {
//                        ret = AVERROR(ENOMEM);
//                        goto close_indexfp;
//                    }
//
//                    prev_segment_time = segment_time;
//                    segment_pts = pkt.pts;
//                    output_index++;
//                }
//            }
//        }
//        segment_time = tmp_segment_time;
//        av_packet_unref(&pkt);
//    }
//
//    duration = segment_time - prev_segment_time;
//    if (duration > options.segment_max_duration) {
//        options.segment_max_duration = duration;
//    }
//    //fprintf(stderr, "helo %d,  %f - %f = %f\n", output_index, segment_time, prev_segment_time, duration);
//    write_ret = write_index_segment(index_fp, write_buf, options, output_index, duration, segment_pts);
//    if (write_ret < 0) {
//        ret = AVERROR(ENOMEM);
//        goto close_indexfp;
//    }
//
//    write_ret = write_index_trailer(index_fp, write_buf);
//    if (write_ret == 0) {
//        fseek(index_fp, 0, SEEK_SET);
//        write_index_header(index_fp, write_buf, options); //修改TARGETDURATION值
//        rename(options.tmp_m3u8_file, options.m3u8_file);
//    }
//
//close_indexfp:
//    fclose(index_fp);
//free_buffer:
//    free(write_buf);
//    write_buf = NULL;
//end:
//    free(options.tmp_m3u8_file);
//    options.tmp_m3u8_file = NULL;
//    avformat_close_input(&ifmt_ctx);
//
//    av_freep(&stream_mapping);
//
//    if (ret < 0 && ret != AVERROR_EOF) {
//        fprintf(stderr, "Error occurred: %s\n", av_err2str(ret));
//        return 1;
//    }
//
//    return 0;
//}
//
//int write_index_header(FILE *index_fp, char *write_buf, const struct options_m3u8 options) {
//    snprintf(write_buf, 1024, "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:%-5lu\n#EXT-X-MEDIA-SEQUENCE:%d\n", (long)options.segment_max_duration, options.sequence);
//    if (fwrite(write_buf, strlen(write_buf), 1, index_fp) != 1) {
//        fprintf(stderr, "Could not write to m3u8 index file, will not continue writing to index file\n");
//        return -1;
//    }
//    return 0;
//}
//
//int write_index_segment(FILE *index_fp, char *write_buf, const struct options_m3u8 options, unsigned int output_index, double duration, int64_t segment_pts) {
//    snprintf(write_buf, 1024, "#EXTINF:%f,\n%s%s-%ld-%u.ts\n", duration, options.ts_url_prefix, options.ts_path_prefix, segment_pts, output_index);
//    if (fwrite(write_buf, strlen(write_buf), 1, index_fp) != 1) {
//        fprintf(stderr, "Could not write to m3u8 index file, will not continue writing to index file\n");
//        return -1;
//    }
//    return 0;
//}
//
//int write_index_trailer(FILE *index_fp, char *write_buf) {
//    snprintf(write_buf, 1024, "#EXT-X-ENDLIST\n");
//    if (fwrite(write_buf, strlen(write_buf), 1, index_fp) != 1) {
//        fprintf(stderr, "Could not write last file and endlist tag to m3u8 index file\n");
//        return -1;
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
 * outFilePre 要输出的m3u8，比如  /tmp/ccc.m3u8
 * tsNamePre 生成的ts的文件名，比如 ccc
 * tsUrl 生成哪些ts文件url前缀,比如 /video/
 */
func M3u8(inFile string, m3u8File string, tsNamePre string, tsUrl string) {
	in_filename := C.CString(inFile)
	defer C.free(unsafe.Pointer(in_filename))

	m3u8_file := C.CString(m3u8File)
	defer C.free(unsafe.Pointer(m3u8_file))

	ts_path_prefix := C.CString(tsNamePre)
	defer C.free(unsafe.Pointer(ts_path_prefix))

	ts_url_prefix := C.CString(tsUrl)
	defer C.free(unsafe.Pointer(ts_url_prefix))
	C.m3u8(in_filename, m3u8_file, ts_path_prefix, ts_url_prefix)
}
