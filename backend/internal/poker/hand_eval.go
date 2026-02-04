package poker

import (
	"sort"
)

// HandRank category values (higher is better).
const (
	HighCard = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
)

// HandValue is a comparable representation of a 5-card hand.
// Category is the hand type (StraightFlush, FourOfAKind, etc.).
// Kickers is a slice of ranks, already ordered by importance.
// Comparing two HandValues lexicographically (category, then kickers)
// is enough to determine which hand is better.
type HandValue struct {
	Category int
	Kickers  []Rank
}

// EvaluateBestHand takes exactly 7 cards (2 hole + 5 community)
// and returns the best 5-card hand value.
func EvaluateBestHand(cards []Card) HandValue {
	if len(cards) != 7 {
		panic("EvaluateBestHand requires exactly 7 cards")
	}

	best := HandValue{Category: HighCard, Kickers: []Rank{Two}} // minimal

	// There are exactly C(7,5) = 21 5-card combinations.
	indexes := []int{0, 1, 2, 3, 4}

	next := func() bool {
		// Generate next combination in lexicographic order.
		n := 7
		k := 5
		for i := k - 1; i >= 0; i-- {
			if indexes[i] != i+n-k {
				indexes[i]++
				for j := i + 1; j < k; j++ {
					indexes[j] = indexes[j-1] + 1
				}
				return true
			}
		}
		return false
	}

	evalCombo := func() HandValue {
		hand := []Card{
			cards[indexes[0]],
			cards[indexes[1]],
			cards[indexes[2]],
			cards[indexes[3]],
			cards[indexes[4]],
		}
		return evaluate5(hand)
	}

	best = evalCombo()
	for next() {
		hv := evalCombo()
		if CompareHandValues(hv, best) > 0 {
			best = hv
		}
	}

	return best
}

// evaluate5 evaluates exactly 5 cards and returns their HandValue.
func evaluate5(cards []Card) HandValue {
	// Sort by rank descending
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Rank > cards[j].Rank
	})

	isFlush, flushSuit := detectFlush(cards)
	isStr, topStr := detectStraight(cards)

	if isFlush {
		flushCards := filterBySuit(cards, flushSuit)
		// For 5-card input, flushCards will be length 5.
		isStrFlush, topStrFlush := detectStraight(flushCards)
		if isStrFlush {
			return HandValue{
				Category: StraightFlush,
				Kickers:  []Rank{topStrFlush},
			}
		}
		return HandValue{
			Category: Flush,
			Kickers:  ranksOnly(flushCards),
		}
	}

	if isStr {
		return HandValue{
			Category: Straight,
			Kickers:  []Rank{topStr},
		}
	}

	// Count occurrences of each rank.
	counts := make(map[Rank]int)
	for _, c := range cards {
		counts[c.Rank]++
	}

	type pair struct {
		r Rank
		c int
	}
	ps := make([]pair, 0, len(counts))
	for r, c := range counts {
		ps = append(ps, pair{r: r, c: c})
	}

	// Sort by count descending, then rank descending.
	sort.Slice(ps, func(i, j int) bool {
		if ps[i].c == ps[j].c {
			return ps[i].r > ps[j].r
		}
		return ps[i].c > ps[j].c
	})

	if ps[0].c == 4 {
		// Four of a kind: [quad_rank, kicker]
		kicker := highestDifferentRank(cards, ps[0].r)
		return HandValue{
			Category: FourOfAKind,
			Kickers:  []Rank{ps[0].r, kicker},
		}
	}

	if ps[0].c == 3 && len(ps) > 1 && ps[1].c >= 2 {
		// Full house: [trip_rank, pair_rank]
		return HandValue{
			Category: FullHouse,
			Kickers:  []Rank{ps[0].r, ps[1].r},
		}
	}

	if ps[0].c == 3 {
		// Trips: [trip_rank, kicker1, kicker2]
		k1, k2 := topKickersExcluding(cards, []Rank{ps[0].r}, 2)
		return HandValue{
			Category: ThreeOfAKind,
			Kickers:  []Rank{ps[0].r, k1, k2},
		}
	}

	if ps[0].c == 2 && len(ps) > 1 && ps[1].c == 2 {
		// Two pair: [high_pair, low_pair, kicker]
		highPair := ps[0].r
		lowPair := ps[1].r
		if lowPair > highPair {
			highPair, lowPair = lowPair, highPair
		}
		kicker := highestDifferentRank(cards, highPair, lowPair)
		return HandValue{
			Category: TwoPair,
			Kickers:  []Rank{highPair, lowPair, kicker},
		}
	}

	if ps[0].c == 2 {
		// One pair: [pair_rank, kicker1, kicker2, kicker3]
		k1, k2, k3 := topThreeKickersExcluding(cards, []Rank{ps[0].r})
		return HandValue{
			Category: OnePair,
			Kickers:  []Rank{ps[0].r, k1, k2, k3},
		}
	}

	// High card: top 5 ranks
	return HandValue{
		Category: HighCard,
		Kickers:  ranksOnly(cards),
	}
}

