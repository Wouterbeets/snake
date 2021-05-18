package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
		runs   = flag.Int("runs", 1, "number of experiments to run")
		iter   = flag.Int("iterations", 1000, "number of iterations for experiment")
		cpath  = flag.String("config", "genn.json", "path to the configuration file")
		epath  = flag.String("efficacy", "xor-samples.txt", "path for efficacy sample file")
		aiOut  = flag.String("aiout", "ai.json", "path for ai")
		loadAi = flag.String("loadai", "", "ai in file")
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

	var netJson []byte
	var best []evo.Genome
	f := func(pop evo.Population) error {

		genomes := make([]evo.Genome, len(pop.Genomes))
		copy(genomes, pop.Genomes)

		// Sort so the best genome is at the end
		evo.SortBy(genomes, evo.BySolved, evo.ByFitness, evo.ByComplexity, evo.ByAge)

		// Output the best
		best = genomes[len(genomes)-10:]

		substrates := make([]evo.Substrate, 0, len(best))
		var sumScore float64
		for _, s := range best {
			substrates = append(substrates, s.Decoded)
			sumScore += s.Fitness
		}
		netJson, err = json.MarshalIndent(substrates, "", "\t")
		file, err := os.Create("end.json")
		if err != nil {
			panic(err.Error())
		}
		fmt.Fprintf(file, string(netJson))
		file.Close()
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
	exp.Searcher = ai.Trainer{}
	if *loadAi == "" {

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
			var highestFitness float64
			saveAIFunc := func(pop evo.Population) error {
				genomes := make([]evo.Genome, len(pop.Genomes))
				copy(genomes, pop.Genomes)

				// Sort so the best genome is at the end
				evo.SortBy(genomes, evo.BySolved, evo.ByFitness, evo.ByComplexity, evo.ByAge)

				var sumScore float64
				for _, g := range genomes {
					sumScore += g.Fitness
				}

				// Output the best
				best := genomes[len(genomes)-10:]
				substrates := make([]evo.Substrate, 0, len(best))
				for _, s := range best {
					substrates = append(substrates, s.Decoded)
				}
				roundBest := best[len(best)-1].Fitness
				fmt.Printf("gen %d \t sum  %.3f \t avg %.3f \t best %.3f \t alltime %.3f\n", pop.Generation, sumScore, sumScore/float64(len(genomes)), roundBest, highestFitness)
				if highestFitness < roundBest {
					highestFitness = roundBest
					netJson, err = json.MarshalIndent(substrates, "", "\t")
					file, err := os.Create(*aiOut)
					if err != nil {
						panic(err.Error())
					}
					fmt.Fprintf(file, string(netJson))
					file.Close()
				}

				return nil
			}
			exp.AddSubscription(evo.Subscription{Event: evo.Evaluated, Callback: saveAIFunc})

			// Stop the experiment if there is a solution
			ctx, fn, cb = evo.WithSolution(ctx)
			defer fn() // ensure the context cancels
			exp.AddSubscription(evo.Subscription{Event: evo.Evaluated, Callback: cb})
			// Execute the experiment
			if _, err = evo.Run(ctx, exp, ai.Evaluator{}); err != nil {
				log.Fatalf("%+v\n", err)
			}
		}
	} else {
		aiFromFile, err := os.Open(*loadAi)
		if err != nil {
			panic(err.Error())
		}
		netJson, err = ioutil.ReadAll(aiFromFile)
		if err != nil {
			panic(err.Error())
		}

	}

	var sub []evo.Substrate
	err = json.Unmarshal(netJson, &sub)
	if err != nil {
		panic(err.Error())
	}

	var nets []*ai.NetWrapper
	for _, s := range sub {
		net, err := exp.Translate(s)
		if err != nil {
			panic(err.Error())
		}
		nets = append(nets, &ai.NetWrapper{Ai: net})
	}

	framerate := 10 * time.Millisecond
	sc := term.Screen{Input: make(chan [][]rune), UserInput: make(chan rune)}
	players := []snake.Player{
		&snake.Human{Input: sc.UserInput, Framerate: framerate},
	}
	for _, n := range nets {
		players = append(players, n)
	}
	g, err := snake.NewGame(50, 50, players, 50)
	if err != nil {
		panic(err)
	}
	done := sc.Run(framerate)

	runes := map[int8]rune{
		-1: 'M',
		0:  ' ',
		1:  '█',
		2:  '2',
		3:  '3',
		4:  '4',
		5:  '5',
		6:  '6',
		7:  '7',
		8:  '8',
		9:  '9',
		10: 'a',
		11: 'b',
		12: 'c',
		13: 'd',
		14: 'e',
	}
	for i := 0; i < 10000000; i++ {
		gameOver, state := g.PlayRound()
		sc.Input <- stateToRune(state, runes)
		if gameOver {
			close(sc.Input)
			return
		}
	}
	<-done
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
