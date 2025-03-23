package common

type ResponseStatus uint8

const (
    OK ResponseStatus = iota
    ERROR
    SEND_WINNERS
    BET_NOT_FINISHED
)