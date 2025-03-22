package common

import (
    "fmt"
    "bytes"
    "encoding/binary"
    "strconv"
    "strings"
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

// TransformStringToBet converts a string representation of a bet to a Bet struct
func TransformStringToBet(stringBet string) (Bet, error) {
    betData := strings.Split(stringBet, BET_FIELDS_SEPARATOR)
    if len(betData) != 5 {
        return Bet{}, fmt.Errorf("Invalid bet format")
    }

    document, err := strconv.ParseUint(betData[2], 10, 32)
    if err != nil {
        return Bet{}, fmt.Errorf("Invalid document number: %v", err)
    }

    number, err := strconv.ParseUint(betData[4], 10, 16)
    if err != nil {
        return Bet{}, fmt.Errorf("Invalid bet number: %v", err)
    }

    return Bet{
        Name:      betData[0],
        Surname:   betData[1],
        Document:  uint32(document),
        BirthDate: betData[3],
        Number:    uint16(number),
    }, nil
}

// REMINDER PARA MI: FALTA ENCODEAR LA AGENCY ID DE LA MISMA FORMA
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
