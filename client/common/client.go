package common

import (
	"encoding/binary"
	"errors"
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
const SUCCESS_MSG = "suc"
const ERROR_MSG = "err"
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
	c.SendMessage([]byte(EXIT_MSG), false)
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
	if err := c.SendAny(message); err != nil {
		return err
	}

	if wait_for_confirm {
		return c.ReceiveConfirmMsg()
	}

	return nil
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
	return err
}

func (c *Client) ReceiveConfirmMsg() error {
	buf, err := c.SafeReceive(CONFIRM_MSG_LENGTH)

	response := string(buf)
	if response == SUCCESS_MSG {
		log.Infof("action: receive_confirm_message | result: success | client_id: %v",
			c.config.ID,
		)
	} else if response == ERROR_MSG {
		log.Errorf("action: receive_confirm_message | result: fail | client_id: %v",
			c.config.ID,
		)
	}

	return err
}

func (c *Client) TryReceiveAll(length int) (n int, res []byte, res_error error) {
	result := make([]byte, length)

	n, err := c.conn.Read(result)

	if err != nil {
		return 0, nil, err
	}

	if n == length {
		return n, result, nil
	}

	return n, result, errors.New("MISSING")
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

func (c *Client) Receive() (res []byte, res_error error) {
	rcv_length, err := c.SafeReceive(MAX_MSG_BYTES)

	c.SendAny([]byte(SUCCESS_MSG))

	msg_length := int(binary.LittleEndian.Uint32(rcv_length))

	if msg_length == 0 {
		return []byte{}, err
	}

	res, _ = c.SafeReceive(msg_length)

	return res, c.SendAny([]byte(SUCCESS_MSG))
}
