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
    "encoding/binary"
    "strings"

    "github.com/op/go-logging"
)

const WINNERS_LEN=2
const DOCUMENT_LEN=4
const MSG_SIZE_LEN=2
const MSG_TYPE_LEN=1

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
        down:  false,
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

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
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
            log.Infof("action: send_batch_message | result: success | client_id: %v | batch_length: %v", c.config.ID, len(batch))

            response, _, err := c.ReceiveMessage()

            if err != nil {
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

    log.Infof("action: get_lottery_winners | result: in_progress | client_id: %v", c.config.ID)
    if err := c.GetLotteryWinners(); err != nil {
        log.Infof("action: get_lottery_winners | result: fail | client_id: %v", c.config.ID)
        return
    }
    log.Infof("action: get_lottery_winners | result: success | client_id: %v", c.config.ID)
}

// TO DO: DOCUMENTACION
func (c *Client) GetLotteryWinners() error {
    for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
        if err := c.createClientSocket(); err != nil {
            return err
        }
        // Send winners request message
        winnersRequestMessage, err := BuildWinnersRequestMessage(c.config.ID)
        if err != nil {
            log.Errorf("action: build_winners_request_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
            return err
        }
        if err := c.SendMessage(winnersRequestMessage); err != nil {
            log.Errorf("action: send_winners_request_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
            return err
        }
        log.Infof("action: send_winners_request_message | result: success | client_id: %v", c.config.ID)

        c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

        winners := c.ReceiveWinnersRequestResponse()
        if winners != nil {
            log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(winners))
            break
        }
        
        c.conn.Close()
        
        time.Sleep(c.config.LoopPeriod)
    }
    return nil
}

// ReceiveWinnersRequestResponse reads the response from the server to the winners request message
// TO DO: TRADUCIR
// Hay dos tipos de tipos de mensajes que el servidor puede enviar: SEND_WINNERS, BET_NOT_FINISHED
// Si el servidor envía SEND_WINNERS, el cliente debe recibir los ganadores y mostrarlos en la consola
// Si el servidor envía BET_NOT_FINISHED, el cliente debe esperar un tiempo y volver a enviar el mensaje de solicitud de ganadores
// Devuelve los ganadores
func (c *Client) ReceiveWinnersRequestResponse() ([]uint32) {
    responseStatus, winnersBytes, err := c.ReceiveMessage()
    if err != nil {
        return nil
    }

    if responseStatus == SEND_WINNERS {
        winnersLen := len(winnersBytes)
        winners := []uint32{}
        for i := 0; i < winnersLen; i += DOCUMENT_LEN {
            winner := binary.BigEndian.Uint32(winnersBytes[i : i+DOCUMENT_LEN])
            winners = append(winners, winner)
        }

        return winners
    }
    return nil
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

    response, _, err := c.ReceiveMessage()
    if err != nil {
        return err
    }

    if response == OK {
        log.Infof("action: receive_server_end_message_confirmation | result: success | client_id: %v", c.config.ID)
    } else {
        log.Errorf("action: receive_server_end_message_confirmation | result: fail | client_id: %v", c.config.ID)
    }

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

// TO DO: DOCU
func (c *Client) ReceiveMessage() (ResponseStatus, []byte, error) {
    message_len_bytes, err := c.ReceiveExactMessage(MSG_SIZE_LEN)
    if err != nil {
        return ResponseStatus(1), nil, err
    }

    message_len := int(binary.BigEndian.Uint16(message_len_bytes))

    received_message, err := c.ReceiveExactMessage(message_len)
    if err != nil {
        return ResponseStatus(1), nil, err
    }

    message_type := int(received_message[0])
    response_status := ResponseStatus(message_type)
    
    return response_status, received_message[MSG_TYPE_LEN:], nil
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
            log.Errorf("action: send_message | result: fail | error: %v", err)
            return err
        }
        totalSent += sent
    }
    return nil
}