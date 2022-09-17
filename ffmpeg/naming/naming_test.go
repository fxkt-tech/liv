package naming_test

import (
	"fmt"
	"testing"

	"github.com/fxkt-tech/liv/ffmpeg/naming"
)

func TestGen(t *testing.T) {
	nm := naming.New()
	name := nm.Gen()
	fmt.Println(name)
}

func TestGen64(t *testing.T) {
	nm := naming.New()
	name := nm.Gen64()
	fmt.Println(name)
}
