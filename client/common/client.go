package common

import (
	"net"
	"time"
	"bytes"
	"errors"
	"os"
	"encoding/binary"
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
	config		ClientConfig
	betGetter	BetGetter
	agency		uint8	
	conn		net.Conn
	end			bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, betGetter BetGetter, agency uint8) *Client {
	client := &Client{
		config: config,
		betGetter: betGetter,
		agency: agency,
		end: false,
	}
	return client
}

func (c *Client) Destroy(){
	if !c.end{
		c.end = true
		if c.conn != nil {
			c.conn.Close()
		}
		log.Infof("action: close_connection | result: success | client_id: %v",
			c.config.ID,
		)
		c.betGetter.Destroy()
	}
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
		// gets signal that done channel was shutdown
		return
	}
	done <- true

	c.Destroy()
	log.Infof("action: shutdown_gracefully | result: success | client_id: %v",
		c.config.ID,
	)
}

// extra goes in case you want to add more information that what is provided by default
func (c *Client) logSendBetStatus(status ResponseStatus, action string, extra string){
	errMsg := status.GetStatusProperties()
	if errMsg != "" {
		log.Infof("action: %v | result: fail | cantidad: %v | error: %v",
			action,
			c.betGetter.lastBatchSize,
			errMsg,
		)
	}
	log.Infof("action: %v | result: success | cantidad: %v",
		action,
		c.betGetter.lastBatchSize,
	)
}

func (c *Client) SendErrorMessageAndExit(action string, err error) error {
	// only log error message if it wasnt because got an exception
	if !c.end {
		log.Errorf("action: %v | result: fail | client_id: %v | error: %v",
			action,
			c.config.ID,
			err,
		)
		if c.conn != nil {
			c.conn.Close()
		}
	}
	return err
}

func (c *Client) SendAll(data []byte) error{
	totalWritten := 0
	for totalWritten < len(data) {
		written, err := c.conn.Write(data[totalWritten:])
		if err != nil {
			return err
		}
		totalWritten += written
	}
	return nil
}

func (c *Client) SendBatch(batch []Bet) error{
	var buffer bytes.Buffer
	
	binary.Write(&buffer, binary.LittleEndian, c.agency)
	binary.Write(&buffer, binary.LittleEndian, BET)
	binary.Write(&buffer, binary.LittleEndian, uint8(len(batch)))
	for _, bet := range batch{
		bet_bytes := bet.EncodeToBytes()
		buffer.Write(bet_bytes)
	}

	return c.SendAll(buffer.Bytes())
}

func (c *Client) SendEndBetting() error {
	var buffer bytes.Buffer
	
	binary.Write(&buffer, binary.LittleEndian, c.agency)
	binary.Write(&buffer, binary.LittleEndian, END_BETTING)

	return c.SendAll(buffer.Bytes())
}

func (c *Client) SendRequestBetWinners() error {
	var buffer bytes.Buffer
	
	binary.Write(&buffer, binary.LittleEndian, c.agency)
	binary.Write(&buffer, binary.LittleEndian, REQUEST_WINNERS)

	return c.SendAll(buffer.Bytes())
}


func (c *Client) MakeBets(done chan bool) error {
	defer c.Destroy()

	for {
		batch, err := c.betGetter.GetBatch()
        if err != nil {
            return c.SendErrorMessageAndExit("get_batch", err)
        }

        if len(batch) == 0 {
            break
        }
		
		err = c.createClientSocket()
		if err != nil {
			return err
		}

		err = c.SendBatch(batch)
		if err != nil {
			return c.SendErrorMessageAndExit("send_message", err)
		}

		buf := make([]byte, 1)
		_, err = c.conn.Read(buf)
		if err != nil {
			return c.SendErrorMessageAndExit("receive_message", err)
		}
		
		ResponseStatus(buf[0]).logSendBatchStatus(c.betGetter.lastBatchSize)

		c.conn.Close()
		c.conn = nil

		select {
		case <-done:
			return errors.New("Should break: got SIGTERM")
		case <-time.After(c.config.LoopPeriod):
			// continue looping	
		}
	}

	return nil
}

func (c *Client) AnnounceEndBet() error {
	err := c.createClientSocket()
	if err != nil {
		return err
	}

	err = c.SendEndBetting()
	if err != nil {
		return c.SendErrorMessageAndExit("send_end_bet", err)
	}

	buf := make([]byte, 1)
	_, err = c.conn.Read(buf)
	if err != nil {
		return c.SendErrorMessageAndExit("receive_end_bet_response", err)
	}
	ResponseStatus(buf[0]).logEndBetsStatus()

	c.conn.Close()
	c.conn = nil

	return nil
}

func (c *Client) GetBetWinners() error{
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		err := c.SendRequestBetWinners()
		if err != nil {
			return err
		}

	}
	return nil
}

func (c *Client) ExecuteLotteryClient(done chan bool) {
	err := c.MakeBets(done)
	if err != nil {
		return
	}

	err = c.AnnounceEndBet()
	if err != nil {
		return
	}
	log.Infof("no hubo errorcipis")
}
