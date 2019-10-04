package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wouterbeets/snake"
)

func TestSnakeNewGame(t *testing.T) {
	for i := 0; i < 100; i++ {
		g, err := snake.NewGame(30, 30, []snake.Player{
			&snake.Random{},
			&snake.Random{},
			&snake.Random{},
			&snake.Random{},
			&snake.Random{},
			&snake.Random{},
			&snake.Random{},
			&snake.Random{},
			&snake.Random{},
			&snake.Random{},
			&snake.Random{},
		})
		require.NoError(t, err)
		for i := 0; i < 1000; i++ {
			gameOver, _ := g.PlayRound()
			if gameOver {
				fmt.Println("game over at round", i)
				break
			}
			if i == 999 {
				fmt.Println("game finished", i)
			}
		}
	}
}
