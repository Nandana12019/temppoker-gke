package poker

import (
	"fmt"
)

// Card is represented as a 2-character string, e.g. "HA", "S7", "CT".
// Suits: H (hearts), D (diamonds), C (clubs), S (spades)
// Ranks: 2-9, T (10), J, Q, K, A

type Suit int
type Rank int

const (
	Hearts Suit = iota
	Diamonds
	Clubs
	Spades
)

const (
	Two Rank = iota + 2
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

type Card struct {
	Suit Suit
	Rank Rank
	Str  string // original string ("HA", etc.) for convenience
}

// ParseCard converts a 2-character string like "HA" into a Card.
func ParseCard(s string) (Card, error) {
	if len(s) != 2 {
		return Card{}, fmt.Errorf("invalid card format: %s", s)
	}

	var r Rank
	switch s[1] {
	case '2':
		r = Two
	case '3':
		r = Three
	case '4':
		r = Four
	case '5':
		r = Five
	case '6':
		r = Six
	case '7':
		r = Seven
	case '8':
		r = Eight
	case '9':
		r = Nine
	case 'T':
		r = Ten
	case 'J':
		r = Jack
	case 'Q':
		r = Queen
	case 'K':
		r = King
	case 'A':
		r = Ace
	default:
		return Card{}, fmt.Errorf("invalid rank: %c", s[1])
	}

	var suit Suit
	switch s[0] {
	case 'H':
		suit = Hearts
	case 'D':
		suit = Diamonds
	case 'C':
		suit = Clubs
	case 'S':
		suit = Spades
	default:
		return Card{}, fmt.Errorf("invalid suit: %c", s[0])
	}

	return Card{Suit: suit, Rank: r, Str: s}, nil
}

// FullDeck returns all 52 cards.
func FullDeck() []Card {
	suits := []Suit{Hearts, Diamonds, Clubs, Spades}
	ranks := []Rank{Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King, Ace}

	deck := make([]Card, 0, 52)
	for _, s := range suits {
		for _, r := range ranks {
			deck = append(deck, Card{
				Suit: s,
				Rank: r,
				Str:  formatCard(s, r),
			})
		}
	}
	return deck
}

func formatCard(s Suit, r Rank) string {
	var suitChar byte
	switch s {
	case Hearts:
		suitChar = 'H'
	case Diamonds:
		suitChar = 'D'
	case Clubs:
		suitChar = 'C'
	case Spades:
		suitChar = 'S'
	}

	var rankChar byte
	switch r {
	case Two:
		rankChar = '2'
	case Three:
		rankChar = '3'
	case Four:
		rankChar = '4'
	case Five:
		rankChar = '5'
	case Six:
		rankChar = '6'
	case Seven:
		rankChar = '7'
	case Eight:
		rankChar = '8'
	case Nine:
		rankChar = '9'
	case Ten:
		rankChar = 'T'
	case Jack:
		rankChar = 'J'
	case Queen:
		rankChar = 'Q'
	case King:
		rankChar = 'K'
	case Ace:
		rankChar = 'A'
	}

	return string([]byte{suitChar, rankChar})
}

