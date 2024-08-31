package common

import (
	"net"
	"time"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) ShutdownGracefully() {
	log.Infof("action: shutdown_gracefully | result: in_progress | client_id: %v | msg: received SIGTERM signal",
			c.config.ID,
		)

	if c.conn != nil {
        c.conn.Close()
		log.Infof("action: close_connection | result: success | client_id: %v",
			c.config.ID,
		)
    }

	log.Infof("action: shutdown_gracefully | result: success | client_id: %v",
		c.config.ID,
	)
}

func (c *Client) SendBet(bet Bet){
	data := EncodeAgencyData(c.config.ID, bet) 

	totalWritten := 0
	for totalWritten < len(data) {
		written, _ := c.conn.Write(data[totalWritten:])
		totalWritten += written
	}
}

func logStatus(status ResponseStatus, bet Bet){

	switch status {
	case OK:
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			bet.ID,
			bet.Number,
		)
	case ERR:
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | error: bet was not saved correctly",
			bet.ID,
			bet.Number,
		)
	case ABORT:
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | error: server aborted",
			bet.ID,
			bet.Number,
		)
	default:
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | error: server returned unknown state",
			bet.ID,
			bet.Number,
		)
	}
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(done chan bool) {
	bet := Bet{
		Name:    "Noah",
		Surname: "Masri",
		ID:      43724680,
		Date:    "2024-08-29",
		Number:  4206,
	}

	e := c.createClientSocket()
	if e != nil {
		done <- true
		return
	}

	c.SendBet(bet)

	buf := make([]byte, 1024)
	_, err := c.conn.Read(buf)
	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		done <- true
		return
	}

	logStatus(ResponseStatus(buf[0]), bet)

	c.conn.Close()
	c.conn = nil


	// Wait a time before finishing
	time.Sleep(c.config.LoopPeriod)
	done <- true
}
