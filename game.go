package snake

import (
	"errors"
	"fmt"
	"sync"
)

const (
	empty = iota
	wall
	food = -1
)

// ID is the player's id on the board
type ID int8

// Game holds the board and the players
type Game struct {
	Board   Board
	Players map[ID]playerInfo
}

// Board holds the game state
type Board [][]int8

type Position struct {
	x int
	y int
}

type playerInfo struct {
	Player
	*snake
}

func newBoard(height, width int) Board {
	board := make(Board, height)
	for i := 0; i < height; i++ {
		board[i] = make([]int8, width)
	}
	for i := range board[0] {
		board[0][i] = wall
		board[len(board)-1][i] = wall
	}
	for i := 0; i < len(board); i++ {
		board[i][0], board[i][len(board[0])-1] = wall, wall
	}
	return board
}

// NewGame inits a new snake game with a size and list of players
func NewGame(height, width int, players []Player) (*Game, error) {
	if height < 5 || width < 5 {
		return nil, errors.New("size too small")
	}

	board := newBoard(height, width)

	g := &Game{
		Board:   board,
		Players: make(map[ID]playerInfo, len(players)),
	}

	for i, p := range players {
		p.SetID(ID(i + 3))
		g.Players[ID(i+3)] = playerInfo{
			Player: p,
			snake:  newSnake(board, ID(i+3)),
		}
	}

	g.newFood()
	g.newFood()
	g.newFood()
	g.newFood()
	g.newFood()
	g.newFood()
	g.newFood()
	g.newFood()
	g.newFood()
	g.newFood()
	return g, nil
}

func (g *Game) PlayerLen(id ID) int {
	return len(g.Players[id].snake.position)
}

func (g *Game) Alive(id ID) bool {
	if _, ok := g.Players[id]; !ok {
		return false
	}
	return true
}

// PlayRound processes one game tick
func (g *Game) PlayRound() (gameOver bool, state Board) {
	cmove := make(chan Move, len(g.Players))

	var wg sync.WaitGroup

	for i := range g.Players {
		wg.Add(1)
		go func(p Player) {
			defer wg.Done()
			m := p.Play(g)
			cmove <- m
		}(g.Players[i])
	}
	wg.Wait()
	close(cmove)

	for move := range cmove {
		if _, ok := g.Players[move.ID]; !ok {
			continue
		}
		if dead := g.PlayMove(move); dead {
			for _, pos := range g.Players[move.ID].position {
				g.Board[pos.y][pos.x] = empty
			}
			delete(g.Players, move.ID)
			if len(g.Players) == 0 {
				return true, state
			}
		}
	}
	return false, g.Board
}

func (g *Game) print() {
	for _, r := range g.Board {
		fmt.Println(r)
	}
}

func (g *Game) newFood() {
	for {
		pos := randomPos(len(g.Board[0]), len(g.Board))
		if g.Board[pos.y][pos.x] == empty {
			g.Board[pos.y][pos.x] = food
			return
		}
	}
}

func (g *Game) Vision(id ID) []int8 {
	s := g.Players[id].snake
	pos := s.head()
	var vis []int8
	switch s.getDir() {
	case north:
		vis = append(vis, g.Board[pos.y][pos.x-1])
		vis = append(vis, g.Board[pos.y-1][pos.x])
		vis = append(vis, g.Board[pos.y][pos.x+1])
		return vis
	case east:
		vis = append(vis, g.Board[pos.y-1][pos.x])
		vis = append(vis, g.Board[pos.y][pos.x+1])
		vis = append(vis, g.Board[pos.y+1][pos.x])
		return vis
	case south:
		vis = append(vis, g.Board[pos.y][pos.x+1])
		vis = append(vis, g.Board[pos.y+1][pos.x])
		vis = append(vis, g.Board[pos.y][pos.x-1])
		return vis
	case west:
		vis = append(vis, g.Board[pos.y+1][pos.x])
		vis = append(vis, g.Board[pos.y][pos.x-1])
		vis = append(vis, g.Board[pos.y-1][pos.x])
		return vis
	default:
		return vis
	}
}

// PlayMove takes a move and aplies it to the game
func (g *Game) PlayMove(m Move) (dead bool) {
	s := g.Players[m.ID].snake
	newPos := s.newHeadPos(m)

	if g.Board[newPos.y][newPos.x] == food {
		s.moveTo(newPos, true)
		g.newFood()
		return false
	}

	if g.Board[newPos.y][newPos.x] != empty {
		return true
	}
	t := s.tail()
	s.moveTo(newPos, false)
	g.Board[newPos.y][newPos.x] = int8(m.ID)
	g.Board[t.y][t.x] = empty
	return false
}
