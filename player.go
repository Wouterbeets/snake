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

type Human struct {
	ID        ID
	Input     chan rune
	Framerate time.Duration
}

func (h *Human) Play(gameState GameState) Move {
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

func (r *Random) Play(gameState GameState) Move {
	v := gameState.Vision(r.ID)
	left := rand.Float64()
	straight := rand.Float64()
	right := rand.Float64()
	if v[4] >= wall {
		left = 0
	}
	if v[5] >= wall {
		straight = 0
	}
	if v[6] >= wall {
		right = 0
	}
	return Move{Move: []float64{left, straight, right}, ID: r.ID}
}

func (r *Random) SetID(id ID) {
	r.ID = id
}
