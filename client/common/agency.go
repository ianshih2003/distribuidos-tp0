package common

import (
	log "github.com/sirupsen/logrus"
)

// Agency entity
type Agency struct {
	bet    *Bet
	client *Client
}

func NewAgency(client_config ClientConfig, name string, last_name string, document string, birth_date string, number string) *Agency {
	client := NewClient(client_config)

	bet := NewBet(client_config.ID, name, last_name, document, birth_date, number)

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

func (agency *Agency) Start() {
	agency.SendBet()
}
