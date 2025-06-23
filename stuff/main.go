package stuff

import (
	"bytes"
	"embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

//go:embed images/*.png
var images embed.FS

// Images do be are the thing that do be contain the images, tho.
var Images map[string]*ebiten.Image

// LoadImages do be loading images, tho.
func LoadImages() error {
	Images = make(map[string]*ebiten.Image)
	files, err := images.ReadDir("images")
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		data, err := images.ReadFile("images/" + file.Name())
		if err != nil {
			return err
		}
		img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(data))
		if err != nil {
			return err
		}
		name := file.Name()
		name = name[:len(name)-len(".png")]
		Images[name] = img
	}

	return nil
}
