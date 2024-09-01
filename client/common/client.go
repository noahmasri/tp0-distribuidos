package common

import (
	"net"
	"time"
	"bytes"
	"os"
	"encoding/binary"
	"strconv"
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
	config	ClientConfig
	bet		Bet		
	conn	net.Conn
	end		bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, bet Bet) *Client {
	client := &Client{
		config: config,
		bet: bet,
		end: false,
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

func (c *Client) ShutdownGracefully(notifier chan os.Signal, done chan bool) {

	select{
	case <-notifier:
		// if it gets notified it should continue with this flow
		log.Infof("action: shutdown_gracefully | result: in_progress | client_id: %v | msg: received SIGTERM signal",
			c.config.ID,
		)
	case <-done:
		// gets done from client channel
		return
	}
	c.end = true
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

func (c *Client) SendBet() error{
	data := c.EncodeAgencyData() 
	totalWritten := 0
	var err error
	for totalWritten < len(data) {
		var written int
		written, err = c.conn.Write(data[totalWritten:])
		totalWritten += written
	}
	return err
}

func (c *Client) EncodeAgencyData() []byte {
	agency, err := strconv.ParseInt(c.config.ID, 10, 8)
	if err != nil {
		log.Fatalf("action: parse_bets | result: fail | error: %v | msg: agency number is invalid", err)
		return nil
	}

	var buffer bytes.Buffer
	bet_bytes := c.bet.EncodeToBytes()
	binary.Write(&buffer, binary.LittleEndian, uint8(agency))
	buffer.Write(bet_bytes)
	return buffer.Bytes()
}

func (c *Client) logStatus(status ResponseStatus){

	switch status {
	case OK:
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			c.bet.ID,
			c.bet.Number,
		)
	case ERR:
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | error: bet was not saved correctly",
			c.bet.ID,
			c.bet.Number,
		)
	case ABORT:
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | error: server aborted",
			c.bet.ID,
			c.bet.Number,
		)
	default:
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | error: server returned unknown state",
			c.bet.ID,
			c.bet.Number,
		)
	}
}

func (c *Client) SendErrorMessageAndExit(done chan bool, action string, err error){
	// only log error message if it wasnt because got an exception
	if !c.end {
		log.Errorf("action: %v | result: fail | client_id: %v | error: %v",
			action,
			c.config.ID,
			err,
		)
		c.conn.Close()
		done <- true
	}
}

func (c *Client) MakeBet(done chan bool) {
	e := c.createClientSocket()
	if e != nil {
		done <- true
		return
	}

	err := c.SendBet()
	if err != nil {
		c.SendErrorMessageAndExit(done, "send_message", err)
		return
	}

	buf := make([]byte, 1024)
	_, err = c.conn.Read(buf)
	if err != nil {
		c.SendErrorMessageAndExit(done, "receive_message", err)
		return
	}

	c.logStatus(ResponseStatus(buf[0]))

	c.conn.Close()
	c.conn = nil

	done <- true
}
