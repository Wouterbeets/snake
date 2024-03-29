package ai

import (
	"github.com/klokare/evo"
	"github.com/wouterbeets/snake"
	"gonum.org/v1/gonum/mat"
)

func (n *NetWrapper) Play(g snake.GameState) snake.Move {
	vis := g.Vision(n.ID)
	life := g.Life(n.ID)
	inf := make([]float64, len(vis))
	inf = append(inf, life)

	for i := range vis {
		inf[i] = float64(vis[i])
	}
	in := mat.NewDense(1, len(inf), inf)
	out, err := n.Ai.Activate(in)
	if err != nil {
		panic("error in ai")
	}
	ret := []float64{out.At(0, 0), out.At(0, 1), out.At(0, 2)}
	return snake.Move{Move: ret, ID: n.ID}
}

func (n *NetWrapper) SetID(id snake.ID) {
	n.ID = id
}

type NetWrapper struct {
	Ai     evo.Network
	ID     snake.ID
	maxLen int
}

func (e Evaluator) Evaluate(p evo.Phenome) (r evo.Result, err error) {

	r.ID = p.ID
	type eval struct {
		in  *mat.Dense
		out *mat.Dense
	}

	player := NetWrapper{Ai: p.Network}
	g, err := snake.NewGame(20, 20, []snake.Player{
		&player,
	}, 1)

	rounds := 1000
	var snakeLen int
	var maxLen int
	for i := 0; i < rounds; i++ {
		snakeLen = g.PlayerLen(player.ID)
		if snakeLen > maxLen {
			maxLen = snakeLen
		}
		gameOver, _ := g.PlayRound()
		if gameOver || !g.Alive(player.ID) {
			r.Fitness += float64(i) / float64(rounds)
			break
		}
		if i == rounds-1 {
			r.Fitness = 1 * (float64(maxLen) / 10)
		}
	}
	return r, nil
}

type Evaluator struct {
}
