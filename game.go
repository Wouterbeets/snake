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
	ret := b[y][x]
	if ret > 2 {
		ret = 2
	}
	return ret
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
	life   float64
	maxLen int
}

func (g *Game) PlayerLen(id ID) int {
	if _, ok := g.Players[id]; ok {
		return len(g.Players[id].snake.position)
	}
	return 0
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
	currentLen := float64(len(g.Players[id].snake.position) - 2)
	maxLen := float64(g.Players[id].maxLen - 1)
	life := g.Players[id].life
	step := (1.0 - (currentLen / maxLen)) / (maxLen - currentLen)
	return currentLen/maxLen + (step * life)
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

func (g *Game) ThirdLayerVision(id ID) []int8 {
	s := g.Players[id].snake
	pos := s.head()
	var vis []int8
	switch s.getDir() {
	case north:
		/*
		      5
		     4V6
		    3VVV7
		   2VVXVV8
		    1VXV9
		     0X10
		*/
		vis = append(vis, g.board.At(pos.y+2, pos.x-1))
		vis = append(vis, g.board.At(pos.y+1, pos.x-2))
		vis = append(vis, g.board.At(pos.y, pos.x-3))
		vis = append(vis, g.board.At(pos.y-1, pos.x-2))
		vis = append(vis, g.board.At(pos.y-2, pos.x-1))
		vis = append(vis, g.board.At(pos.y-3, pos.x))
		vis = append(vis, g.SecondLayerVision(id)...)
		vis = append(vis, g.board.At(pos.y-2, pos.x+1))
		vis = append(vis, g.board.At(pos.y-1, pos.x+2))
		vis = append(vis, g.board.At(pos.y, pos.x+3))
		vis = append(vis, g.board.At(pos.y+1, pos.x+2))
		vis = append(vis, g.board.At(pos.y+2, pos.x+1))
	case east:
		/*2
		 1V3
		0VVV4
		XXXVV5
		1VVV6
		 9V7
		  8
		*/
		vis = append(vis, g.board.At(pos.y-1, pos.x-2))
		vis = append(vis, g.board.At(pos.y-2, pos.x-1))
		vis = append(vis, g.board.At(pos.y-3, pos.x))
		vis = append(vis, g.board.At(pos.y-2, pos.x+1))
		vis = append(vis, g.board.At(pos.y-1, pos.x+2))
		vis = append(vis, g.board.At(pos.y, pos.x+3))
		vis = append(vis, g.SecondLayerVision(id)...)
		vis = append(vis, g.board.At(pos.y+1, pos.x+2))
		vis = append(vis, g.board.At(pos.y+2, pos.x+1))
		vis = append(vis, g.board.At(pos.y+3, pos.x))
		vis = append(vis, g.board.At(pos.y+2, pos.x-1))
		vis = append(vis, g.board.At(pos.y+1, pos.x-2))
	case south:
		vis = append(vis, g.board.At(pos.y-2, pos.x+1))
		vis = append(vis, g.board.At(pos.y-1, pos.x+2))
		vis = append(vis, g.board.At(pos.y, pos.x+3))
		vis = append(vis, g.board.At(pos.y+1, pos.x+2))
		vis = append(vis, g.board.At(pos.y+2, pos.x+1))
		vis = append(vis, g.board.At(pos.y+3, pos.x))
		vis = append(vis, g.SecondLayerVision(id)...)
		vis = append(vis, g.board.At(pos.y+2, pos.x-1))
		vis = append(vis, g.board.At(pos.y+1, pos.x-2))
		vis = append(vis, g.board.At(pos.y, pos.x-3))
		vis = append(vis, g.board.At(pos.y-1, pos.x-2))
		vis = append(vis, g.board.At(pos.y-2, pos.x-1))
	case west:
		vis = append(vis, g.board.At(pos.y+1, pos.x+2))
		vis = append(vis, g.board.At(pos.y+2, pos.x+1))
		vis = append(vis, g.board.At(pos.y+3, pos.x))
		vis = append(vis, g.board.At(pos.y+2, pos.x-1))
		vis = append(vis, g.board.At(pos.y+1, pos.x-2))
		vis = append(vis, g.board.At(pos.y, pos.x-3))
		vis = append(vis, g.SecondLayerVision(id)...)
		vis = append(vis, g.board.At(pos.y-1, pos.x-2))
		vis = append(vis, g.board.At(pos.y-2, pos.x-1))
		vis = append(vis, g.board.At(pos.y-3, pos.x))
		vis = append(vis, g.board.At(pos.y-2, pos.x+1))
		vis = append(vis, g.board.At(pos.y-1, pos.x+2))
	}
	return vis
}

/*
     3
    2V4
   1VXV5
    0X6
*/
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
		vis = append(vis, g.PrimordialVision(id)...)
		vis = append(vis, g.board.At(pos.y-1, pos.x+1))
		vis = append(vis, g.board.At(pos.y, pos.x+2))
		vis = append(vis, g.board.At(pos.y+1, pos.x+2))
	case east:
		vis = append(vis, g.board.At(pos.y-1, pos.x-1))
		vis = append(vis, g.board.At(pos.y-2, pos.x))
		vis = append(vis, g.board.At(pos.y-1, pos.x+1))
		vis = append(vis, g.board.At(pos.y, pos.x+2))
		vis = append(vis, g.PrimordialVision(id)...)
		vis = append(vis, g.board.At(pos.y+1, pos.x+1))
		vis = append(vis, g.board.At(pos.y+2, pos.x))
		vis = append(vis, g.board.At(pos.y+1, pos.x-1))
	case south:
		vis = append(vis, g.board.At(pos.y-1, pos.x+1))
		vis = append(vis, g.board.At(pos.y, pos.x+2))
		vis = append(vis, g.board.At(pos.y+1, pos.x+1))
		vis = append(vis, g.board.At(pos.y+2, pos.x))
		vis = append(vis, g.PrimordialVision(id)...)
		vis = append(vis, g.board.At(pos.y+1, pos.x-1))
		vis = append(vis, g.board.At(pos.y, pos.x-2))
		vis = append(vis, g.board.At(pos.y-1, pos.x-2))
	case west:
		vis = append(vis, g.board.At(pos.y+1, pos.x+1))
		vis = append(vis, g.board.At(pos.y+2, pos.x))
		vis = append(vis, g.board.At(pos.y+1, pos.x-1))
		vis = append(vis, g.board.At(pos.y, pos.x-2))
		vis = append(vis, g.PrimordialVision(id)...)
		vis = append(vis, g.board.At(pos.y-1, pos.x-1))
		vis = append(vis, g.board.At(pos.y-2, pos.x))
		vis = append(vis, g.board.At(pos.y-1, pos.x+1))
	}
	return vis
}

