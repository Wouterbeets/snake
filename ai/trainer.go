package ai

import (
	"github.com/klokare/evo"
	"github.com/wouterbeets/snake"
)

type Trainer struct {
}

// Search doesn't use the eval fuction
func (s Trainer) Search(eval evo.Evaluator, phenomes []evo.Phenome) (results []evo.Result, err error) {
	var players []snake.Player
	for _, p := range phenomes {
		players = append(players, &NetWrapper{Ai: p.Network})
	}
	g, _ := snake.NewGame(len(players)*3, len(players)*3, players, len(players))
	rounds := 1000
	for i := 0; i < rounds; i++ {
		gameOver, _ := g.PlayRound()
		for _, player := range players {
			// if we have performance issues we can implement a Dead() interface
			// this interface could allow us to notify which snake is dead by sending their id on a channel
			if !g.Alive(player.(*NetWrapper).ID) {
				// add result
				//if i == rounds-1 {
				//	r.Fitness = 1 * (float64(maxLen) / 10)
				//}
			}
		}
		if gameOver {
			// check all players and return result
		}
	}
	return nil, nil
}
