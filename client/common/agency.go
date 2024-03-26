package common

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const MAX_BETS = 32

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
		// Para tener en cuenta que se agrega la agencia
		buffer := make([]byte, agency.client.config.MaxBatchSize-1024)

		n, err := file.Read(buffer)

		if n == 0 {
			return
		}

		if err != nil {
			log.Errorf("action: leer_archivo | result: fail | client_id %s | error %v", agency.client.config.ID, err)
			return
		}

		bets := readBets(file, buffer, agency.client.config.ID)

		buffer = nil

		agency.SendBets(bets)
	}

}

func readBets(file *os.File, buffer []byte, id string) []*Bet {
	unparsed_bets := string(bytes.Trim(bytes.Trim(buffer, "\r"), "\x00"))

	bets_str := strings.Split(unparsed_bets, "\n")

	if !strings.HasSuffix(unparsed_bets, "\n") && !strings.HasSuffix(unparsed_bets, "\r") {
		bets_str = bets_str[:len(bets_str)-1]

		offset := int64(len(unparsed_bets) - strings.LastIndex(unparsed_bets, "\n") - 1)

		file.Seek(-offset, io.SeekCurrent)
	}

	return parseBets(bets_str, id)
}

func parseBets(bets_str []string, id string) []*Bet {
	bets := make([]*Bet, len(bets_str))
	for i, bet_str := range bets_str {
		fields := strings.Split(bet_str, ",")

		if len(fields) != 5 {
			continue
		}

		bets[i] = NewBet(id, fields[0], fields[1], fields[2], fields[3], fields[4])
	}

	return bets
}

func (agency *Agency) SendBets(bets []*Bet) {
	serialized := serialize_multiple(bets)

	agency.client.SendMessage(serialized)
}
