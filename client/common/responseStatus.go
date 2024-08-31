package common

type ResponseStatus uint8

const (
    OK    ResponseStatus = 0
    ERR   ResponseStatus = 1
    ABORT ResponseStatus = 2
)