package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/pkg/math"
	"github.com/fxkt-tech/liv/pkg/sugar"
)

func main() {
	var (
		mainfile string
		subfile  string
		outfile  string
	)
	flag.StringVar(&mainfile, "main", "main.mp4", "")
	flag.StringVar(&subfile, "sub", "sub.mp4", "")
	flag.StringVar(&outfile, "output", "output.mp4", "")
	flag.Parse()

	f := NewFakery(outfile)
	err := f.ExtractMainFrames(mainfile)
	if err != nil {
		log.Fatal(err)
	}
	err = f.ExtractSubFrames(subfile)
	if err != nil {
		log.Fatal(err)
	}
	f.Rearrange()
	err = f.Combine()
	if err != nil {
		log.Fatal(err)
	}
}

type Fakery struct {
	mainFile string
	subFile  string
	outfile  string
}

func NewFakery(outfile string) *Fakery {
	f := &Fakery{outfile: outfile}
	f.fix()
	return f
}

func (f *Fakery) fix() {
	os.RemoveAll("main")
	os.Mkdir("main", os.ModePerm)
	os.RemoveAll("sub")
	os.Mkdir("sub", os.ModePerm)
	os.RemoveAll("output")
	os.MkdirAll("output", os.ModePerm)
}

func (f *Fakery) ExtractMainFrames(filename string) error {
	f.mainFile = filename
	var (
		ctx = context.Background()

		ffmpegOpts = []ffmpeg.Option{
			ffmpeg.WithDebug(true),
		}

		// inputs
		iMain = input.WithSimple(filename)

		// filters
		fFPSMain = filter.FPS(math.Fraction(25, 1)).Use(iMain.V())
		// fscale   = filter.Scale(1280, 720).Use(fFPSMain)
		fscale = filter.Scale(1920, 1080).Use(fFPSMain)

		// output
		oOnlyMain = output.New(
			output.Map(fscale),
			output.File("main/%05d.jpg"),
		)
	)

	err := ffmpeg.New(ffmpegOpts...).
		AddInput(iMain).
		AddFilter(fFPSMain, fscale).
		AddOutput(oOnlyMain).
		Run(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (f *Fakery) ExtractSubFrames(filename string) error {
	f.subFile = filename
	var (
		ctx = context.Background()

		ffmpegOpts = []ffmpeg.Option{
			ffmpeg.WithDebug(true),
		}

		// inputs
		iSub = input.WithSimple(filename)

		// filters
		fFPSSub = filter.FPS(math.Fraction(20, 1)).Use(iSub.V())

		// output

		oOnlySub = output.New(
			output.Map(fFPSSub),
			output.File("sub/%05d.jpg"),
		)
	)

	err := ffmpeg.New(ffmpegOpts...).
		AddInput(iSub).
		AddFilter(fFPSSub).
		AddOutput(oOnlySub).
		Run(ctx)
	if err != nil {
		return err
	}

	return nil
}

// 重新排列
func (f *Fakery) Rearrange() {
	mainDir := "main"
	subDir := "sub"
	outputDir := "output"

	// 读取main和sub目录中的文件
	mainFilesCount, err := filesCount(mainDir)
	if err != nil {
		log.Fatalf("读取main目录失败: %v\n", err)
	}

	subFilesCount, err := filesCount(subDir)
	if err != nil {
		log.Fatalf("读取sub目录失败: %v\n", err)
	}

	sequence := []int{2, 1, 2, 1, 2, 1, 2, 1, 1}

	mainIndex, subIndex := 1, 1

	for i := 0; ; {
		if mainIndex >= mainFilesCount {
			break
		}

		for _, num := range sequence {
			if mainIndex >= mainFilesCount {
				break
			}

			var fileName string
			if num == 1 {
				fileName = fmt.Sprintf("%s/%05d.jpg", mainDir, mainIndex)
				mainIndex++
			} else if num == 2 && sugar.In([]int{1, 3, 46}, i+1) {
				var idx int
				switch i + 1 {
				case 1:
					idx = 2
				case 3:
					idx = 4
				case 46:
					idx = 26
				}
				fileName = fmt.Sprintf("%s/%05d.jpg", mainDir, idx)
				subIndex = subIndex%subFilesCount + 1
			} else if num == 2 {
				fileName = fmt.Sprintf("%s/%05d.jpg", subDir, subIndex)
				subIndex = subIndex%subFilesCount + 1
			}

			err := copyFrame(fileName, fmt.Sprintf("%s/%05d.jpg", outputDir, i+1))
			if err != nil {
				log.Fatalf("无法复制帧文件: %v\n", err)
			}

			i++
		}
	}
}

// 合并图片们和音频
func (f *Fakery) Combine() error {
	var (
		ctx = context.Background()

		ffmpegOpts = []ffmpeg.Option{
			ffmpeg.WithDebug(true),
			ffmpeg.WithLogLevel(""),
		}

		// inputs
		iImgs = input.New(
			input.FPS("45"),
			input.I("output/%05d.jpg"),
		)
		iaudioFromVideo = input.WithSimple(f.mainFile)

		// output
		oOnly = output.New(
			output.Map(iImgs.V()),
			output.Map(iaudioFromVideo.MayA()),
			output.VideoCodec(codec.X264),
			output.AudioCodec(codec.Copy),
			output.MovFlags("faststart"),
			output.Shortest(),
			output.File(f.outfile),
		)
	)

	err := ffmpeg.New(ffmpegOpts...).
		AddInput(iImgs, iaudioFromVideo).
		AddOutput(oOnly).
		Run(ctx)
	if err != nil {
		return err
	}

	return nil
}

func filesCount(dir string) (int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	var count int
	for _, file := range files {
		if !file.IsDir() {
			count++
		}
	}

	return count, nil
}

func copyFrame(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
