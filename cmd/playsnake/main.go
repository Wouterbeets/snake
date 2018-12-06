package main

import (
	"github.com/wouterbeets/snake"
	"github.com/wouterbeets/term"
)

func main() {
	g, err := snake.NewGame(20, 20, []snake.Player{&snake.Human{}})
	if err != nil {
		panic(err)
	}
	s := term.Screen{Input: make(chan [][]rune)}
	go s.Run()
	for {
		gameOver, state := g.PlayRound()
		s.Input <- stateToRune(state)
		if gameOver {
			return
		}
	}
}

func stateToRune(state snake.Board) (disp [][]rune) {
	disp = make([][]rune, len(state))
	for i := range disp {
		disp[i] = make([]rune, len(state[i]))
	}

	for y, row := range state {
		for x := range row {
			disp[y][x] = runes[state[y][x]]
		}
	}
	return disp
}

var runes map[int8]rune

func init() {
	runes = map[int8]rune{
		0: ' ',
		1: '■',
		2: '•',
		3: 'x',
	}
}
