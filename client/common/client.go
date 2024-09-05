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
			if err:=c.SendSimpleMessage(END_CONNECTION); err == nil {
				log.Infof("action: send_exit | result: success | client_id: %v",
					c.config.ID,
				)
			}
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

func (c *Client) SendErrorMessageAndExit(action string, err error) error {
	// socket failed because goroutine closed it
	if c.end {
		return nil
	}

	log.Errorf("action: %v | result: fail | client_id: %v | error: %v",
		action,
		c.config.ID,
		err,
	)
	c.Destroy()
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

func (c *Client) SendSimpleMessage(mc MessageCode) error {
	var buffer bytes.Buffer
	
	binary.Write(&buffer, binary.LittleEndian, c.agency)
	binary.Write(&buffer, binary.LittleEndian, mc)

	return c.SendAll(buffer.Bytes())
}

func (c *Client) MakeBets(done chan bool) error {

	for {
		batch, err := c.betGetter.GetBatch()
        if err != nil {
            return c.SendErrorMessageAndExit("get_batch", err)
        }

        if len(batch) == 0 {
            break
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
	err := c.SendSimpleMessage(END_BETTING)
	if err != nil {
		return c.SendErrorMessageAndExit("send_end_bet", err)
	}

	log.Infof("action: send_end_bet | result: success | agency: %v",
			c.agency,
	)

	buf := make([]byte, 1)
	_, err = c.conn.Read(buf)
	if err != nil {
		return c.SendErrorMessageAndExit("receive_end_bet_response", err)
	}
	ResponseStatus(buf[0]).logStatus("receive_end_bet_response")

	return nil
}

// returns whether it should keep asking or if it already got what it wanted
func (c *Client) ParseBetWinnersResponse() []uint32 {
	winners := []uint32{}
	buf := make([]byte, 1024)
	r, err := c.conn.Read(buf)
	if err != nil {
		c.SendErrorMessageAndExit("recibir_ganadores", err)
		return winners
	}

	i:=0
	status := ResponseStatus(buf[i])
	i += 1
	if status == LOTTERY_NOT_DONE {
		status.logStatus("consulta_ganadores")
		return winners
	} else if status != SEND_WINNERS{
		status.logStatus("consulta_ganadores")
		return winners
	}

	winner_num := int(binary.LittleEndian.Uint16(buf[i:i+2]))
	i+=2

	for win := 1; win <= winner_num; win++{
		if i + 4 > r {
			c.SendErrorMessageAndExit("consulta_ganadores", errors.New("Server sent winners incorrectly"))
			return []uint32{}
		}
		winner := binary.LittleEndian.Uint32(buf[i:i+4])
		winners = append(winners, winner)
		i+=4
	}

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v",
		winner_num,
	)
	
	return winners 
}

func (c *Client) GetBetWinners(done chan bool) []uint32 {

	err := c.SendSimpleMessage(REQUEST_WINNERS) 
	if err != nil {
		return []uint32{}
	}
	
	return c.ParseBetWinnersResponse()
}

func (c *Client) ExecuteLotteryClient(done chan bool) []uint32{
	defer c.Destroy()
	
	err := c.createClientSocket()
	if err != nil {
		return []uint32{}
	}

	err = c.MakeBets(done)
	if err != nil {
		return []uint32{}
	}
	err = c.AnnounceEndBet()
	if err != nil {
		return []uint32{}
	}

	return c.GetBetWinners(done)
}
