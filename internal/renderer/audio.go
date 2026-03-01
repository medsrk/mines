package renderer

import (
	"bytes"
	"io"
	"log"
	"mines/assets"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

type Audio struct {
	audioCtx           *audio.Context
	tileRevealSound    *audio.Player
	initialRevealSound *audio.Player
	flagSound          *audio.Player
	unflagSound        *audio.Player
	explodeSound       *audio.Player
	winSound           *audio.Player
}

func loadAudio() *Audio {
	audioCtx := audio.NewContext(44100)
	tileRevealSound := loadSound(audioCtx, "audio/reveal.ogg")
	initialRevealSound := loadSound(audioCtx, "audio/click.ogg")
	flagSound := loadSound(audioCtx, "audio/flag.ogg")
	unflagSound := loadSound(audioCtx, "audio/unflag.ogg")
	explodeSound := loadSound(audioCtx, "audio/explode.ogg")
	winSound := loadSound(audioCtx, "audio/win.ogg")

	return &Audio{
		audioCtx:           audioCtx,
		tileRevealSound:    tileRevealSound,
		initialRevealSound: initialRevealSound,
		flagSound:          flagSound,
		unflagSound:        unflagSound,
		explodeSound:       explodeSound,
		winSound:           winSound,
	}
}

func loadSound(audioCtx *audio.Context, path string) *audio.Player {
	file, err := assets.FS.Open(path)
	if err != nil {
		log.Printf("Warning: loadSound: failed to open sound file %s: %v", path, err)
		return nil
	}
	defer file.Close()

	decodedStream, err := vorbis.DecodeWithSampleRate(audioCtx.SampleRate(), file)
	if err != nil {
		log.Printf("Warning: loadSound: failed to decode OGG sound file %s: %v", path, err)
		return nil
	}

	var pcmBuffer bytes.Buffer
	if _, err := io.Copy(&pcmBuffer, decodedStream); err != nil {
		log.Printf("Warning: loadSound: failed to buffer decoded OGG stream for %s: %v", path, err)
		return nil
	}

	pcmDataReader := bytes.NewReader(pcmBuffer.Bytes())
	player, err := audioCtx.NewPlayer(pcmDataReader)
	if err != nil {
		log.Printf("Warning: loadSound: failed to create audio player for %s from buffered data: %v", path, err)
		return nil
	}
	return player
}
