package main

import (
	"bytes"
	"embed"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"image"
	_ "image/png"
)

//go:embed stuff/*png
var stuff embed.FS

var images map[string]*ebiten.Image

func loadImages() error {
	images = make(map[string]*ebiten.Image)
	files, err := stuff.ReadDir("stuff")
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() {
			ext := filepath.Ext(file.Name())
			if ext != ".png" {
				continue
			}

			data, err := stuff.ReadFile(filepath.Join("stuff", file.Name()))
			if err != nil {
				return err
			}
			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				return err
			}

			shortname := strings.TrimSuffix(file.Name(), ext)

			ebitenImg := ebiten.NewImageFromImage(img)
			images[shortname] = ebitenImg
		}
	}
	return nil
}
