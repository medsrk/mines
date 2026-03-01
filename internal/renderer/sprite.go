package renderer

import (
	"mines/assets"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Sprites struct {
	hidden   *ebiten.Image
	revealed *ebiten.Image
	flag     *ebiten.Image
	mine     *ebiten.Image
}

func LoadSprites() (*Sprites, error) {
	// load image from file
	hidden, err := loadImage("hidden.png")
	if err != nil {
		return nil, err
	}

	revealed, err := loadImage("revealed.png")
	if err != nil {
		return nil, err
	}

	flag, err := loadImage("flag.png")
	if err != nil {
		return nil, err
	}

	mine, err := loadImage("mine.png")
	if err != nil {
		return nil, err
	}

	return &Sprites{
		hidden:   hidden,
		revealed: revealed,
		flag:     flag,
		mine:     mine,
	}, nil
}

func loadImage(name string) (*ebiten.Image, error) {
	f, err := assets.FS.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := ebitenutil.NewImageFromReader(f)
	return img, err
}
