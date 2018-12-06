package snake

import (
	"math/rand"
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
	ID ID
}

func (h *Human) Play(gameState *Game) move {
	return move{move: []float64{rand.Float64(), rand.Float64(), rand.Float64()}, ID: h.ID}
}

func (h *Human) SetID(id ID) {
	h.ID = id
}
