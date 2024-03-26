package common

import (
	"fmt"
	"os"
	"strings"
	"time"

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
			return
		}

		agency.SendBets(bets)
	}
}

func (agency *Agency) SendBets(bets []*Bet) {
	serialized := serializeMultiple(bets)

	agency.client.SendMessage(serialized, true)
}

func (agency *Agency) AskForWinners() error {

	message := fmt.Sprintf("winners,%s", agency.client.config.ID)

	var err error

	for {
		agency.client.createClientSocket()

		agency.client.SendMessage([]byte(message), false)

		res, err := agency.client.Receive()

		if err != nil {
			break
		}

		if res != nil && string(res) != "waiting" {
			winners := parseWinners(res)
			agency.AnnounceWinners(winners)
			break
		}

		agency.client.Shutdown()

		time.Sleep(time.Duration(1))
	}

	agency.client.Shutdown()

	return err
}

func parseWinners(bytes []byte) []string {
	log.Infof("%d", len(strings.Split(string(bytes), ",")))
	return strings.Split(string(bytes), ",")
}

func (agency *Agency) AnnounceWinners(winners []string) {
	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", len(winners))
}
