package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fxkt-tech/liv/ffcut"
	ffcutrenderer "github.com/fxkt-tech/liv/ffcut/renderer"
)

func main() {
	projectJSON, err := os.ReadFile("project.ffcut.json")
	if err != nil {
		log.Fatal(err)
	}
	project, err := ffcut.Unmarshal(projectJSON)
	if err != nil {
		log.Fatal(err)
	}
	output, err := filepath.Abs("final.mp4")
	if err != nil {
		log.Fatal(err)
	}
	if err := ffcutrenderer.Render(
		context.Background(),
		project,
		output,
		ffcutrenderer.WithDebug(os.Stdout),
	); err != nil {
		log.Fatal(err)
	}
	fmt.Println(output)
}
