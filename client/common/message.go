package common

import (
    "fmt"
    "bytes"
    "encoding/binary"
    "strconv"
)

type MessageType int

const (
    BATCH MessageType = iota
    BATCH_END
    WINNERS_REQUEST
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

// SerializeMessageTypeToBytes serializes the message type into a byte slice
func SerializeMessageTypeToBytes(messageType MessageType) ([]byte, error) {
    buffer := new(bytes.Buffer)

    if err := binary.Write(buffer, binary.BigEndian, uint8(messageType)); err != nil {
        return nil, fmt.Errorf("Failed to write message type: %w", err)
    }

    return buffer.Bytes(), nil
}

// BuildHeader builds the header with message length, message type, and agency ID
func BuildHeader(messageType MessageType, agencyId string, messageLength int) ([]byte, error) {
    buffer := new(bytes.Buffer)

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

    // Serialize message type
    messageTypeBytes, err := SerializeMessageTypeToBytes(messageType)
    if err != nil {
        return nil, err
    }
    buffer.Write(messageTypeBytes)

    return buffer.Bytes(), nil
}

// BuildBatchMessage builds the batch message from a list of bets
func BuildBatchMessage(agencyId string, batch []Bet) ([]byte, error) {
    buffer := new(bytes.Buffer)

    // Serialize each bet
    for _, bet := range batch {
        betBytes, err := bet.SerializeToBytes()
        if err != nil {
            return nil, err
        }
        buffer.Write(betBytes)
    }

    // Calculate message length
    // 2 due to 1 byte for agency ID and 1 byte for message type
    messageLength := 2 + buffer.Len()

    // Serialize header
    headerBytes, err := BuildHeader(BATCH, agencyId, messageLength)
    if err != nil {
        return nil, err
    }

    // Prepend header to the buffer
    finalBuffer := new(bytes.Buffer)
    finalBuffer.Write(headerBytes)
    finalBuffer.Write(buffer.Bytes())

    return finalBuffer.Bytes(), nil
}

// BuildBatchEndMessage builds the batch end message
func BuildBatchEndMessage(agencyId string) ([]byte, error) {
    buffer := new(bytes.Buffer)

    // Calculate message length (header only)
    messageLength := 2 // 1 byte for agency ID, 1 byte for message type

    // Serialize header
    headerBytes, err := BuildHeader(BATCH_END, agencyId, messageLength)
    if err != nil {
        return nil, err
    }
    buffer.Write(headerBytes)

    return buffer.Bytes(), nil
}

// BuildWinnersRequestMessage builds the request winners message
func BuildWinnersRequestMessage(agencyId string) ([]byte, error) {
    buffer := new(bytes.Buffer)

    // Calculate message length (header only)
    messageLength := 2 // 1 byte for agency ID, 1 byte for message type

    // Serialize header
    headerBytes, err := BuildHeader(WINNERS_REQUEST, agencyId, messageLength)
    if err != nil {
        return nil, err
    }
    buffer.Write(headerBytes)

    return buffer.Bytes(), nil
}
