package common

type ResponseStatus uint8

const (
    OK ResponseStatus = iota
    ERR
    BAD_REQUEST
    ABORT
    LOTTERY_NOT_DONE // from here on it is unhandled
    SEND_WINNERS
    NO_MORE_BETS_ALLOWED
)

func (status ResponseStatus) GetStatusProperties() (errorMsg string){
        switch status {
        case OK:
        return ""
        case ERR:
        return "bet batch was not stores correctly by server"
        case BAD_REQUEST:
        return "bet batch was not sent appropriately. either there was data missing, or the batch amount was parsed incorrectly"
        case ABORT:
        return "server aborted"
        default:
        return "server returned unknown state"
        }
}

// even tho it belongs to client, it keeps it cleaner to leave it here
func (s ResponseStatus) logSendBatchStatus(batchSize int){
	errMsg := s.GetStatusProperties()
	if errMsg != "" {
		log.Infof("action: apuesta_enviada | result: fail | cantidad: %v | error: %v",
            batchSize,
			errMsg,
		)
	}
	log.Infof("action: apuesta_enviada | result: success | cantidad: %v",
		batchSize,
	)
}

// even tho it belongs to client, it keeps it cleaner to leave it here
func (s ResponseStatus) logEndBetsStatus(){
	errMsg := s.GetStatusProperties()
	if errMsg != "" {
		log.Infof("action: receive_end_bet_response | result: fail | error: %v",
			errMsg,
		)
	}
	log.Infof("action: receive_end_bet_response | result: success")
}