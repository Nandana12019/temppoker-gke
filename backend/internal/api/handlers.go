package api

import (
	"encoding/json"
	"net/http"

	"github.com/example/texas-holdem-backend/internal/poker"
)

type evaluateRequest struct {
	Hole      []string `json:"hole"`      // exactly 2 cards
	Community []string `json:"community"` // 0-5 cards
}

type evaluateResponse struct {
	Category string      `json:"category"`
	Kickers  []string    `json:"kickers"`
	Value    poker.HandValue `json:"-"`
}

type winnerRequest struct {
	Player1Hole []string `json:"player1Hole"`
	Player2Hole []string `json:"player2Hole"`
	Community   []string `json:"community"`
}

type winnerResponse struct {
	Winner string `json:"winner"` // "player1", "player2", or "tie"
}

type simulateRequest struct {
	Hole         []string `json:"hole"`         // hero hole (2)
	Community    []string `json:"community"`    // 0, 3, 4, 5
	NumOpponents int      `json:"numOpponents"` // >= 1
	Trials       int      `json:"trials"`       // e.g. 5000, 10000
}

type simulateResponse struct {
	HeroWinPct    float64 `json:"heroWinPct"`
	VillainWinPct float64 `json:"villainWinPct"`
	TiePct        float64 `json:"tiePct"`
	TrialsRun     int     `json:"trialsRun"`
}

// RegisterRoutes attaches the REST endpoints to the given mux.
// It also enables CORS so the Flutter web app can call the API.
func RegisterRoutes(mux *http.ServeMux) {
	// Simple CORS wrapper for all API routes.
	withCORS := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			h(w, r)
		}
	}

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/api/evaluate", withCORS(handleEvaluate))
	mux.HandleFunc("/api/winner", withCORS(handleWinner))
	mux.HandleFunc("/api/simulate", withCORS(handleSimulate))
}

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req evaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.Hole) != 2 || len(req.Community) < 0 || len(req.Community) > 5 {
		http.Error(w, "invalid card counts", http.StatusBadRequest)
		return
	}

	var cards []poker.Card
	for _, s := range append(req.Hole, req.Community...) {
		c, err := poker.ParseCard(s)
		if err != nil {
			http.Error(w, "invalid card: "+err.Error(), http.StatusBadRequest)
			return
		}
		cards = append(cards, c)
	}
	if len(cards) != 2+len(req.Community) {
		http.Error(w, "invalid card count after parse", http.StatusBadRequest)
		return
	}

	// Pad community to 5 with dummy? No: evaluation is defined for 7 cards.
	// Here we expect exactly 7 total (2 + 5).
	if len(cards) != 7 {
		http.Error(w, "must supply exactly 2 hole cards and 5 community cards", http.StatusBadRequest)
		return
	}

	hv := poker.EvaluateBestHand(cards)

	resp := evaluateResponse{
		Category: categoryToString(hv.Category),
		Kickers:  ranksToStrings(hv.Kickers),
	}

	writeJSON(w, resp)
}

func handleWinner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req winnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.Player1Hole) != 2 || len(req.Player2Hole) != 2 || len(req.Community) != 5 {
		http.Error(w, "require 2 hole cards for each player and 5 community cards", http.StatusBadRequest)
		return
	}

	parseCards := func(strs []string) ([]poker.Card, error) {
		cs := make([]poker.Card, 0, len(strs))
		for _, s := range strs {
			c, err := poker.ParseCard(s)
			if err != nil {
				return nil, err
			}
			cs = append(cs, c)
		}
		return cs, nil
	}

	p1Hole, err := parseCards(req.Player1Hole)
	if err != nil {
		http.Error(w, "invalid player1 hole: "+err.Error(), http.StatusBadRequest)
		return
	}
	p2Hole, err := parseCards(req.Player2Hole)
	if err != nil {
		http.Error(w, "invalid player2 hole: "+err.Error(), http.StatusBadRequest)
		return
	}
	community, err := parseCards(req.Community)
	if err != nil {
		http.Error(w, "invalid community: "+err.Error(), http.StatusBadRequest)
		return
	}

	p1Seven := append([]poker.Card{}, p1Hole...)
	p1Seven = append(p1Seven, community...)
	p2Seven := append([]poker.Card{}, p2Hole...)
	p2Seven = append(p2Seven, community...)

	p1Best := poker.EvaluateBestHand(p1Seven)
	p2Best := poker.EvaluateBestHand(p2Seven)

	cmp := poker.CompareHandValues(p1Best, p2Best)
	var winner string
	switch {
	case cmp > 0:
		winner = "player1"
	case cmp < 0:
		winner = "player2"
	default:
		winner = "tie"
	}

	writeJSON(w, winnerResponse{Winner: winner})
}

