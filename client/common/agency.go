package common

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Agency entity
type Agency struct {
	bet    *Bet
	client *Client
}

func NewAgency(client_config ClientConfig, v *viper.Viper) *Agency {
	client := NewClient(client_config)

	bet := NewBet(v.GetString("id"), v.GetString("nombre"), v.GetString("apellido"), v.GetString("documento"), v.GetString("nacimiento"), v.GetString("numero"))

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
