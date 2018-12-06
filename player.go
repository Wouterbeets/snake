package snake

import (
	"time"
)

type move struct {
	move []float64
	ID   ID
}

type choice string

const (
	left     choice = "left"
	right    choice = "right"
	straight choice = "straight"
)

func (m move) getChoice() choice {
	if len(m.move) != 3 {
		return straight
	}
	if m.move[0] > m.move[1] && m.move[0] > m.move[2] {
		return left
	} else if m.move[1] > m.move[2] && m.move[1] > m.move[0] {
		return straight
	} else if m.move[2] > m.move[0] && m.move[2] > m.move[1] {
		return right
	}
	return straight
}

type Player interface {
	Play(*Game) move
	SetID(ID)
}

type Human struct {
	ID        ID
	Input     chan rune
	Framerate time.Duration
}

func (h *Human) Play(gameState *Game) move {
	var key rune
	select {
	case key = <-h.Input:
	case <-time.After(h.Framerate):
		key = '0'
	}
	switch key {
	case 'a':
		return move{move: []float64{1, 0, 0}, ID: h.ID}
	case 'w':
		return move{move: []float64{0, 1, 0}, ID: h.ID}
	case 'd':
		return move{move: []float64{0, 0, 1}, ID: h.ID}
	default:
		return move{move: []float64{0, 1, 0}, ID: h.ID}
	}
}

func (h *Human) SetID(id ID) {
	h.ID = id
}
