package fftool

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

type CoreFunc func(x, y int, w, h int) (uint8, uint8, uint8, uint8)

func GenMask(coref CoreFunc, width int, height int, outfile string) error {
	canvas := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := coref(x, y, width, height)
			canvas.Set(x, y, color.NRGBA{r, g, b, a})
		}
	}
	imgf, err := os.Create(outfile)
	if err != nil {
		return err
	}
	return png.Encode(imgf, canvas)
}

// 核函数：alpha线性渐变
// offset：根据高度向下偏移
// step：实际作用高度
// maxAlpha：渐变时alpha最大值
// return：rgba
func CoreAlphaLinear(_, y int, _, h int) (uint8, uint8, uint8, uint8) {
	offset := h / 3
	if y < offset {
		return 0, 0, 0, 0
	}
	step := h - offset
	maxAlpha := 80
	return 0, 0, 0, uint8((y - offset) * maxAlpha / step)
}
