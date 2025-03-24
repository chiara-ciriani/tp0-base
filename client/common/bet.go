package common

import (
    "fmt"
    "bytes"
    "encoding/binary"
    "os"
    "strconv"
)

const BET_FIELDS_SEPARATOR=","

// Bet represents a bet made by a person
type Bet struct {
    Name       string
    Surname    string
    Document   uint32
    BirthDate  string
    Number     uint16
}

// NewBetFromEnv creates a new Bet instance from environment variables
func NewBetFromEnv() (Bet, error) {
    document, err := strconv.ParseUint(os.Getenv("DOCUMENTO"), 10, 32)
    if err != nil {
        return Bet{}, fmt.Errorf("Failed to parse DOCUMENTO: %w", err)
    }

    number, err := strconv.ParseUint(os.Getenv("NUMERO"), 10, 16)
    if err != nil {
        return Bet{}, fmt.Errorf("Failed to parse NUMERO: %w", err)
    }

    return Bet{
        Name:      os.Getenv("NOMBRE"),
        Surname:   os.Getenv("APELLIDO"),
        Document:  uint32(document),
        BirthDate: os.Getenv("NACIMIENTO"),
        Number:    uint16(number),
    }, nil
}

// SerializeToBytes serializes the Bet struct into a byte slice
func (bet *Bet) SerializeToBytes() ([]byte, error) {
    buffer := new(bytes.Buffer)

    // Serialize Name
    if err := binary.Write(buffer, binary.BigEndian, uint8(len(bet.Name))); err != nil {
        return nil, fmt.Errorf("Failed to write bet name length: %w", err)
    }

    if err := binary.Write(buffer, binary.BigEndian, []byte(bet.Name)); err != nil {
        return nil, fmt.Errorf("Failed to write bet name: %w", err)
    }

    // Serialize Surname
    if err := binary.Write(buffer, binary.BigEndian, uint8(len(bet.Surname))); err != nil {
        return nil, fmt.Errorf("Failed to write bet surname length: %w", err)
    }

    if err := binary.Write(buffer, binary.BigEndian, []byte(bet.Surname)); err != nil {
        return nil, fmt.Errorf("Failed to write bet surname: %w", err)
    }

    // Serialize Document
    if err := binary.Write(buffer, binary.BigEndian, bet.Document); err != nil {
        return nil, fmt.Errorf("Failed to write bet document: %w", err)
    }

    // Serialize BirthDate
    if err := binary.Write(buffer, binary.BigEndian, []byte(bet.BirthDate)); err != nil {
        return nil, fmt.Errorf("Failed to write bet birth date: %w", err)
    }

    // Serialize Number
    if err := binary.Write(buffer, binary.BigEndian, bet.Number); err != nil {
        return nil, fmt.Errorf("Failed to write bet number: %w", err)
    }

    return buffer.Bytes(), nil
}
