package common

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const WAITING_MESSAGE = "waiting"
const WINNERS_SEPARATOR = ","
const CSV_SEPARATOR = ","

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

func (agency *Agency) StartBetSendingProcess() error {
	file, err := os.Open(fmt.Sprintf("/dataset/agency-%s.csv", agency.client.config.ID))

	defer file.Close()

	defer agency.client.Shutdown()

	if err != nil {
		log.Errorf("action: abrir_archivo | result: fail | client_id %s | error %v", agency.client.config.ID, err)
		return err
	}

	err = agency.client.createClientSocket()

	if err != nil {
		return err
	}

	for {

		bets, err := readBets(file, agency.client.config.ID, agency.client.config.MaxBatchSize)

		if fmt.Sprint(err) == "EOF" {
			return nil
		}

		if err != nil {
			log.Errorf("action: leer_apuestas | result: fail | client_id %s | error %v", agency.client.config.ID, err)
			return err
		}

		agency.SendBets(bets)
	}
}

func (agency *Agency) Start() {
	if err := agency.StartBetSendingProcess(); err != nil {
		logrus.Infof("action: apuestas_enviadas | result: fail | client_id: %s | err: %v", agency.client.config.ID, err)

		return
	}

	logrus.Infof("action: apuestas_enviadas | result: success | client_id: %s", agency.client.config.ID)

	if err := agency.AskForWinners(); err != nil {
		logrus.Infof("action: pedir_ganadores | result: fail | client_id: %s", agency.client.config.ID)

		return
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

		res, err := agency.client.ReceiveAndWaitConfirm()

		if err != nil {
			break
		}

		if checkWinnersAnnouncementMsg(res) {
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
	return strings.Split(string(bytes), WINNERS_SEPARATOR)
}

func (agency *Agency) AnnounceWinners(winners []string) {
	length := len(winners)
	if len(winners[0]) == 0 {
		length = 0
	}
	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", length)
}

func checkWinnersAnnouncementMsg(message []byte) bool {
	return message != nil && string(message) != WAITING_MESSAGE
}
