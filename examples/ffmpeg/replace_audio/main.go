package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/ffmpeg/stream"
)

func main() {
	var (
		ctx    = context.Background()
		input1 = input.WithSimple("v.mp4") // 视频输入
		input2 = input.WithSimple("a.mp3") // 音频输入
		// 创建音频循环滤镜，loop=-1表示无限循环，size=0表示循环整个音频文件
		fAloop = filter.ALoop(-1, 0).Use(stream.A(1))
		// 音频响度标准化滤镜，使不同音乐的音量保持一致
		// i=-16: 目标响度为-16 LUFS (适合一般视频)
		// lra=11: 响度范围目标值
		// tp=-1.5: 真峰值限制
		fLoudnorm = filter.Loudnorm(-16, 11, -1.5).Use(fAloop)
	)

	err := ffmpeg.New(
		ffmpeg.WithLogLevel(""),
		ffmpeg.WithDebug(true),
		// ffmpeg.WithDry(true),
	).AddInput(
		input1, input2,
	).AddFilter(
		fAloop, fLoudnorm,
	).AddOutput(
		output.New(
			// 从第一个输入(视频)中取视频流
			output.Map(stream.Select(0, stream.Video)),
			// 使用循环并标准化后的音频流
			output.Map(fLoudnorm),
			// 使用最短的流（即视频）作为输出时长，这样会在视频结束时停止
			output.Shortest(),
			// 视频流直接复制,不重新编码
			output.VideoCodec(codec.Copy),
			// 音频流使用AAC编码
			output.AudioCodec(codec.AAC),
			output.File("out2.mp4"),
		),
	).Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
