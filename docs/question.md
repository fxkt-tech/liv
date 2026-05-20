1. quality参数怎么设置？
2. -threads 4 的含义是4核配一路吗？
3. 转码需要支持的功能有：缩放/加logo/遮标/裁剪/多分辨率/写入metadata，以下是转码命令：
```bash
# 普通转码
ffmpeg -y -i in.mp4 -metadata comment=fu789sg -filter_complex 'scale=-1:720' -c:v libwz264 -c:a copy -threads 4 -max_muxing_queue_size 4086 -movflags faststart out.mp4

# 加logo（右上角10px）
ffmpeg -y -i in.mp4 -i logo.png -metadata comment=fu789sg -filter_complex '[0:v]scale=-1:720[p1];[p1][1]overlay=W-w-10:10' -c:v libwz264 -c:a copy -threads 4 -max_muxing_queue_size 4086 -movflags faststart out_logo.mp4

# 遮标
ffmpeg -y -i in.mp4 -metadata comment=fu789sg -filter_complex '[0]delogo=1490:40:400:100[p1];[p1]scale=-1:720' -c:v libwz264 -c:a copy -threads 4 -max_muxing_queue_size 4086 -movflags faststart out_delogo.mp4

# 遮标后加logo
ffmpeg -y -i in.mp4 -i logo.png -metadata comment=fu789sg -filter_complex '[0]delogo=1490:40:400:100[p1];[p1]scale=-1:720[p2];[p2][1]overlay=W-w-10:10' -c:v libwz264 -c:a copy -threads 4 -max_muxing_queue_size 4086 -movflags faststart out_mixin.mp4

# 多路输出 这种方式可以么？但发现转码后sd的码率要比hd的要大
ffmpeg -y -t 5 -i in.mp4 -i logo.png -metadata comment=fu789sg -filter_complex '[0]delogo=1490:40:400:100[p1];[p1]split=2[p1_hd][p1_sd];[p1_hd]scale=-1:720[p2_hd];[p2_hd][1]overlay=W-w-10:10[p_hd];[p1_sd]scale=-1:540[p2_sd];[p2_sd][1]overlay=W-w-10:10[p_sd]' -c:v libwz264 -c:a copy -threads 4 -max_muxing_queue_size 4086 -movflags faststart -map '[p_hd]' -map 0:a out_hd.mp4 -map '[p_sd]' -map 0:a out_sd.mp4
# 修改后的
ffmpeg -y -t 5 -i in.mp4 -i logo.png  -filter_complex '[0]delogo=1490:40:400:100[p1];[p1]split=2[p1_hd][p1_sd];[p1_hd]scale=-1:720[p2_hd];[p2_hd][1]overlay=W-w-10:10[p_hd];[p1_sd]scale=-1:540[p2_sd];[p2_sd][1]overlay=W-w-10:10[p_sd]' -map '[p_hd]' -map 0:a -metadata comment=fu789sg -c:v libwz264 -c:a copy -threads 4 -max_muxing_queue_size 4086 -movflags faststart out_hd.mp4 -map '[p_sd]' -map 0:a -metadata comment=fu789sg -c:v libwz264 -c:a copy -threads 4 -max_muxing_queue_size 4086 -movflags faststart out_sd.mp4

# 以上这些命令可以么？有没有更好的建议？
```

```bash
ffmpeg
-y
-t 5
-i in.mp4 -i logo.png
-filter_complex '[0]delogo=1490:40:400:100[p1];[p1]split=2[p1_hd][p1_sd];[p1_hd]scale=trunc(oh*a/2)*2:720[p2_hd];[p2_hd][1]overlay=W-w-10:10[p_hd];[p1_sd]scale=-1:540[p2_sd];[p2_sd][1]overlay=W-w-10:10[p_sd]'

-map '[p_hd]'
-map 0:a
-metadata comment=fu789sg
-c:v libwz264
-c:a copy
-threads 4
-max_muxing_queue_size 4086
-movflags faststart
out_hd.mp4

-map '[p_sd]' -map 0:a -metadata comment=fu789sg -c:v libwz264 -c:a copy -threads 4 -max_muxing_queue_size 4086 -movflags faststart out_sd.mp4

trunc(oh*a/2)*2:720
720:trunc(ow/a/2)*2
```

