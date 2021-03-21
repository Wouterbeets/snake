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
	board   Board
	Players map[ID]playerInfo
}

// The Board holds the game state
type Board [][]int8

// The Player is the inteface called by the game get a move from the player
type Player interface {
	Play(GameState) Move
	SetID(ID)
}

type GameState interface {
	Vision(id ID) []int8
	Life(id ID) float64 // 0 is dead
	Board() Board
}

func (b Board) At(y, x int) int8 {
	if y >= len(b) {
		return wall
	}
	if y < 0 {
		return wall
	}
	if x >= len(b[y]) {
		return wall
	}
	if x < 0 {
		return wall
	}
	return b[y][x]
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
func NewGame(height, width int, players []Player, nbFoodOnMap int) (*Game, error) {
	if height < 5 || width < 5 {
		return nil, errors.New("size too small")
	}

	g := &Game{
		board:   newBoard(height, width),
		Players: make(map[ID]playerInfo, len(players)),
	}

	// Init players
	for i, p := range players {
		p.SetID(ID(i + 2))
		g.Players[ID(i+2)] = playerInfo{
			Player: p,
			snake:  newSnake(g.board, ID(i+2)),
			life:   1,
		}
	}

	// Generate food
	for i := 0; i <= nbFoodOnMap; i++ {
		g.newFood()
	}
	return g, nil
}

type Position struct {
	x int
	y int
}

type playerInfo struct {
	Player
	*snake
	life float64
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

// Board returns a copy of the board
func (g *Game) Board() (b Board) {
	copy(b, g.board)
	for i := range g.board {
		copy(b[i], g.board[i])
	}
	return
}

func (g *Game) Life(id ID) float64 {
	return g.Players[id].life * float64(g.PlayerLen(id))
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
				g.board[pos.y][pos.x] = empty
			}
			delete(g.Players, move.ID)
			if len(g.Players) == 0 {
				return true, state
			}
		}
	}
	return false, g.board
}

func (g *Game) print() {
	for _, r := range g.board {
		fmt.Println(r)
	}
}

func (g *Game) newFood() {
	for {
		pos := randomPos(len(g.board[0]), len(g.board))
		if g.board[pos.y][pos.x] == empty {
			g.board[pos.y][pos.x] = food
			return
		}
	}
}

func (g *Game) SecondLayerVision(id ID) []int8 {
	s := g.Players[id].snake
	pos := s.head()
	var vis []int8
	switch s.getDir() {
	case north:
		vis = append(vis, g.board.At(pos.y+1, pos.x-1))
		vis = append(vis, g.board.At(pos.y, pos.x-2))
		vis = append(vis, g.board.At(pos.y-1, pos.x-1))
		vis = append(vis, g.board.At(pos.y-2, pos.x))
		vis = append(vis, g.board.At(pos.y-1, pos.x+1))
		vis = append(vis, g.board.At(pos.y, pos.x+2))
		vis = append(vis, g.board.At(pos.y+1, pos.x+2))
	case east:
		vis = append(vis, g.board.At(pos.y-1, pos.x-1))
		vis = append(vis, g.board.At(pos.y-2, pos.x))
		vis = append(vis, g.board.At(pos.y-1, pos.x+1))
		vis = append(vis, g.board.At(pos.y, pos.x+2))
		vis = append(vis, g.board.At(pos.y+1, pos.x+1))
		vis = append(vis, g.board.At(pos.y+2, pos.x))
		vis = append(vis, g.board.At(pos.y+1, pos.x-1))
	case south:
		vis = append(vis, g.board.At(pos.y-1, pos.x+1))
		vis = append(vis, g.board.At(pos.y, pos.x+2))
		vis = append(vis, g.board.At(pos.y+1, pos.x+1))
		vis = append(vis, g.board.At(pos.y+2, pos.x))
		vis = append(vis, g.board.At(pos.y+1, pos.x-1))
		vis = append(vis, g.board.At(pos.y, pos.x-2))
		vis = append(vis, g.board.At(pos.y-1, pos.x-2))
	case west:
		vis = append(vis, g.board.At(pos.y+1, pos.x+1))
		vis = append(vis, g.board.At(pos.y+2, pos.x))
		vis = append(vis, g.board.At(pos.y+1, pos.x-1))
		vis = append(vis, g.board.At(pos.y, pos.x-2))
		vis = append(vis, g.board.At(pos.y-1, pos.x-1))
		vis = append(vis, g.board.At(pos.y-2, pos.x))
		vis = append(vis, g.board.At(pos.y-1, pos.x+1))
	}
	return vis
}

func (g *Game) PrimordialVision(id ID) []int8 {
	s := g.Players[id].snake
	pos := s.head()
	var vis []int8
	switch s.getDir() {
	case north:
		vis = append(vis, g.board.At(pos.y, pos.x-1))
		vis = append(vis, g.board.At(pos.y-1, pos.x))
		vis = append(vis, g.board.At(pos.y, pos.x+1))
	case east:
		vis = append(vis, g.board.At(pos.y-1, pos.x))
		vis = append(vis, g.board.At(pos.y, pos.x+1))
		vis = append(vis, g.board.At(pos.y+1, pos.x))
	case south:
		vis = append(vis, g.board.At(pos.y, pos.x+1))
		vis = append(vis, g.board.At(pos.y+1, pos.x))
		vis = append(vis, g.board.At(pos.y, pos.x-1))
	case west:
		vis = append(vis, g.board.At(pos.y+1, pos.x))
		vis = append(vis, g.board.At(pos.y, pos.x+1))
		vis = append(vis, g.board.At(pos.y-1, pos.x))
	}
	return vis
}

func (g *Game) Vision(id ID) []int8 {
	vis1 := g.PrimordialVision(id)
	vis2 := g.SecondLayerVision(id)
	return append(vis2[:3], vis1[0], vis1[1], vis1[2], vis2[3], vis2[4], vis2[5], vis2[6])
}

// PlayMove takes a move and aplies it to the game
func (g *Game) PlayMove(m Move) (dead bool) {
	s := g.Players[m.ID].snake
	newPos := s.newHeadPos(m)

	if g.board[newPos.y][newPos.x] == food {
		s.moveTo(newPos, true)
		g.restoreLife(m.ID)
		g.board[newPos.y][newPos.x] = int8(m.ID)
		g.newFood()
		return false
	}

	if g.board[newPos.y][newPos.x] != empty {
		return true
	}
	t := s.tail()
	s.moveTo(newPos, false)
	g.board[newPos.y][newPos.x] = int8(m.ID)
	g.board[t.y][t.x] = empty
	dead = g.reduceLife(m.ID)
	return
}

func (g *Game) reduceLife(id ID) (dead bool) {
	p := g.Players[id]
	p.life -= 0.01
	if p.life <= 0 {
		p.life = 1
		t := p.snake.tail()
		dead = p.reduceSize()
		g.board[t.y][t.x] = empty
	}
	g.Players[id] = p
	return
}

func (g *Game) restoreLife(id ID) {
	p := g.Players[id]
	p.life = 1
	g.Players[id] = p
}
