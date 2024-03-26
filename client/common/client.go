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
	c.SendMessage([]byte("exit"), false)
	c.conn.Close()
	c.isFinished = true
	return nil
}

func (c *Client) SendMessageLength(message_length int) error {
	bs := make([]byte, MAX_MSG_BYTES)

	binary.LittleEndian.PutUint32(bs, uint32(message_length))

	return c.SendAndWaitConfirm(bs, true)
}

func (c *Client) SendMessage(message []byte, wait_for_confirm bool) error {
	if err := c.SendMessageLength(len(message)); err != nil {
		return err
	}
	return c.SendAndWaitConfirm(message, wait_for_confirm)
}

func (c *Client) SendAndWaitConfirm(message []byte, wait_for_confirm bool) error {
	if err := c.SafeSend(message); err != nil {
		return err
	}

	if wait_for_confirm {
		return c.ReceiveConfirmMsg()
	}

	return nil
}

func (c *Client) SafeSend(message []byte) error {
	n := 0
	var err error

	message_length := len(message)

	for n < message_length {
		n, err = c.conn.Write(message)

		if err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}
	}
	return err
}

func (c *Client) ReceiveConfirmMsg() error {
	buf, err := c.SafeReceive(CONFIRM_MSG_LENGTH)

	response := string(buf)
	if response == "suc" {
		log.Infof("action: receive_confirm_message | result: success | client_id: %v",
			c.config.ID,
		)
	} else if response == "err" {
		log.Errorf("action: receive_confirm_message | result: fail | client_id: %v",
			c.config.ID,
		)
	}

	return err
}

func (c *Client) SafeReceive(length int) (res []byte, res_error error) {
	buf := make([]byte, length)

	_, err := c.conn.Read(buf)

	return buf, err

}

func (c *Client) Receive() (res []byte, res_error error) {
	rcv_length, err := c.SafeReceive(MAX_MSG_BYTES)

	c.SafeSend([]byte("suc"))

	msg_length := int(binary.LittleEndian.Uint32(rcv_length))

	if msg_length == 0 {
		return []byte{}, err
	}

	res, _ = c.SafeReceive(msg_length)

	return res, c.SafeSend([]byte("suc"))
}
