-> SNAKE <-
===========

```ascii
   _____             _        
  / ____|           | |       
 | (___  _ __   __ _| | _____ 
  \___ \| '_ \ / _` | |/ / _ \\
  ____) | | | | (_| |   <  __/
 |_____/|_| |_|\__,_|_|\_\___|
 							 
___________o/    \o___________
___________|_ MM _|___________

```

-------------------------------

-> The classic snake game playable from your terminal <-
========================================================

This package simulates the classic nokia snake game
It's inteded purpose is to entertain my 6 year old son
and also to train neural networks through genetic algorithms
When training is done it allows you to play a game againts
the trained neural nets

-------------------------------

-> usage <-
===========

```sh
	cd cmd/playsnake/
	go build && playsnake > log ; reset 
	# the reset because signals are not interpreted properly
```

The saves the best ai generated from the run to ai.json 
which can be loaded next run by using the -loadai flag.
When an AI is loaded playsnake wont attempt to train 
an AI for the game reducing start-up time

The human snake can be controled using the 
* `a` key for left
* `d` key for right

The snakes are controlled relative to the snake.
Left is left for the snake not west

------------------------------

-> Implementing your own snake <-
=================================

The interface to which to adhere for the game to use your own implementation is as follows
the snake is as follows

```go
// Player us the interface used by the game for getting moves from the player
type Player interface {
	Play(GameState) Move
	SetID(ID)
}

// GamseState allows the player to get a snapshot of the board and gives the player acces to some helper functions
type GameState interface {
	Vision(id ID) []int8
	Life(id ID) float64 // 0 is dead
	Board() Board
}

// Moves must have an ID, the move interpreded as follows
// Move[0] indicates how much you want to go left
// Move[1] indicates how much you want to go straight
// Move[2] indicates how much you want to go right
// The game will move the player to whichever value is higher
type Move struct {
	Move []float64
	ID   ID
}

```


