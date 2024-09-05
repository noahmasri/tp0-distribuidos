package common

type ResponseStatus uint8

const (
    OK ResponseStatus = iota
    ERR
    BAD_REQUEST
    ABORT
    LOTTERY_NOT_DONE
    SEND_WINNERS
    NO_MORE_BETS_ALLOWED
)

func (status ResponseStatus) GetStatusProperties() (errorMsg string){
        switch status {
        case OK, SEND_WINNERS:
            return ""
        case ERR:
            return "bet batch was not stores correctly by server"
        case BAD_REQUEST:
            return "bet batch was not sent appropriately. either there was data missing, or the batch amount was parsed incorrectly"
        case ABORT:
            return "server aborted"
        case LOTTERY_NOT_DONE:
            return "server waiting for clients to end"
        case NO_MORE_BETS_ALLOWED:
            return "cannot send any more bets, since already stated we were done betting"
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

// generic log status
func (s ResponseStatus) logStatus(action string){
	errMsg := s.GetStatusProperties()
	if errMsg != "" {
		log.Infof("action: %v | result: fail | error: %v",
            action,
			errMsg,
		)
	}
	log.Infof("action: %v | result: success",
        action,
    )
}