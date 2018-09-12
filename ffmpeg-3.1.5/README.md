ffmpeg编译参数
===
```
./configure --prefix=/usr/local/ffmpeg-3.1.5 --enable-libmp3lame --enable-gpl --enable-libx264
make
make install
```

.bashrc 环境变量
===
```
export PATH=$PATH:/usr/local/go/bin
export GOROOT=/usr/local/go
export GOPATH=/root/golang

export FFMPEG_ROOT=/usr/local/ffmpeg-3.1.5
export CGO_LDFLAGS="-L$FFMPEG_ROOT/lib/ -lavcodec -lavformat -lavutil -lswscale -lswresample -lavdevice -lavfilter -lm"
export CGO_CFLAGS="-I$FFMPEG_ROOT/include"
export PKG_CONFIG_PATH=$FFMPEG_ROOT/lib/pkgconfig/
export LD_LIBRARY_PATH=$FFMPEG_ROOT/lib:/usr/local/lib:/usr/lib
```

C 语言使用
===
将remuxing.c文件copy到 ffmpeg-3.1.5/doc/examples/下的同名文件

在ffmpeg-3.1.5/doc/examples 下执行make进行编译

go 语言使用
===
go run remuxing.go 即可


