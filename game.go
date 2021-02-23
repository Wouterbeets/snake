package snake

import (
	"errors"
	"fmt"
	"sync"
)

const (
	empty = iota
	wall
	food
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

type position struct {
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
	return g, nil
}

// PlayRound processes one game tick
func (g *Game) PlayRound() (gameOver bool, state Board) {
	cmove := make(chan move, len(g.Players))
	for i, p := range g.Players {
		fmt.Printf("player %d %+v %+v %+v\n", i, p, p.Player, p.snake)
	}

	var wg sync.WaitGroup

	for i := range g.Players {
		wg.Add(1)
		go func(p Player) {
			defer wg.Done()
			cmove <- p.Play(g)
		}(g.Players[i])
	}
	wg.Wait()
	close(cmove)

	for move := range cmove {
		if _, ok := g.Players[move.ID]; !ok {
			continue
		}
		if dead := g.PlayMove(move); dead {
			fmt.Println("dead", dead)
			for _, pos := range g.Players[move.ID].position {
				g.Board[pos.y][pos.x] = empty
			}
			delete(g.Players, move.ID)
			if len(g.Players) == 0 {
				return true, state
			}
		}
	}
	fmt.Println("------------------------------------------")
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

// PlayMove takes a move and aplies is to the game
func (g *Game) PlayMove(m move) (dead bool) {
	fmt.Println("playmove-----------")
	s := g.Players[m.ID].snake
	newPos := s.newHeadPos(m)
	g.print()
	fmt.Println("newpos", newPos)
	fmt.Println(m)

	if g.Board[newPos.y][newPos.x] == food {
		s.moveTo(newPos, true)
		g.newFood()
		return false
	}

	if g.Board[newPos.y][newPos.x] != empty {
		fmt.Println(g.Board[newPos.y][newPos.x])
		return true
	}
	t := s.tail()
	s.moveTo(newPos, false)
	g.Board[newPos.y][newPos.x] = int8(m.ID)
	g.Board[t.y][t.x] = empty
	return false
}
