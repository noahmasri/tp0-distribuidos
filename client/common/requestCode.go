package common

type MessageCode uint8

const (
    BET	MessageCode = iota
    END_BETTING
    REQUEST_WINNERS
)

