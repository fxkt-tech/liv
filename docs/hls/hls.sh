rm -rf master.m3u8
rm -rf 0/
rm -rf 1/

ffmpeg \
-i input.mp4 \
-filter_complex '[0:v]scale=-2:720[stream0];[0:v]scale=-2:360[stream1]' \
-map '[stream0]' -map 0:a \
-map '[stream1]' -map 0:a \
-f hls -var_stream_map "v:0,a:0 v:1,a:1" \
-hls_segment_type mpegts \
-hls_flags independent_segments \
-hls_playlist_type vod \
-hls_start_number_source generic \
-hls_time 10 -hls_list_size 0 \
-master_pl_name "master.m3u8" \
-hls_segment_filename "%v/%5d.ts" "%v/rendition.m3u8"