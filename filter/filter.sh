ffmpeg -y -i video.mp4 -filter_complex 
'
[0:v]split[mn][dlsub1]

[dlsub1]trim=1:3[tm]

[tm]delogo=500:300:200:100[dl]

[mn][dl]overlay=eof_action=pass[ovl]
'
 -map '[ovl]' -map '0:a?' -c:v libx264 -c:a aac xu.mp4

ffmpeg -y -i video.mp4 -filter_complex 'delogo=500:300:200:100[ovl]' -map '[ovl]' -map '0:a?' -c:v libx264 -c:a aac xu2.mp4


[0:v]split[split_main][split_sub0][split_sub1]

[split_sub0]trim[split_sub0_trim]

[split_sub0_trim]delogo[split_sub0_trim_delogo0]
[split_sub0_trim_delogo0]delogo[split_sub0_trim_delogo0]

[split_main][split_sub0_trim_delogo0]overlay[main2]



ffmpeg -y 
-i http://b-spiderman.oss-cn-beijing.aliyuncs.com/joker%2Fvideo.mp4?Expires=1624933394&OSSAccessKeyId=LTAIJSnaWScr7IGO&Signature=H%2BQZFsd38jvApOeKepz3Fm9UuHc%3D 
-i http://b-spiderman.oss-cn-beijing.aliyuncs.com/joker%2Flogo_hd.png?Expires=1624933394&OSSAccessKeyId=LTAIJSnaWScr7IGO&Signature=tlvVsFH3SCAXfXQo8y5txUwOMis%3D 
-filter_complex 
[0:v]split[split_main][split_sub0]

[split_sub0]trim=start=1[split_sub0_trim]
[split_sub0_trim]delogo=10:10:200:100[split_sub0_trim_delogo0]
[split_sub0_trim_delogo0]overlay=eof_action=pass[delogo_done_0]

[delogo_done_0]scale=1280:720[scale0]
[scale0][1]overlay=10:y=H-h-10[logo0]
[delogo_done_0]scale=960:540[scale1] 
-map [logo0] -map 0:a? -c:v libx264 -c:a aac -max_muxing_queue_size 4086 -movflags faststart -ss 2 -t 2 /Users/dengxiaochuan/Desktop/7PWwegyKJNonpt8dDhQY6e/boegdyKD1obYwTlAzcqyEn.mp4 
-map [scale1] -map 0:a? -c:v libx264 -c:a aac -max_muxing_queue_size 4086 -movflags faststart /Users/dengxiaochuan/Desktop/7PWwegyKJNonpt8dDhQY6e/nOeor6mavJQeEc63eFZkEa.mp4

[0:v]split[split_main][split_sub0]
[split_sub0]trim=start=1[split_sub0_trim]
[split_sub0_trim]delogo=10:10:200:100[split_sub0_trim_delogo0]
[split_sub0_trim_delogo0]overlay=eof_action=pass[delogo_done_0]
[delogo_done_0]split[delogo_done_0_0][delogo_done_0_1]
[delogo_done_0_0]scale=1280:720[scale0]
[scale0][1]overlay=10:y=H-h-10[logo0]
[delogo_done_0_1]scale=960:540[scale1]