package stuff

import (
	"bytes"
	"embed"
	"fmt"

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

// GetImage do be getting an image by name, tho.
func GetImage(name string) *ebiten.Image {
	if img, ok := Images[name]; ok {
		return img
	}
	return nil
}

var names []string

//go:embed names-moby-word-lists-grady-ward.txt
var namesFile embed.FS

// LoadNames gets our names.
func LoadNames() error {
	data, err := namesFile.ReadFile("names-moby-word-lists-grady-ward.txt")
	if err != nil {
		return err
	}
	// Split strings.
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) > 0 {
			names = append(names, string(line))
		}
	}

	return nil
}

// GetName do be getting a name by number, tho. It modulo wraps.
func GetName(num int) string {
	return names[num%len(names)]
}

//go:embed audio/*.ogg
var audio embed.FS
var Audio map[string][]byte

func LoadAudio() error {
	Audio = make(map[string][]byte)
	files, err := audio.ReadDir("audio")
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		data, err := audio.ReadFile("audio/" + file.Name())
		if err != nil {
			return err
		}
		name := file.Name()
		name = name[:len(name)-len(".ogg")]
		fmt.Println("Loading audio:", name)
		Audio[name] = data
	}

	return nil
}

func GetAudio(name string) []byte {
	if audioData, ok := Audio[name]; ok {
		return audioData
	}
	return nil
}
