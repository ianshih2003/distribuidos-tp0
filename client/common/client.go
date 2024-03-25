package common

import (
	"encoding/binary"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

const CONFIRM_MSG_LENGTH = 3
const MAX_MSG_BYTES = 4

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
	MaxBatchSize  int
}

// Client Entity that encapsulates how
type Client struct {
	config     ClientConfig
	conn       net.Conn
	isFinished bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:     config,
		isFinished: false,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Fatalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func (c *Client) Shutdown() error {
	c.conn.Close()
	c.isFinished = true
	return nil
}

func (c *Client) SendMessageLength(message_length int) error {
	bs := make([]byte, MAX_MSG_BYTES)

	binary.LittleEndian.PutUint32(bs, uint32(message_length))

	return c.SendAny(bs)
}

func (c *Client) SendMessage(message []byte) error {

	err := c.SendMessageLength(len(message))

	if err != nil {
		return err
	}

	err = c.SendAny(message)

	if err != nil {
		return err
	}

	return err
}

func (c *Client) SendAny(message []byte) error {
	n := 0
	var err error

	message_length := len(message)

	for n != message_length {
		n, err = c.conn.Write(message)

		if err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}
	}

	err = c.ReceiveConfirmMsg()

	if err != nil {
		return err
	}

	return err
}

func (c *Client) ReceiveConfirmMsg() error {
	buf := make([]byte, CONFIRM_MSG_LENGTH)
	n, err := c.conn.Read(buf)

	if err != nil || n != CONFIRM_MSG_LENGTH {
		log.Errorf("action: receive_confirm_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	log.Errorf("action: receive_confirm_message | result: sucess | client_id: %v",
		c.config.ID,
	)

	return err
}