func handleSimulate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req simulateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.Hole) != 2 {
		http.Error(w, "hero hole must be 2 cards", http.StatusBadRequest)
		return
	}
	if !(len(req.Community) == 0 || len(req.Community) == 3 || len(req.Community) == 4 || len(req.Community) == 5) {
		http.Error(w, "community must be 0, 3, 4, or 5 cards", http.StatusBadRequest)
		return
	}
	if req.NumOpponents < 1 {
		http.Error(w, "numOpponents must be >= 1", http.StatusBadRequest)
		return
	}
	if req.Trials <= 0 {
		http.Error(w, "trials must be > 0", http.StatusBadRequest)
		return
	}

	parseCards := func(strs []string) ([]poker.Card, error) {
		cs := make([]poker.Card, 0, len(strs))
		for _, s := range strs {
			c, err := poker.ParseCard(s)
			if err != nil {
				return nil, err
			}
			cs = append(cs, c)
		}
		return cs, nil
	}

	hole, err := parseCards(req.Hole)
	if err != nil {
		http.Error(w, "invalid hero hole: "+err.Error(), http.StatusBadRequest)
		return
	}
	community, err := parseCards(req.Community)
	if err != nil {
		http.Error(w, "invalid community: "+err.Error(), http.StatusBadRequest)
		return
	}

	res := poker.SimulateEquity(hole, community, req.NumOpponents, req.Trials)

	total := float64(res.TrialsRun)
	resp := simulateResponse{
		HeroWinPct:    float64(res.HeroWins) / total * 100.0,
		VillainWinPct: float64(res.VillainWins) / total * 100.0,
		TiePct:        float64(res.Ties) / total * 100.0,
		TrialsRun:     res.TrialsRun,
	}

	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func categoryToString(cat int) string {
	switch cat {
	case poker.StraightFlush:
		return "Straight Flush"
	case poker.FourOfAKind:
		return "Four of a Kind"
	case poker.FullHouse:
		return "Full House"
	case poker.Flush:
		return "Flush"
	case poker.Straight:
		return "Straight"
	case poker.ThreeOfAKind:
		return "Three of a Kind"
	case poker.TwoPair:
		return "Two Pair"
	case poker.OnePair:
		return "One Pair"
	default:
		return "High Card"
	}
}

func ranksToStrings(rs []poker.Rank) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		switch r {
		case poker.Two:
			out[i] = "2"
		case poker.Three:
			out[i] = "3"
		case poker.Four:
			out[i] = "4"
		case poker.Five:
			out[i] = "5"
		case poker.Six:
			out[i] = "6"
		case poker.Seven:
			out[i] = "7"
		case poker.Eight:
			out[i] = "8"
		case poker.Nine:
			out[i] = "9"
		case poker.Ten:
			out[i] = "T"
		case poker.Jack:
			out[i] = "J"
		case poker.Queen:
			out[i] = "Q"
		case poker.King:
			out[i] = "K"
		case poker.Ace:
			out[i] = "A"
		default:
			out[i] = "?"
		}
	}
	return out
}

