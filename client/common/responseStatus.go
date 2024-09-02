package common

type ResponseStatus uint8

const (
    OK          ResponseStatus = 0
    ERR         ResponseStatus = 1
    BAD_REQUEST ResponseStatus = 2
    ABORT       ResponseStatus = 3
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