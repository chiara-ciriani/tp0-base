package common

import (
    "fmt"
    "bytes"
    "encoding/binary"
    "strconv"
)

// SerializeMessageLengthToBytes serializes the length of the message into a byte slice
func SerializeMessageLengthToBytes(length int) ([]byte, error) {
    buffer := new(bytes.Buffer)

    if err := binary.Write(buffer, binary.BigEndian, uint16(length)); err != nil {
        return nil, fmt.Errorf("Failed to write message length: %w", err)
    }

    return buffer.Bytes(), nil
}

// SerializeAgencyIdToBytes serializes the agency ID into a byte slice
func SerializeAgencyIdToBytes(agencyId string) ([]byte, error) {
    id, err := strconv.Atoi(agencyId)
    if err != nil {
        return nil, fmt.Errorf("Invalid agency ID: %v", err)
    }

    buffer := new(bytes.Buffer)

    if err := binary.Write(buffer, binary.BigEndian, uint8(id)); err != nil {
        return nil, fmt.Errorf("Failed to write agency ID: %w", err)
    }

    return buffer.Bytes(), nil
}

// BuildBetMessage builds the bet message from a bet
func BuildBetMessage(agencyId string, bet Bet) ([]byte, error) {
    buffer := new(bytes.Buffer)

    // Serialize bet
    betBytes, err := bet.SerializeToBytes()
    if err != nil {
        return nil, err
    }

    messageLength := len(betBytes) + 1 // +1 from agencyIdBytes

    // Serialize message length
    messageLengthBytes, err := SerializeMessageLengthToBytes(messageLength)
    if err != nil {
        return nil, err
    }
    buffer.Write(messageLengthBytes)

    // Serialize agency ID
    agencyIdBytes, err := SerializeAgencyIdToBytes(agencyId)
    if err != nil {
        return nil, err
    }
    buffer.Write(agencyIdBytes)

    buffer.Write(betBytes)

    return buffer.Bytes(), nil
}
