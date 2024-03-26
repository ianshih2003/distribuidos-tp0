package common

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// Agency entity
type Agency struct {
	client *Client
}

func NewAgency(client_config ClientConfig) *Agency {
	client := NewClient(client_config)

	agency := &Agency{
		client: client,
	}

	return agency
}

func (agency *Agency) Start() {
	file, err := os.Open(fmt.Sprintf("/dataset/agency-%s.csv", agency.client.config.ID))

	defer file.Close()

	defer agency.client.Shutdown()

	if err != nil {
		log.Errorf("action: abrir_archivo | result: fail | client_id %s | error %v", agency.client.config.ID, err)
		return
	}

	err = agency.client.createClientSocket()

	if err != nil {
		return
	}

	for {

		bets, err := readBets(file, agency.client.config.ID, agency.client.config.MaxBatchSize)

		if err != nil {
			log.Errorf("action: leer_apuestas | result: fail | client_id %s | error %v", agency.client.config.ID, err)
		}

		agency.SendBets(bets)
	}

}

func (agency *Agency) SendBets(bets []*Bet) {
	serialized := serializeMultiple(bets)

	agency.client.SendMessage(serialized)
}
