package poker

import (
	"math/rand"
	"time"
)

// SimulationResult holds the outcome of a Monte Carlo equity simulation.
type SimulationResult struct {
	HeroWins   int
	VillainWins int
	Ties       int
	TrialsRun  int
}

// SimulateEquity estimates the probability that hero's hand wins against
// `numOpponents` players, given optional community cards (0, 3, 4, or 5).
//
// heroHole: exactly 2 cards
// community: 0, 3, 4, or 5 cards
// numOpponents: number of other players (1+)
// trials: number of random simulations
//
// It uses simple goroutine-based parallelism to split work across CPU cores.
func SimulateEquity(heroHole []Card, community []Card, numOpponents, trials int) SimulationResult {
	if len(heroHole) != 2 {
		panic("heroHole must have length 2")
	}
	if len(community) != 0 && len(community) != 3 && len(community) != 4 && len(community) != 5 {
		panic("community must be 0, 3, 4, or 5 cards")
	}
	if numOpponents < 1 {
		panic("numOpponents must be >= 1")
	}
	if trials <= 0 {
		return SimulationResult{}
	}

	// Build deck without known cards.
	deck := FullDeck()
	used := make(map[string]bool)
	for _, c := range heroHole {
		used[c.Str] = true
	}
	for _, c := range community {
		used[c.Str] = true
	}
	filtered := make([]Card, 0, len(deck))
	for _, c := range deck {
		if !used[c.Str] {
			filtered = append(filtered, c)
		}
	}

	// Parallelism: number of worker goroutines.
	workers := 4
	if trials < workers {
		workers = trials
	}
	trialsPerWorker := trials / workers
	remaining := trials % workers

	results := make(chan SimulationResult, workers)

	for w := 0; w < workers; w++ {
		tw := trialsPerWorker
		if w == 0 {
			tw += remaining
		}
		go func(trialsForWorker int) {
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			local := SimulationResult{}

			for i := 0; i < trialsForWorker; i++ {
				heroWin, villainWin, tie := simulateOnce(rng, heroHole, community, filtered, numOpponents)
				if heroWin {
					local.HeroWins++
				} else if villainWin {
					local.VillainWins++
				} else if tie {
					local.Ties++
				}
				local.TrialsRun++
			}

			results <- local
		}(tw)
	}

	final := SimulationResult{}
	for i := 0; i < workers; i++ {
		r := <-results
		final.HeroWins += r.HeroWins
		final.VillainWins += r.VillainWins
		final.Ties += r.Ties
		final.TrialsRun += r.TrialsRun
	}

	return final
}

func simulateOnce(rng *rand.Rand, heroHole []Card, community []Card, deck []Card, numOpponents int) (heroWin, villainWin, tie bool) {
	// Make a copy of deck for shuffling.
	tmp := make([]Card, len(deck))
	copy(tmp, deck)
	rng.Shuffle(len(tmp), func(i, j int) {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	})

	// Determine how many more community cards we need to draw.
	toDraw := 5 - len(community)
	drawIdx := 0

	simCommunity := make([]Card, len(community))
	copy(simCommunity, community)
	for i := 0; i < toDraw; i++ {
		simCommunity = append(simCommunity, tmp[drawIdx])
		drawIdx++
	}

	// Hero 7-card hand.
	heroSeven := append([]Card{}, heroHole...)
	heroSeven = append(heroSeven, simCommunity...)
	heroBest := EvaluateBestHand(heroSeven)

	// Opponents.
	villainBetter := false
	equalCount := 0

	for opp := 0; opp < numOpponents; opp++ {
		if drawIdx+2 > len(tmp) {
			// Defensive; should not happen if deck is sized correctly.
			break
		}
		oppHole := []Card{tmp[drawIdx], tmp[drawIdx+1]}
		drawIdx += 2

		oppSeven := append([]Card{}, oppHole...)
		oppSeven = append(oppSeven, simCommunity...)
		oppBest := EvaluateBestHand(oppSeven)

		cmp := CompareHandValues(oppBest, heroBest)
		if cmp > 0 {
			villainBetter = true
		} else if cmp == 0 {
			equalCount++
		}
	}

	if villainBetter {
		return false, true, false
	}

	if equalCount > 0 {
		return false, false, true
	}

	return true, false, false
}

