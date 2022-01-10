#!/bin/sh
BASE_URL=https://yilan-open.oss-cn-beijing.aliyuncs.com
openssl rand 16 > file.key
echo $BASE_URL/file.key > file.keyinfo
echo file.key >> file.keyinfo
echo $(openssl rand -hex 16) >> file.keyinfo
ffmpeg -i xx.mp4 -c:v h264 -hls_flags delete_segments \
  -hls_key_info_file file.keyinfo out.m3u8
