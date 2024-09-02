package common

type ResponseStatus uint8

const (
    OK ResponseStatus = iota
    ERR
    BAD_REQUEST
    ABORT
    LOTTERY_NOT_DONE
    SEND_WINNERS
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