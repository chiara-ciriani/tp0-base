package common

import (
    "encoding/csv"
    "fmt"
    "net"
    "time"
    "io"
    "os"
    "os/signal"
    "syscall"
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

// HandleSigterm handles the SIGTERM signal to gracefully shutdown the client
func (c *Client) HandleSigterm() {
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGTERM)

    go func() {
        sig := <-sigs
        log.Infof("action: sigterm_received | result: in_progress | signal: %v", sig)
        c.Shutdown()
        log.Infof("action: sigterm_received | result: success | signal: %v", sig)
    }()
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
		down: false,
	}

	go client.HandleSigterm()

	return client
}

// Shutdown Shuts down gracefully the client by closing the connection
func (c *Client) Shutdown() {
    log.Infof("action: shutdown | result: in_progress | client_id: %v", c.config.ID)
    if c.conn != nil {
		log.Infof("action: close_connection | result: in_progress | client_id: %v", c.config.ID)
        c.conn.Close()
		log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
    }
	log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
	c.down = true
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

// StartClientLoop initializes the client, reads bets from a CSV file, and sends them to the server.
// It creates a socket connection to the server and processes bets in batches.
func (c *Client) StartClientLoop() {
    file, err := os.Open(fmt.Sprintf("./.data/agency-%v.csv", c.config.ID))
    if err != nil {
        log.Errorf("action: open_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
        return
    }
    defer file.Close()

    if err := c.createClientSocket(); err != nil {
        return
    }

    reader := csv.NewReader(file)
    if err := c.StartPlacingBets(reader); err != nil {
        c.conn.Close()
        return
    }

    c.conn.Close()

    log.Infof("action: finished_sending_bets | result: success | client_id: %v", c.config.ID)
}


// StartPlacingBets reads batches of bets from a CSV file and sends them to the server.
// It handles the process of building batch messages, sending them, and receiving server confirmations.
// If an error occurs, the process stops.
func (c *Client) StartPlacingBets(reader *csv.Reader) error {
    for {
        batch, err := c.GetBetBatch(reader)
        if len(batch) > 0 {
            batchMessage, err := BuildBatchMessage(c.config.ID, batch)
            if err != nil {
                log.Errorf("action: build_batch_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
                return err
            }

            log.Infof("action: send_batch_message | result: in_progress | client_id: %v | batch_length: %v", c.config.ID, len(batch))
            if err := c.SendMessage(batchMessage); err != nil {
                log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
                return err
            }
            log.Infof("action: send_batch_message | result: success | client_id: %v | batch_length: %v", c.config.ID, len(batch))

            received_message, err := c.ReceiveExactMessage(1)

            message_type := int(received_message[0])
            response_status := ResponseStatus(message_type)

            if err != nil {
                return err
            }

            // Log status
            if response_status == OK {
                log.Infof("action: receive_server_confirmation | result: success | client_id: %v", c.config.ID)
            } else {
                log.Errorf("action: receive_server_confirmation | result: fail | client_id: %v", c.config.ID)
                return fmt.Errorf("Server response not OK")
            }
        }
        if err != nil {
            if err == io.EOF {
                break
            }
            log.Errorf("action: read_bets | result: fail | client_id: %v | error: %v", c.config.ID, err)
            return err
        }
    }

    log.Infof("action: send_end_message | result: in_progress | client_id: %v", c.config.ID)
    if err := c.SendEndMessage(); err != nil {
        log.Infof("action: send_end_message | result: fail | client_id: %v", c.config.ID)
        return err
    }
    log.Infof("action: send_end_message | result: success | client_id: %v", c.config.ID)
    time.Sleep(c.config.LoopPeriod)
    return nil
}

// SendEndMessage Sends the end message to the server
func (c *Client) SendEndMessage() error {
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
    return nil
}


// GetBetBatch reads a batch of bets from a CSV file and transform them to a Bet struct
func (c *Client) GetBetBatch(reader *csv.Reader)([]Bet, error) {
    bets := []Bet{}
    for i := 0; i < c.config.BatchMaxAmount; i++ {
        data, err := reader.Read()
        if err != nil {
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

// ReceiveExactMessage reads an exact number of bytes from the server.
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
            log.Errorf("action: send_message | result: fail | error: %v", err)
            return err
        }
        totalSent += sent
    }
    return nil
}