/*
    1
   0X2
*/
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

func (g *Game) SensorVision(id ID) []int8 {
	s := g.Players[id].snake
	pos := s.head()
	var vis []int8
	switch s.getDir() {
	case north:
		vis = append(vis, g.board.At(pos.y, pos.x-1))
		vis = append(vis, g.board.At(pos.y, pos.x-2))
		vis = append(vis, g.board.At(pos.y, pos.x-3))
		vis = append(vis, g.board.At(pos.y, pos.x-4))
		vis = append(vis, g.board.At(pos.y, pos.x-5))
		vis = append(vis, g.board.At(pos.y-1, pos.x-1))
		vis = append(vis, g.board.At(pos.y-2, pos.x-2))
		vis = append(vis, g.board.At(pos.y-3, pos.x-3))
		vis = append(vis, g.board.At(pos.y-4, pos.x-4))
		vis = append(vis, g.board.At(pos.y-5, pos.x-5))
		vis = append(vis, g.board.At(pos.y-1, pos.x))
		vis = append(vis, g.board.At(pos.y-2, pos.x))
		vis = append(vis, g.board.At(pos.y-3, pos.x))
		vis = append(vis, g.board.At(pos.y-4, pos.x))
		vis = append(vis, g.board.At(pos.y-5, pos.x))
		vis = append(vis, g.board.At(pos.y-1, pos.x+1))
		vis = append(vis, g.board.At(pos.y-2, pos.x+2))
		vis = append(vis, g.board.At(pos.y-3, pos.x+3))
		vis = append(vis, g.board.At(pos.y-4, pos.x+4))
		vis = append(vis, g.board.At(pos.y-5, pos.x+5))
		vis = append(vis, g.board.At(pos.y, pos.x+1))
		vis = append(vis, g.board.At(pos.y, pos.x+2))
		vis = append(vis, g.board.At(pos.y, pos.x+3))
		vis = append(vis, g.board.At(pos.y, pos.x+4))
		vis = append(vis, g.board.At(pos.y, pos.x+5))
	case east:
		vis = append(vis, g.board.At(pos.y-1, pos.x))
		vis = append(vis, g.board.At(pos.y-2, pos.x))
		vis = append(vis, g.board.At(pos.y-3, pos.x))
		vis = append(vis, g.board.At(pos.y-4, pos.x))
		vis = append(vis, g.board.At(pos.y-5, pos.x))
		vis = append(vis, g.board.At(pos.y-1, pos.x+1))
		vis = append(vis, g.board.At(pos.y-2, pos.x+2))
		vis = append(vis, g.board.At(pos.y-3, pos.x+3))
		vis = append(vis, g.board.At(pos.y-4, pos.x+4))
		vis = append(vis, g.board.At(pos.y-5, pos.x+5))
		vis = append(vis, g.board.At(pos.y, pos.x+1))
		vis = append(vis, g.board.At(pos.y, pos.x+2))
		vis = append(vis, g.board.At(pos.y, pos.x+3))
		vis = append(vis, g.board.At(pos.y, pos.x+4))
		vis = append(vis, g.board.At(pos.y, pos.x+5))
		vis = append(vis, g.board.At(pos.y+1, pos.x+1))
		vis = append(vis, g.board.At(pos.y+2, pos.x+2))
		vis = append(vis, g.board.At(pos.y+3, pos.x+3))
		vis = append(vis, g.board.At(pos.y+4, pos.x+4))
		vis = append(vis, g.board.At(pos.y+5, pos.x+5))
		vis = append(vis, g.board.At(pos.y+1, pos.x))
		vis = append(vis, g.board.At(pos.y+2, pos.x))
		vis = append(vis, g.board.At(pos.y+3, pos.x))
		vis = append(vis, g.board.At(pos.y+4, pos.x))
		vis = append(vis, g.board.At(pos.y+5, pos.x))
	case south:
		vis = append(vis, g.board.At(pos.y, pos.x+1))
		vis = append(vis, g.board.At(pos.y, pos.x+2))
		vis = append(vis, g.board.At(pos.y, pos.x+3))
		vis = append(vis, g.board.At(pos.y, pos.x+4))
		vis = append(vis, g.board.At(pos.y, pos.x+5))
		vis = append(vis, g.board.At(pos.y+1, pos.x+1))
		vis = append(vis, g.board.At(pos.y+2, pos.x+2))
		vis = append(vis, g.board.At(pos.y+3, pos.x+3))
		vis = append(vis, g.board.At(pos.y+4, pos.x+4))
		vis = append(vis, g.board.At(pos.y+5, pos.x+5))
		vis = append(vis, g.board.At(pos.y+1, pos.x))
		vis = append(vis, g.board.At(pos.y+2, pos.x))
		vis = append(vis, g.board.At(pos.y+3, pos.x))
		vis = append(vis, g.board.At(pos.y+4, pos.x))
		vis = append(vis, g.board.At(pos.y+5, pos.x))
		vis = append(vis, g.board.At(pos.y+1, pos.x-1))
		vis = append(vis, g.board.At(pos.y+2, pos.x-2))
		vis = append(vis, g.board.At(pos.y+3, pos.x-3))
		vis = append(vis, g.board.At(pos.y+4, pos.x-4))
		vis = append(vis, g.board.At(pos.y+5, pos.x-5))
		vis = append(vis, g.board.At(pos.y, pos.x-1))
		vis = append(vis, g.board.At(pos.y, pos.x-2))
		vis = append(vis, g.board.At(pos.y, pos.x-3))
		vis = append(vis, g.board.At(pos.y, pos.x-4))
		vis = append(vis, g.board.At(pos.y, pos.x-5))
	case west:
		vis = append(vis, g.board.At(pos.y+1, pos.x))
		vis = append(vis, g.board.At(pos.y+2, pos.x))
		vis = append(vis, g.board.At(pos.y+3, pos.x))
		vis = append(vis, g.board.At(pos.y+4, pos.x))
		vis = append(vis, g.board.At(pos.y+5, pos.x))
		vis = append(vis, g.board.At(pos.y+1, pos.x+1))
		vis = append(vis, g.board.At(pos.y+2, pos.x+2))
		vis = append(vis, g.board.At(pos.y+3, pos.x+3))
		vis = append(vis, g.board.At(pos.y+4, pos.x+4))
		vis = append(vis, g.board.At(pos.y+5, pos.x+5))
		vis = append(vis, g.board.At(pos.y, pos.x+1))
		vis = append(vis, g.board.At(pos.y, pos.x+2))
		vis = append(vis, g.board.At(pos.y, pos.x+3))
		vis = append(vis, g.board.At(pos.y, pos.x+4))
		vis = append(vis, g.board.At(pos.y, pos.x+5))
		vis = append(vis, g.board.At(pos.y-1, pos.x+1))
		vis = append(vis, g.board.At(pos.y-2, pos.x+2))
		vis = append(vis, g.board.At(pos.y-3, pos.x+3))
		vis = append(vis, g.board.At(pos.y-4, pos.x+4))
		vis = append(vis, g.board.At(pos.y-5, pos.x+5))
		vis = append(vis, g.board.At(pos.y-1, pos.x))
		vis = append(vis, g.board.At(pos.y-2, pos.x))
		vis = append(vis, g.board.At(pos.y-3, pos.x))
		vis = append(vis, g.board.At(pos.y-4, pos.x))
		vis = append(vis, g.board.At(pos.y-5, pos.x))
	}
	return vis
}

func (g *Game) Vision(id ID) []int8 {
	return g.SensorVision(id)
}

// PlayMove takes a move and aplies it to the game
func (g *Game) PlayMove(m Move) (dead bool) {
	p := g.Players[m.ID]
	newPos := p.snake.newHeadPos(m)

	if g.board[newPos.y][newPos.x] == food {
		p.snake.moveTo(newPos, true)
		g.restoreLife(m.ID)
		g.board[newPos.y][newPos.x] = int8(m.ID)
		g.newFood()
		if g.Players[m.ID].maxLen < len(p.snake.position) {
			p.maxLen = len(p.snake.position)
		}
		g.Players[m.ID] = p
		return false
	}

	if g.board[newPos.y][newPos.x] != empty {
		return true
	}
	t := p.snake.tail()
	p.snake.moveTo(newPos, false)
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
