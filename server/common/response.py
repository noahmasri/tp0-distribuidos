from enum import Enum

class ResponseStatus(Enum):
    OK = 0
    ERROR = 1
    BAD_REQUEST = 2
    ABORT = 3