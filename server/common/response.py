from enum import Enum

class ResponseStatus(Enum):
    OK = 0
    ERROR = 1
    BAD_REQUEST = 2
    ABORT = 3
    LOTTERY_NOT_DONE = 4
    SEND_WINNERS = 5
    NO_MORE_BETS_ALLOWED = 6