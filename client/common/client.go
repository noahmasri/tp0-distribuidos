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
	config		ClientConfig
	betGetter	BetGetter	
	conn		net.Conn
	end			bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, betGetter BetGetter) *Client {
	client := &Client{
		config: config,
		betGetter: betGetter,
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

func (c *Client) logStatus(status ResponseStatus){
	errMsg := status.GetStatusProperties()
	if errMsg != "" {
		log.Infof("action: apuesta_enviada | result: fail | cantidad: %v | error: %v",
			c.betGetter.lastBatchSize,
			errMsg,
		)
	}
	log.Infof("action: apuesta_enviada | result: success | cantidad: %v",
		c.betGetter.lastBatchSize,
	)
}

func (c *Client) SendErrorMessageAndExit(done chan bool, action string, err error){
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
		done <- true
	}
}

func (c *Client) SendBatch(batch []Bet) error{
	data, err := c.EncodeBatchData(batch) 
	if err != nil{
		return err
	}

	totalWritten := 0
	for totalWritten < len(data) {
		var written int
		written, err = c.conn.Write(data[totalWritten:])
		totalWritten += written
	}
	return err
}

func (c *Client) EncodeBatchData(batch []Bet) ([]byte, error) {
	var buffer bytes.Buffer
	agency, err := strconv.ParseInt(c.config.ID, 10, 8)
	if err != nil {
		log.Fatalf("action: parse_bets | result: fail | error: %v | msg: agency number is invalid", err)
		return buffer.Bytes(), err
	}

	binary.Write(&buffer, binary.LittleEndian, uint8(agency))
	binary.Write(&buffer, binary.LittleEndian, BET)
	binary.Write(&buffer, binary.LittleEndian, uint8(len(batch)))
	for _, bet := range batch{
		bet_bytes := bet.EncodeToBytes()
		buffer.Write(bet_bytes)
	}

	return buffer.Bytes(), nil
}


func (c *Client) MakeBet(done chan bool) {
	defer c.Destroy()

	for {
		batch, err := c.betGetter.GetBatch()
        if err != nil {
            c.SendErrorMessageAndExit(done, "get_batch", err)
            break
        }

        if len(batch) == 0 {
            break
        }

		err = c.createClientSocket()
		if err != nil {
			done <- true
			return
		}

		err = c.SendBatch(batch)
		if err != nil {
			c.SendErrorMessageAndExit(done, "send_message", err)
			return
		}

		buf := make([]byte, 1)
		_, err = c.conn.Read(buf)
		if err != nil {
			c.SendErrorMessageAndExit(done, "receive_message", err)
			return
		}

		c.logStatus(ResponseStatus(buf[0]))

		c.conn.Close()
		c.conn = nil

		select {
		case <-done:
			return
		case <-time.After(c.config.LoopPeriod):
			// continue looping	
		}
	}
	done <- true
}
