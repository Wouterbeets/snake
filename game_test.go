package snake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGame(t *testing.T) {
	width := 20
	height := 20
	game := NewGame(20, 20, []Player{&Human{}})
	require.Equal(t, width, len(game.Board))
	require.Equal(t, height, len(game.Board[0]))
}

func TestNewBoard(t *testing.T) {
	width, height := 20, 10
	board := newBoard(height, width)
	require.Equal(t, height, len(board))
	require.Equal(t, width, len(board[0]))
}
