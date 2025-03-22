package common

type ResponseStatus uint8

const (
    OK ResponseStatus = iota
    ERROR
)