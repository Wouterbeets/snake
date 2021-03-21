package snake

import (
	"math/rand"
	"time"
)

type snake struct {
	position []Position
}

func randomPos(width, height int) Position {
	return Position{x: rand.Intn(width-2) + 1, y: rand.Intn(height-2) + 1}
}

func newSnake(board Board, id ID) *snake {
	rand.Seed(time.Now().UnixNano())
	s := snake{position: make([]Position, 2)}
	for {
		pos := randomPos(len(board[0]), len(board))
		x := pos.x
		y := pos.y
		if board[y][x] == empty && board[y][x+1] == empty {
			board[y][x] = int8(id)
			board[y][x+1] = int8(id)
			s.position[0].x, s.position[0].y = x, y
			s.position[1].x, s.position[1].y = x+1, y
			return &s
		}
	}
}

type direction string

const (
	north direction = "north"
	south direction = "south"
	west  direction = "west"
	east  direction = "east"
)

func (s *snake) reduceSize() bool {
	if len(s.position) <= 2 {
		return true
	}
	s.position = s.position[1:]
	return false
}

func (s *snake) tail() Position {
	return s.position[0]
}

func (s *snake) head() Position {
	if s == nil || len(s.position) == 0 {
		return Position{0, 0}
	}
	return s.position[len(s.position)-1]
}

func (s *snake) body() Position {
	if s == nil || len(s.position) == 0 {
		return Position{0, 0}
	}
	if len(s.position) < 2 {
		return Position{0, 0}
	}
	return s.position[len(s.position)-2]
}

func (s *snake) getDir() direction {
	head := s.head()
	body := s.body()
	if head.x == body.x {
		if head.y > body.y {
			return south
		}
		return north
	} else if head.x > body.x {
		return east
	}
	return west
}

func (s *snake) newHeadPos(m Move) Position {
	dir := s.getDir()
	head := s.head()
	c := m.getChoice()
	var newPos Position
	switch dir {
	case north:
		if choice(c) == left {
			newPos = Position{
				x: head.x - 1,
				y: head.y,
			}
		} else if choice(c) == right {
			newPos = Position{
				x: head.x + 1,
				y: head.y,
			}
		} else if choice(c) == straight {
			newPos = Position{
				x: head.x,
				y: head.y - 1,
			}
		}
	case south:
		if choice(c) == left {
			newPos = Position{
				x: head.x + 1,
				y: head.y,
			}
		} else if choice(c) == right {
			newPos = Position{
				x: head.x - 1,
				y: head.y,
			}
		} else if choice(c) == straight {
			newPos = Position{
				x: head.x,
				y: head.y + 1,
			}
		}
	case west:
		if choice(c) == left {
			newPos = Position{
				x: head.x,
				y: head.y + 1,
			}
		} else if choice(c) == right {
			newPos = Position{
				x: head.x,
				y: head.y - 1,
			}
		} else if choice(c) == straight {
			newPos = Position{
				x: head.x - 1,
				y: head.y,
			}
		}
	case east:
		if choice(c) == left {
			newPos = Position{
				x: head.x,
				y: head.y - 1,
			}
		} else if choice(c) == right {
			newPos = Position{
				x: head.x,
				y: head.y + 1,
			}
		} else if choice(c) == straight {
			newPos = Position{
				x: head.x + 1,
				y: head.y,
			}
		}
	}
	return newPos
}

func (s *snake) moveTo(p Position, food bool) {
	s.position = append(s.position, p)
	if !food {
		s.position = s.position[1:]
	}
}
