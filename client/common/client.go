package common

import (
    "bufio"
    "fmt"
    "net"
    "time"
    "os"
    "os/signal"
    "syscall"
    "strings"

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

        bet := NewBetFromEnv()
        betMessage := BuildBetMessage(c.config.ID, bet)

        log.Infof("action: send_message | result: in_progress | client_id: %v | msg: %v", c.config.ID, betMessage)
        if err := c.SendMessage(betMessage); err != nil {
            log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
                c.config.ID,
                err,
            )
            return
        }

        msg, err := c.ReceiveMessage()
        c.conn.Close()

        if err != nil {
            log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
                c.config.ID,
                err,
            )
            return
        }

        log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
            c.config.ID,
            msg,
        )

        parsedMessage, err := Deserialize(msg)
        if err != nil {
            log.Errorf("action: parse_message | result: fail | client_id: %v | error: %v",
                c.config.ID,
                err,
            )
            return
        }

        parsedMessage.LogMessage(os.Getenv("DOCUMENTO"), os.Getenv("NUMERO"), c.config.ID)

        // Wait a time between sending one message and the next one
        time.Sleep(c.config.LoopPeriod)

    }
    log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

// BuildBetMessage constructs the message to be sent to the server with the bet information
func BuildBetMessage(id string, bet Bet) string {
    return fmt.Sprintf("%s %s\n", id, bet.Serialize())
}

// ReceiveMessage reads a message until a newline is encountered
func (c *Client) ReceiveMessage() (string, error) {
    reader := bufio.NewReader(c.conn)
    message, err := reader.ReadString('\n')
    if err != nil {
        log.Errorf("action: receive_message | result: fail | dni: %v | numero: %v | error: %v", 
            os.Getenv("DOCUMENTO"), 
            os.Getenv("NUMERO"), 
            err,
        )	
        return "", err
    }
    return strings.TrimSpace(message), nil
}

// SendMessage ensures that all the message is sent to the server
func (c *Client) SendMessage(message string) error {
    messageBytes := []byte(message)
    totalSent := 0
    for totalSent < len(messageBytes) {
        sent, err := c.conn.Write(messageBytes[totalSent:])
        if err != nil {
            log.Errorf("action: send_message | result: fail | dni: %v | numero: %v | error: %v", 
                os.Getenv("DOCUMENTO"), 
                os.Getenv("NUMERO"), 
                err,
            )
            return err
        }
        totalSent += sent
    }
    return nil
}
