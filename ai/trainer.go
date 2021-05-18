package ai

import (
	"github.com/klokare/evo"
	"github.com/wouterbeets/snake"
)

type Trainer struct {
}

// Search doesn't use the eval fuction
func (s Trainer) Search(eval evo.Evaluator, phenomes []evo.Phenome) (results []evo.Result, err error) {
	players := make(map[int64]snake.Player, len(phenomes))
	var playerSlice []snake.Player
	for _, p := range phenomes {
		ai := &NetWrapper{Ai: p.Network}
		players[p.ID] = ai
		playerSlice = append(playerSlice, ai)
	}

	g, _ := snake.NewGame(len(players)*3, len(players)*3, playerSlice, len(players)*20)
	rounds := 100000
	for i := 0; i < rounds; i++ {
		gameOver, _ := g.PlayRound()
		for id, player := range players {
			// if we have performance issues we can implement a Dead() interface
			// this interface could allow us to notify which snake is dead by sending their id on a channel
			pl := g.PlayerLen(player.(*NetWrapper).ID)
			if pl > player.(*NetWrapper).maxLen {
				player.(*NetWrapper).maxLen = pl
			}
			if !g.Alive(player.(*NetWrapper).ID) {
				// add result
				fit := float64(i) / float64(rounds)
				maxLen := float64(player.(*NetWrapper).maxLen)
				r := evo.Result{
					ID:      id,
					Fitness: fit + maxLen/10,
				}
				results = append(results, r)
				delete(players, id)
			}
		}
		if len(players) == 0 || gameOver {
			return results, nil
		}
	}
	return
}
