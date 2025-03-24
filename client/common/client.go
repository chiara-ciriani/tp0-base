package common

import (
    "fmt"
    "net"
    "time"
    "os"
    "os/signal"
    "syscall"

    "github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
    ID            string
    ServerAddress string
    LoopAmount    int
    LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
    config ClientConfig
    conn   net.Conn
    down   bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
    client := &Client{
        config: config,
        down:  false,
    }
    return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
    conn, err := net.Dial("tcp", c.config.ServerAddress)
    if err != nil {
        log.Criticalf(
            "action: connect | result: fail | client_id: %v | error: %v",
            c.config.ID,
            err,
        )
    }
    c.conn = conn
    return nil
}

// Shutdown Shuts down gracefully the client by closing the connection
func (c *Client) Shutdown() {
    log.Infof("action: shutdown | result: in_progress | client_id: %v", c.config.ID)
    if c.conn != nil {
        c.conn.Close()
    }
    c.down = true
    log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGTERM)

    go func() {
        sig := <-sigs
        log.Infof("action: sigterm_received | result: in_progress | signal: %v", sig)
        c.Shutdown()
        log.Infof("action: sigterm_received | result: success | signal: %v", sig)
    }()

    // There is an autoincremental msgID to identify every message sent
    // Messages if the message amount threshold has not been surpassed
    for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
        // If the client is down, stop the loop
        if c.down {
            log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
            return
        }

        // Create the connection the server in every loop iteration. Send an
        c.createClientSocket()

        bet, err := NewBetFromEnv()
        if err != nil {
            log.Errorf("action: build_bet_from_env | result: fail | client_id: %v | error: %v",
                c.config.ID,
                err,
            )
            return
        }

        betMessage, err := BuildBetMessage(c.config.ID, bet)
        if err != nil {
            log.Errorf("action: build_bet_message | result: fail | client_id: %v | error: %v",
                c.config.ID,
                err,
            )
            return
        }

        log.Infof("action: send_message | result: in_progress | client_id: %v", c.config.ID)
        if err := c.SendMessage(betMessage); err != nil {
            log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
                c.config.ID,
                err,
            )
            return
        }

        received_message, err := c.ReceiveExactMessage(1)
        c.conn.Close()

        if err != nil {
            log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
                c.config.ID,
                err,
            )
            return
        }
        
        message_type := int(received_message[0])
        response_status := ResponseStatus(message_type)

        log.Infof("action: receive_message | result: success | client_id: %v | response_status: %v",
            c.config.ID,
            response_status,
        )

        // Log status
        if response_status == OK {
            log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", os.Getenv("DOCUMENTO"), os.Getenv("NUMERO"))
        } else if response_status == ERROR {
            log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v", os.Getenv("DOCUMENTO"), os.Getenv("NUMERO"))
        } else {
            log.Errorf("action: unknown_message_type | result: fail | client_id: %v", c.config.ID)
            return
        }

        // Wait a time between sending one message and the next one
        time.Sleep(c.config.LoopPeriod)

    }
    log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

// TO DO: DOCU
func (c *Client) ReceiveExactMessage(lengthToRead int) ([]byte, error) {
    data := make([]byte, lengthToRead)
    bytesRead, err := c.conn.Read(data)
    if err != nil {
        return nil, err
    }
    totalBytesRead := bytesRead
    for totalBytesRead < lengthToRead {
        bytesRead, err = c.conn.Read(data[totalBytesRead:])
        if err != nil {
            return nil, err
        }
        if bytesRead == 0 {
            return nil, fmt.Errorf("EOF")
        }
        totalBytesRead += bytesRead
    }
    return data, nil
}

// SendMessage ensures that all the message is sent to the server
func (c *Client) SendMessage(message []byte) error {
    totalSent := 0
    for totalSent < len(message) {
        sent, err := c.conn.Write(message[totalSent:])
        if err != nil {
            return err
        }
        totalSent += sent
    }
    return nil
}