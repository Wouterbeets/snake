package ai

import (
	"fmt"
	"log"
	"math"

	"github.com/klokare/evo"
	"github.com/wouterbeets/snake"
	"gonum.org/v1/gonum/mat"
)

func (n *NetWrapper) Play(g *snake.Game) snake.Move {
	vis := g.Vision(n.ID)
	fmt.Println(vis)
	inf := make([]float64, len(vis))
	for i := range vis {
		inf[i] = float64(vis[i])
	}
	in := mat.NewDense(1, len(vis), inf)
	out, err := n.Ai.Activate(in)
	if err != nil {
		panic("error in ai")
		//		return snake.Move{Move: []float64{0, 0, 0}, ID: n.ID}
	}
	ret := []float64{out.At(0, 0), out.At(0, 1), out.At(0, 2)}
	return snake.Move{Move: ret, ID: n.ID}
}

func (n *NetWrapper) SetID(id snake.ID) {
	n.ID = id
}

type NetWrapper struct {
	Ai evo.Network
	ID snake.ID
}

func (e Evaluator) Evaluate(p evo.Phenome) (r evo.Result, err error) {

	r.ID = p.ID
	type eval struct {
		in  *mat.Dense
		out *mat.Dense
	}

	evals := []eval{
		{
			in: mat.NewDense(1, 27, []float64{
				0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}),
			out: mat.NewDense(1, 3, []float64{1, 1, 1}),
		},
		{
			in: mat.NewDense(1, 27, []float64{
				0, 0, 1, 0, 0,
				0, 0, 0, 1, 1, 1, 0, 0, 0,
				0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0,
			}),
			out: mat.NewDense(1, 3, []float64{1, 0, 1}),
		},
		{
			in: mat.NewDense(1, 27, []float64{
				1, 1, 0, 1, 1,
				1, 1, 1, 0, 0, 0, 1, 1, 1,
				1, 1, 1, 1, 0, 0, 0, 0, 0, 1, 1, 1, 1,
			}),
			out: mat.NewDense(1, 3, []float64{0, 1, 0}),
		},
		{
			in: mat.NewDense(1, 27, []float64{
				0, 1, 1, 1, 1,
				0, 1, 1, 1, 1, 1, 1, 1, 1,
				0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			}),
			out: mat.NewDense(1, 3, []float64{1, 0, 0}),
		},
		{
			in: mat.NewDense(1, 27, []float64{
				1, 1, 1, 1, 0,
				1, 1, 1, 1, 1, 1, 1, 1, 0,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0,
			}),
			out: mat.NewDense(1, 3, []float64{0, 0, 1}),
		},
		{
			in: mat.NewDense(1, 27, []float64{
				1, 0, 0, 0, 0,
				1, 1, 0, 0, 0, 0, 0, 0, 0,
				1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}),
			out: mat.NewDense(1, 3, []float64{0, 1, 1}),
		},
		{
			in: mat.NewDense(1, 27, []float64{
				0, 0, 0, 0, 1,
				0, 0, 0, 0, 0, 0, 0, 1, 1,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1,
			}),
			out: mat.NewDense(1, 3, []float64{1, 1, 0}),
		},
	}

	for e := range evals {
		m, err := p.Activate(evals[e].in)

		if err != nil {
			log.Fatal(err)
		}
		row, col := m.Dims()
		for i := 0; i < row; i++ {
			for j := 0; j < col; j++ {
				r.Fitness += 1 - (math.Abs(evals[e].out.At(i, j) - m.At(i, j)))
			}
		}
	}
	fmt.Println(r.Fitness)
	if r.Fitness < 20.9 {
		return r, err
	}
	player := NetWrapper{Ai: p.Network}
	g, err := snake.NewGame(20, 20, []snake.Player{
		&player,
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
	})

	rounds := 30
	var snakeLen int
	for i := 0; i < rounds; i++ {
		snakeLen = g.PlayerLen(player.ID)
		gameOver, _ := g.PlayRound()
		if gameOver || !g.Alive(player.ID) {
			r.Fitness = float64(i) / float64(rounds)
			break
		}
	}
	r.Fitness += float64(snakeLen)
	return r, err
}

type Evaluator struct {
}
