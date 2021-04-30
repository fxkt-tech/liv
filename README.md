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