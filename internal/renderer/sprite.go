package renderer

import (
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
	hidden, _, err := ebitenutil.NewImageFromFile("assets/hidden.png")
	if err != nil {
		return nil, err
	}

	revealed, _, err := ebitenutil.NewImageFromFile("assets/revealed.png")
	if err != nil {
		return nil, err
	}

	flag, _, err := ebitenutil.NewImageFromFile("assets/flag.png")
	if err != nil {
		return nil, err
	}

	mine, _, err := ebitenutil.NewImageFromFile("assets/mine.png")
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
