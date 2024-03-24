package common

import (
	log "github.com/sirupsen/logrus"
)

// Agency entity
type Agency struct {
	bet    *Bet
	client *Client
}

func NewAgency(bet *Bet, client_config ClientConfig) *Agency {
	client := NewClient(client_config)
	agency := &Agency{
		bet:    bet,
		client: client,
	}

	return agency
}

func (agency *Agency) SendBet() {
	bet := agency.bet
	err := agency.client.StartClient(bet.serialize())

	if err != nil {
		return
	}

	log.Infof("action: apuesta_enviada | result: success | dni: %s | numero: %s",
		bet.document,
		bet.number,
	)
}
