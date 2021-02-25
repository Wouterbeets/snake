package snake

import (
	"math/rand"
	"time"
)

type Move struct {
	Move []float64
	ID   ID
}

type choice string

const (
	left     choice = "left"
	right    choice = "right"
	straight choice = "straight"
)

func (m Move) getChoice() choice {
	if len(m.Move) != 3 {
		return straight
	}
	if m.Move[0] > m.Move[1] && m.Move[0] > m.Move[2] {
		return left
	} else if m.Move[1] > m.Move[2] && m.Move[1] > m.Move[0] {
		return straight
	} else if m.Move[2] > m.Move[0] && m.Move[2] > m.Move[1] {
		return right
	}
	return straight
}

type Player interface {
	Play(*Game) Move
	SetID(ID)
}

type Human struct {
	ID        ID
	Input     chan rune
	Framerate time.Duration
}

func (h *Human) Play(gameState *Game) Move {
	var key rune
	select {
	case key = <-h.Input:
	case <-time.After(h.Framerate):
		key = '0'
	}
	switch key {
	case 'a':
		return Move{Move: []float64{1, 0, 0}, ID: h.ID}
	case 'w':
		return Move{Move: []float64{0, 1, 0}, ID: h.ID}
	case 'd':
		return Move{Move: []float64{0, 0, 1}, ID: h.ID}
	default:
		return Move{Move: []float64{0, 1, 0}, ID: h.ID}
	}
}

func (h *Human) SetID(id ID) {
	h.ID = id
}

type Random struct {
	ID ID
}

func (r *Random) Play(gameState *Game) Move {
	return Move{Move: []float64{rand.Float64(), rand.Float64(), rand.Float64()}, ID: r.ID}
}

func (r *Random) SetID(id ID) {
	r.ID = id
}
