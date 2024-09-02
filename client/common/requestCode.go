package common

type MessageCode uint8

const (
    BET				MessageCode = 0
    END_BETTING		MessageCode = 1
    REQUEST_WINNERS MessageCode = 2
)

