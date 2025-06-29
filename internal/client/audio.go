package client

import (
	"bytes"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/ketMix/ebijam25/stuff"
)

var audioContext = audio.NewContext(48000)
var audioPlayer *audio.Player

func PlayAudio(name string) {
	if audioContext == nil {
		return
	}
	if audioPlayer != nil {
		audioPlayer.Close()
	}
	data := stuff.GetAudio(name)
	if data == nil {
		return
	}

	s, err := vorbis.DecodeWithSampleRate(audioContext.SampleRate(), bytes.NewReader(data))
	if err != nil {
		return
	}
	audioPlayer, err = audio.NewPlayer(audioContext, s)
	if err != nil {
		return
	}
	audioPlayer.SetVolume(0.25)
	audioPlayer.Play()
}

func StopAudio() {
	if audioPlayer != nil {
		audioPlayer.Close()
		audioPlayer = nil
	}
}

func AudioPlaying() bool {
	if audioPlayer == nil {
		return false
	}
	return audioPlayer.IsPlaying()
}
