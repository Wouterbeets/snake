package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/klokare/evo"
	"github.com/klokare/evo/config"
	"github.com/klokare/evo/config/source"
	"github.com/klokare/evo/efficacy"
	"github.com/klokare/evo/neat"
	"github.com/wouterbeets/snake"
	"github.com/wouterbeets/snake/ai"
	"github.com/wouterbeets/term"
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
		if _, err = evo.Run(ctx, exp, ai.Evaluator{}); err != nil {
			log.Fatalf("%+v\n", err)
		}
	}
	fmt.Printf("best: %+v\n ", best)
	net, err := exp.Translate(best.Decoded)
	e := ai.Evaluator{}
	r, err := e.EvaluateNet(net)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(r)

	framerate := 20 * time.Millisecond
	sc := term.Screen{Input: make(chan [][]rune), UserInput: make(chan rune)}
	players := []snake.Player{
		&snake.Human{Input: sc.UserInput, Framerate: framerate},
		&ai.NetWrapper{Ai: net},
		&ai.NetWrapper{Ai: net},
		&ai.NetWrapper{Ai: net},
		&ai.NetWrapper{Ai: net},
		&ai.NetWrapper{Ai: net},
		&ai.NetWrapper{Ai: net},
		&snake.Random{},
		&snake.Random{},
		&snake.Random{},
	}
	g, err := snake.NewGame(20, 20, players)
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
	for i := 0; i < 1000; i++ {
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
