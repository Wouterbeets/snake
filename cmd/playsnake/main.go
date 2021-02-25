package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/klokare/evo"
	"github.com/klokare/evo/config"
	"github.com/klokare/evo/config/source"
	"github.com/klokare/evo/efficacy"
	"github.com/klokare/evo/neat"
	"github.com/wouterbeets/snake"
	"github.com/wouterbeets/term"
	"gonum.org/v1/gonum/mat"
)

func main() {

	var (
		runs  = flag.Int("runs", 1, "number of experiments to run")
		iter  = flag.Int("iterations", 100, "number of iterations for experiment")
		cpath = flag.String("config", "genn.json", "path to the configuration file")
		epath = flag.String("efficacy", "xor-samples.txt", "path for efficacy sample file")
	)
	flag.Parse()

	// Load the configuration
	src, err := source.NewJSONFromFile(*cpath)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	cfg := config.Configurer{Source: source.Multi([]config.Source{
		source.Flag{},        // Check flags  first
		source.Environment{}, // Then check environment variables
		src,                  // Lastly, consult the configuration file
	})}

	var best evo.Genome
	f := func(pop evo.Population) error {

		genomes := make([]evo.Genome, len(pop.Genomes))
		copy(genomes, pop.Genomes)

		// Sort so the best genome is at the end
		evo.SortBy(genomes, evo.BySolved, evo.ByFitness, evo.ByComplexity, evo.ByAge)
		for i := range genomes {
			fmt.Printf("%.2f\n", genomes[i].Fitness)
		}

		// Output the best
		best = genomes[len(genomes)-1]
		return nil
	}

	// Create a sample file if performing multiple runs
	var s *efficacy.Sampler
	if *runs > 1 {
		if s, err = efficacy.NewSampler(*epath); err != nil {
			log.Fatalf("%+v\n", err)
		}
		defer s.Close()
	}

	exp := neat.NewExperiment(cfg)
	for r := 0; r < *runs; r++ {
		if s == nil {
			exp.AddSubscription(evo.Subscription{Event: evo.Completed, Callback: f}) // Show summary upon completion
		} else {
			c0, c1 := s.Callbacks(r)
			exp.AddSubscription(evo.Subscription{Event: evo.Started, Callback: c0})   // Begin the efficacy sample
			exp.AddSubscription(evo.Subscription{Event: evo.Completed, Callback: c1}) // End the efficacy sample
		}

		// Run the experiment for a set number of iterations
		ctx, fn, cb := evo.WithIterations(context.Background(), *iter)
		defer fn() // ensure the context cancels
		exp.AddSubscription(evo.Subscription{Event: evo.Evaluated, Callback: cb})

		// Stop the experiment if there is a solution
		ctx, fn, cb = evo.WithSolution(ctx)
		defer fn() // ensure the context cancels
		exp.AddSubscription(evo.Subscription{Event: evo.Evaluated, Callback: cb})
		// Execute the experiment
		if _, err = evo.Run(ctx, exp, Evaluator{}); err != nil {
			log.Fatalf("%+v\n", err)
		}
	}
	fmt.Printf("best: %+v\n ", best)
	net, err := exp.Translate(best.Decoded)
	e := Evaluator{}
	r, err := e.EvaluateNet(net)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(r)

	debug = true
	framerate := 20 * time.Millisecond
	sc := term.Screen{Input: make(chan [][]rune), UserInput: make(chan rune)}
	players := []snake.Player{
		&snake.Human{Input: sc.UserInput, Framerate: framerate},
		&NetWrapper{Ai: net},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
	}
	g, err := snake.NewGame(40, 40, players)
	if err != nil {
		panic(err)
	}
	go sc.Run(framerate)

	runes := map[int8]rune{
		-1: 'M',
		0:  ' ',
		1:  '█',
		2:  '█',
	}
	for i := range players {
		runes[int8(i)+3] = '█'
	}
	for i := 0; i < 300; i++ {
		gameOver, state := g.PlayRound()
		sc.Input <- stateToRune(state, runes)
		if gameOver {
			return
		}
	}
}

func stateToRune(state snake.Board, runes map[int8]rune) (disp [][]rune) {
	disp = make([][]rune, len(state))
	for i := range disp {
		disp[i] = make([]rune, len(state[i]))
	}

	for y, row := range state {
		for x := range row {
			disp[y][x] = runes[state[y][x]]
		}
	}
	return disp
}

type Evaluator struct {
	ID snake.ID
}

func (e Evaluator) Evaluate(p evo.Phenome) (r evo.Result, err error) {

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
	if r.Fitness < 17.0 {
		return r, err
	}
	player := NetWrapper{Ai: p.Network}
	g, err := snake.NewGame(20, 20, []snake.Player{
		&player,
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
	})

	rounds := 100
	var snakeLen int
	for i := 0; i < rounds; i++ {
		//snakeLen = g.PlayerLen(player.ID)
		gameOver, _ := g.PlayRound()
		if gameOver || !g.Alive(player.ID) {
			r.Fitness = float64(i) / float64(rounds)
			break
		}
	}
	r.Fitness += float64(snakeLen)
	fmt.Printf("id: %d, fit: %.2f\n", p.ID, r.Fitness)
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

type NetWrapper struct {
	Ai evo.Network
	ID snake.ID
}

var debug bool

func (n *NetWrapper) Play(g *snake.Game) snake.Move {
	vis := g.Vision(n.ID)
	in := mat.NewDense(1, 3, []float64{float64(vis[0]), float64(vis[1]), float64(vis[2])})
	out, err := n.Ai.Activate(in)
	if err != nil {
		panic("error in ai")
		//		return snake.Move{Move: []float64{0, 0, 0}, ID: n.ID}
	}
	ret := []float64{out.At(0, 0), out.At(0, 1), out.At(0, 2)}
	if debug {
		for i := range ret {
			fmt.Printf("%d -> %.2f\n", vis[i], ret[i])
		}
		fmt.Println("")
	}
	return snake.Move{Move: ret, ID: n.ID}
}

func (n *NetWrapper) SetID(id snake.ID) {
	n.ID = id
}
