ffmpeg编译参数
===
./configure --prefix=/usr/local/ffmpeg-3.1.5 --enable-libmp3lame --enable-gpl --enable-libx264
make
make install


C 语言使用
===
将remuxing.c文件copy到 ffmpeg-3.1.5/doc/examples/下的同名文件

在ffmpeg-3.1.5/doc/examples 下执行make进行编译

go 语言使用
===
go run remuxing.go 即可
