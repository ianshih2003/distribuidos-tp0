package common

import (
	"bytes"
	"fmt"
)

// Bet entity
type Bet struct {
	agency    string
	name      string
	last_name string
	document  string
	birthdate string
	number    string
}

func NewBet(agency string, name string, last_name string, document string, birth_date string, number string) *Bet {
	bet := &Bet{
		agency:    agency,
		name:      name,
		last_name: last_name,
		document:  document,
		birthdate: birth_date,
		number:    number,
	}

	return bet
}

func (b *Bet) serialize() []byte {
	return []byte(fmt.Sprintf("%s|%s|%s|%s|%s|%s", b.agency, b.name, b.last_name, b.document, b.birthdate, b.number))
}

func serializeMultiple(bets []*Bet) []byte {
	serialized_bets := make([][]byte, len(bets))
	for i, bet := range bets {
		if bet != nil {
			serialized_bets[i] = bet.serialize()
		}
	}

	return bytes.Join(serialized_bets, []byte(";"))
}
