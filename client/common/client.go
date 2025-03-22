package common

import (
    "bufio"
    "encoding/csv"
    "fmt"
    "net"
    "time"
    "io"
    "os"
    "os/signal"
    "syscall"
    "strconv"
    "strings"

    "github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
    ID              string
    ServerAddress   string
    LoopAmount      int
    LoopPeriod      time.Duration
    BatchMaxAmount  int
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
        return err
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

    file, err := os.Open(fmt.Sprintf("./.data/agency-%v.csv", c.config.ID))
    if err != nil {
        log.Errorf("action: open_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
        return
    }
    defer file.Close()

    reader := csv.NewReader(file)

    for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
        if c.down {
            log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
            return
        }

        batch, err := c.GetBetBatch(reader)
        if err != nil {
            log.Errorf("action: get_bet_batch | result: fail | client_id: %v | error: %v", c.config.ID, err)
            return
        }

        if len(batch) > 0 {
            if err := c.createClientSocket(); err != nil {
                return
            }

            batchMessage, err := BuildBatchMessage(c.config.ID, batch)
            if err != nil {
                log.Errorf("action: build_batch_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
                return
            }

            log.Infof("action: send_batch_message | result: in_progress | client_id: %v | batch_length: %v", c.config.ID, len(batch))
            if err := c.SendMessage(batchMessage); err != nil {
                log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
                return
            }
            log.Infof("action: send_batch_message | result: success| client_id: %v | batch_length: %v", c.config.ID, len(batch))

            response, err := c.ReceiveMessage()

            if err != nil {
                log.Errorf("action: receive_message| result: fail | client_id: %v | error: %v", c.config.ID, err)
                return
            }

            // Log status
            if response == OK {
                log.Infof("action: receive_server_confirmation | result: success | client_id: %v", c.config.ID)
            } else {
                log.Errorf("action: receive_server_confirmation | result: fail | client_id: %v", c.config.ID)
                return
            }

            time.Sleep(c.config.LoopPeriod)
        }
        if err == io.EOF {
            break
        }
    }

    log.Infof("action: send_end_message | result: in_progress | client_id: %v", c.config.ID)
    if err := c.SendEndMessage(); err != nil {
        log.Infof("action: send_end_message | result: fail | client_id: %v", c.config.ID)
        return
    }
    log.Infof("action: send_end_message | result: success | client_id: %v", c.config.ID)
    log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

// SendEndMessage Sends the end message to the server
func (c *Client) SendEndMessage() error {
    if err := c.createClientSocket(); err != nil {
        return err
    }
    // Send end message
    endMessage, err := BuildBatchEndMessage(c.config.ID)
    if err != nil {
        log.Errorf("action: build_batch_end_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
        return err
    }
    if err := c.SendMessage(endMessage); err != nil {
        log.Errorf("action: send_end_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
        return err
    }
    log.Infof("action: send_end_message | result: success | client_id: %v", c.config.ID)
    c.conn.Close()
    return nil
}


// GetBetBatch reads a batch of bets from a CSV file and transform them to a Bet struct
func (c *Client) GetBetBatch(reader *csv.Reader)([]Bet, error) {
    bets := []Bet{}
    for i := 0; i < c.config.BatchMaxAmount; i++ {
        data, err := reader.Read()
        if err != nil {
            if err.Error() == "EOF" {
                break
            }
            log.Errorf("action: read_csv | result: fail | client_id: %v | error: %v", c.config.ID, err)
            return bets, err
        }

        bet, err := TransformStringToBet(strings.Join(data, BET_FIELDS_SEPARATOR))
        if err != nil {
            log.Errorf("action: transform_bet | result: fail | client_id: %v | error: %v", c.config.ID, err)
            return bets, err
        }

        bets = append(bets, bet)
    }
    return bets, nil
}

// ReceiveMessage reads a message from the server
func (c *Client) ReceiveMessage() (ResponseStatus, error) {
    reader := bufio.NewReader(c.conn)
    message, err := reader.ReadString('\n')
    if err != nil {
        log.Errorf("action: receive_message | result: fail | error: %v", err)	
        return ResponseStatus(1), err
    }
    response, err := strconv.Atoi(strings.TrimSpace(message))
    if err != nil {
        log.Errorf("action: string_conversion | result: fail | error: %v", err)	
        return ResponseStatus(1), err
    }
    return ResponseStatus(response), nil
}

// SendMessage ensures that all the message is sent to the server
func (c *Client) SendMessage(message []byte) error {
    totalSent := 0
    for totalSent < len(message) {
        sent, err := c.conn.Write(message[totalSent:])
        if err != nil {
            log.Errorf("action: send_message | result: fail | error: %v", err)
            return err
        }
        totalSent += sent
    }
    return nil
}