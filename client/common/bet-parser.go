package common

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
)

func readBets(file *os.File, id string, batch_size int) ([]*Bet, error) {
	buffer := make([]byte, batch_size)

	n, err := file.Read(buffer)

	if n == 0 {
		return nil, errors.New("EOF")
	}

	if err != nil {
		return nil, err
	}

	unparsed_bets := string(bytes.Trim(bytes.Trim(buffer, "\r"), "\x00"))

	bets_str := strings.Split(unparsed_bets, "\n")

	if !strings.HasSuffix(unparsed_bets, "\n") && !strings.HasSuffix(unparsed_bets, "\r") {
		bets_str = bets_str[:len(bets_str)-1]

		offset := int64(len(unparsed_bets) - strings.LastIndex(unparsed_bets, "\n") - 1)

		file.Seek(-offset, io.SeekCurrent)
	}

	return parseBets(bets_str, id), nil
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
