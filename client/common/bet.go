package common

import (
    "os"
    "strings"
)

const PROTOCOL_DELIMITER=","

// Bet represents a bet made by a person
type Bet struct {
    Name       string
    Surname    string
    Document   string
    BirthDate  string
    Number     string
}

// Serialize converts a Bet struct into a string in the format:
// Name,Surname,Document,BirthDate,Number
func (b *Bet) Serialize() string {
    return strings.Join([]string{
        b.Name,
        b.Surname,
        b.Document,
        b.BirthDate,
        b.Number,
    }, PROTOCOL_DELIMITER)
}

// NewBetFromEnv creates a new Bet instance from environment variables
func NewBetFromEnv() Bet {
    return Bet{
        Name:      os.Getenv("NOMBRE"),
        Surname:   os.Getenv("APELLIDO"),
        Document:  os.Getenv("DOCUMENTO"),
        BirthDate: os.Getenv("NACIMIENTO"),
        Number:    os.Getenv("NUMERO"),
    }
}
