package common

import (
    "fmt"
    "strings"

    "github.com/op/go-logging"
)

const (
    CONFIRMATION = "CONFIRMATION"
    ERROR        = "ERROR"
)

const MESSAGE_DELIMITER=":"

var messageLog = logging.MustGetLogger("messageLog")

// Message represents a message with a type and content
type Message struct {
    Type    string
    Content string
}

// Serialize converts a Message struct into a string
func (m *Message) Serialize() string {
    return fmt.Sprintf("%s%s%s", m.Type, MESSAGE_DELIMITER, m.Content)
}

// Deserialize converts a string into a Message struct
func Deserialize(message string) (*Message, error) {
    parts := strings.SplitN(message, MESSAGE_DELIMITER, 2)
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid message format: %s", message)
    }
    return &Message{
        Type:    parts[0],
        Content: parts[1],
    }, nil
}

// LogMessage logs the message based on its type
func (message *Message) LogMessage(dni string, numero string, id string) {
    switch message.Type {
    case CONFIRMATION:
        messageLog.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", dni, numero)
    case ERROR:
        messageLog.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | error: %v", 
            dni, 
            numero, 
            message.Content)
    default:
        messageLog.Errorf("action: unknown_message_type | result: fail | id: %v | msg: %v", id, message.Serialize())
    }
}