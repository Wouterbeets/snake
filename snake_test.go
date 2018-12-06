package snake

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSnake(t *testing.T) {
	b := newBoard(10, 10)
	id := ID(7)
	s := newSnake(b, id)
	require.Equal(t, 2, len(s.position))
	for _, pos := range s.position {
		require.True(t, b[pos.y][pos.x] == int8(id), fmt.Sprintf("%+v\n", b))
	}
}

func TestSnakeHead(t *testing.T) {
	s := snake{position: []position{
		{
			x: 5,
			y: 5,
		},
		{
			x: 6,
			y: 5,
		},
	}}
	require.Equal(t, position{x: 6, y: 5}, s.head())
}

func TestSnake(t *testing.T) {
	s := snake{position: []position{
		{
			x: 5,
			y: 5,
		},
		{
			x: 6,
			y: 5,
		},
	}}
	require.Equal(t, position{x: 5, y: 5}, s.body())
	require.Equal(t, position{x: 5, y: 5}, s.tail())
	require.Equal(t, east, s.getDir())

	m := move{move: []float64{0, 1, 0}, ID: ID(7)}
	require.Equal(t, position{x: 7, y: 5}, s.newHeadPos(m))

	m = move{move: []float64{1, 0, 0}, ID: ID(7)}
	require.Equal(t, position{x: 6, y: 4}, s.newHeadPos(m))

	m = move{move: []float64{0, 0, 1}, ID: ID(7)}
	require.Equal(t, position{x: 6, y: 6}, s.newHeadPos(m))

	m = move{move: []float64{0, 1, 0}, ID: ID(7)}
	s.moveTo(s.newHeadPos(m), false)
	require.Equal(t, position{x: 7, y: 5}, s.head())
	require.Equal(t, position{x: 6, y: 5}, s.body())
	require.Equal(t, position{x: 6, y: 5}, s.tail())

	m = move{move: []float64{1, 0, 0}, ID: ID(7)}
	s.moveTo(s.newHeadPos(m), false)
	require.Equal(t, position{x: 7, y: 4}, s.head())
	require.Equal(t, position{x: 7, y: 5}, s.body())
	require.Equal(t, position{x: 7, y: 5}, s.tail())
}
