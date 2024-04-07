package common

import (
	"encoding/binary"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const CONFIRM_MSG_LENGTH = 3
const MAX_MSG_BYTES = 4
const EXIT_MSG = "exit"

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

func InitializeSignalListener(client *Client) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGTERM)

	go func(client *Client) {
		sig := <-sigs
		logrus.Infof("action: received termination signal | result: in_progress | signal: %s", sig)
		err := client.Shutdown()

		if err != nil {
			logrus.Infof("action: received termination signal | result: error | signal: %s | error: %v", sig, err)
		}
		logrus.Infof("action: received termination signal | result: success | signal: %s", sig)
	}(client)
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:     config,
		isFinished: false,
	}

	InitializeSignalListener(client)

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
	c.SendMessage([]byte(EXIT_MSG))
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
	var err error

	total_bytes_sent := 0

	message_length := len(message)

	for total_bytes_sent < message_length {
		n, err := c.conn.Write(message[total_bytes_sent:])

		total_bytes_sent += n

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
	n, err := c.SafeReceive(CONFIRM_MSG_LENGTH)

	if err != nil || len(n) != CONFIRM_MSG_LENGTH {
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

func (c *Client) SafeReceive(length int) (res []byte, res_error error) {
	buf := make([]byte, length)
	bytes_read := 0
	result := make([]byte, length)

	var err error

	for bytes_read < length {
		n, err := c.conn.Read(buf)

		if err != nil {
			break
		} else if n == 0 {
			return result, net.ErrClosed
		}

		copy(result[:len(buf)], buf)

		bytes_read += n

		buf = make([]byte, length)

	}

	return result, err

}
