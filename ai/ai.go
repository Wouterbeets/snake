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
	inf := make([]float64, len(vis))
	for i := range vis {
		inf[i] = float64(vis[i])
	}
	in := mat.NewDense(1, len(vis), []float64{float64(vis[0]), float64(vis[1]), float64(vis[2])})
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
		in *mat.Dense
	}

	evals := []eval{
		{
			in: mat.NewDense(1, 3, []float64{
				0, 0, 0,
			}),
		},
		{
			in: mat.NewDense(1, 3, []float64{
				0, 1, 0,
			}),
		},
		{
			in: mat.NewDense(1, 3, []float64{
				0, 1, 1,
			}),
		},
		{
			in: mat.NewDense(1, 3, []float64{
				0, 0, 1,
			}),
		},
		{
			in: mat.NewDense(1, 3, []float64{
				1, 0, 0,
			}),
		},
		{
			in: mat.NewDense(1, 3, []float64{
				1, 1, 0,
			}),
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
				r.Fitness += math.Abs(evals[e].in.At(i, j) - m.At(i, j))
			}
		}
	}
	if r.Fitness < 17.9 {
		return r, err
	}
	player := NetWrapper{Ai: p.Network}
	g, err := snake.NewGame(20, 20, []snake.Player{
		&player,
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

func (e Evaluator) EvaluateNet(n evo.Network) (r evo.Result, err error) {
	type eval struct {
		in *mat.Dense
	}

	evals := []eval{
		{
			in: mat.NewDense(1, 3, []float64{
				0, 0, 0,
			}),
		},
		{
			in: mat.NewDense(1, 3, []float64{
				0, 1, 0,
			}),
		},
		{
			in: mat.NewDense(1, 3, []float64{
				0, 1, 1,
			}),
		},
		{
			in: mat.NewDense(1, 3, []float64{
				0, 0, 1,
			}),
		},
		{
			in: mat.NewDense(1, 3, []float64{
				1, 0, 0,
			}),
		},
		{
			in: mat.NewDense(1, 3, []float64{
				1, 1, 0,
			}),
		},
	}

	for e := range evals {
		m, err := n.Activate(evals[e].in)

		if err != nil {
			log.Fatal(err)
		}
		row, col := m.Dims()
		for i := 0; i < row; i++ {
			for j := 0; j < col; j++ {
				fmt.Printf("%.2f\t -> %.2f\n", evals[e].in.At(i, j), m.At(i, j))
				r.Fitness += math.Abs(evals[e].in.At(i, j) - m.At(i, j))
			}
		}
		fmt.Println("")
	}
	r.Solved = r.Fitness > float64(len(evals))*3.0
	return r, err
}

type Evaluator struct {
}
