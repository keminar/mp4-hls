生成索引
===
./m3u8 /tmp2/cat.mp4 /tmp2/uuu.m3u8 yyy

生成demo
===
```
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:11
#EXT-X-MEDIA-SEQUENCE:1
EXTINF:9.984580,
yyy-0-1.ts
#EXTINF:7.964444,
yyy-128000-2.ts
#EXTINF:9.589841,
```

生成分片
===
```
./seek /tmp2/cat.mp4 /tmp2/yyy  0 1
./seek /tmp2/cat.mp4 /tmp2/yyy  128000 2
```