// CompareHandValues returns 1 if a > b, -1 if a < b, 0 if equal.
func CompareHandValues(a, b HandValue) int {
	if a.Category != b.Category {
		if a.Category > b.Category {
			return 1
		}
		return -1
	}
	// Compare kickers lexicographically.
	for i := 0; i < len(a.Kickers) && i < len(b.Kickers); i++ {
		if a.Kickers[i] > b.Kickers[i] {
			return 1
		}
		if a.Kickers[i] < b.Kickers[i] {
			return -1
		}
	}
	if len(a.Kickers) == len(b.Kickers) {
		return 0
	}
	if len(a.Kickers) > len(b.Kickers) {
		return 1
	}
	return -1
}

func detectFlush(cards []Card) (bool, Suit) {
	count := make(map[Suit]int)
	for _, c := range cards {
		count[c.Suit]++
	}
	for s, c := range count {
		if c >= 5 {
			return true, s
		}
	}
	return false, Hearts
}

func filterBySuit(cards []Card, s Suit) []Card {
	var res []Card
	for _, c := range cards {
		if c.Suit == s {
			res = append(res, c)
		}
	}
	return res
}

// detectStraight returns whether there is a straight and its top rank.
// It supports wheel straights (A-2-3-4-5).
func detectStraight(cards []Card) (bool, Rank) {
	seen := make(map[Rank]bool)
	for _, c := range cards {
		seen[c.Rank] = true
	}

	// Handle wheel straight: A-2-3-4-5
	if seen[Ace] && seen[Two] && seen[Three] && seen[Four] && seen[Five] {
		return true, Five
	}

	// Check sequences from Ace down to Five.
	for start := Ace; start >= Five; start-- {
		ok := true
		for offset := 0; offset < 5; offset++ {
			if !seen[start-Rank(offset)] {
				ok = false
				break
			}
		}
		if ok {
			return true, start
		}
	}
	return false, Two
}

func ranksOnly(cards []Card) []Rank {
	r := make([]Rank, len(cards))
	for i, c := range cards {
		r[i] = c.Rank
	}
	return r
}

func highestDifferentRank(cards []Card, exclude ...Rank) Rank {
	ex := make(map[Rank]bool)
	for _, r := range exclude {
		ex[r] = true
	}
	for _, c := range cards {
		if !ex[c.Rank] {
			return c.Rank
		}
	}
	return Two
}

func topKickersExcluding(cards []Card, exclude []Rank, n int) (Rank, Rank) {
	ex := make(map[Rank]bool)
	for _, r := range exclude {
		ex[r] = true
	}
	var res []Rank
	for _, c := range cards {
		if !ex[c.Rank] {
			res = append(res, c.Rank)
		}
		if len(res) == n {
			break
		}
	}
	// Safe because we only call with n=2 and there are enough cards.
	return res[0], res[1]
}

func topThreeKickersExcluding(cards []Card, exclude []Rank) (Rank, Rank, Rank) {
	ex := make(map[Rank]bool)
	for _, r := range exclude {
		ex[r] = true
	}
	var res []Rank
	for _, c := range cards {
		if !ex[c.Rank] {
			res = append(res, c.Rank)
		}
		if len(res) == 3 {
			break
		}
	}
	return res[0], res[1], res[2]
}

