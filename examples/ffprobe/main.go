package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffprobe"
)

func main() {
	ctx := context.Background()

	fp, err := ffprobe.New().
		Input("").
		Extract(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	vStream := fp.GetFirstVideoStream()
	if vStream == nil {
		fmt.Println("file has no video stream")
		return
	}

	fmt.Println(vStream)
}
