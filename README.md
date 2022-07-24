# ffmpeg
friendly ffmpeg sdk for go.


- ffmpeg
  - 多路输入
    - -i 输入文件
    - -ss 开始时间
    - -t 持续时间
  - 复合滤波器
    - 多路输入
    - 多路输出
  - 多路输出
    - -map 选择流
    - -metadata kv
    - -c 流编码方式
    - -threads 多线程转码
    - -max_muxing_queue_size 容器封装队列大小（默认4086）
    - -movflags（默认faststart）
    - 输出文件


### 支持的滤镜
- 视频属性video：
  - 缩放clip：可以设置width和height，不设置默认-2（自适应偶数长度）
- 音频属性audio：
  - 码率vb：设置音频的最大码率，如果高于原音频码率则使用原音频码率
- 加水印logo：
  - 可以设置水印图片，水印位置，水印位置偏移量，水印图片缩放大小，均支持绝对像素和百分比
  - 只可以全程加水印，不可以指定时间段
- 去水印delogo：
  - 指定一个开始时间，在这个时间点之后遮一个rect列表
- 裁剪clip：
  - 指定开始时间和持续时长，持续时长不指定的话默认到视频结束
- 添加元信息metadata：
  - 指定一组key-value
  - key仅支持部分字符串，可用的字符串可以看ffmpeg