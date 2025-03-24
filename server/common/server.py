import signal
import socket
import logging

from common.exceptions import ClientClosedConnection
from common.response_status import ResponseStatus
from common.utils import Bet, store_bets

MSG_SIZE_LEN = 2
AGENCY_ID_LEN = 1
MSG_TYPE_LEN = 1

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_socket = None
        self._down = False

        signal.signal(signal.SIGTERM, self.__handle_sigterm)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        # TODO: Modify this program to handle signal to graceful shutdown
        # the server
        while not self._down:
            try:
                self.__accept_new_connection()
                if self._down: return
                self.__handle_client_connection()
            except OSError:
                break

    def __handle_sigterm(self, signum, frame):
        logging.info('action: sigterm_received | result: in_progress')

        logging.info('action: close_client_socket | result: in_progress')
        if self._client_socket:
            self._client_socket.shutdown(socket.SHUT_RDWR)
            self._client_socket.close()
            self._client_socket = None
        logging.info('action: close_client_socket | result: success')

        logging.info('action: close_server_socket | result: in_progress')
        self._down = True
        self._server_socket.shutdown(socket.SHUT_RDWR)
        self._server_socket.close()
        logging.info('action: close_server_socket | result: success')

        logging.info('action: sigterm_received | result: success')

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            agency_id, message = self.__receive_message()
            logging.info(f'action: receive_message | result: success | agency_id: {agency_id}')
            bet = Bet.deserialize(agency_id, message)
            store_bets([bet])
            logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.get_document()} | numero: {bet.get_number()}')
            
            response_message = ResponseStatus.OK
            self.__send_message(response_message.value)

        except ClientClosedConnection as e:
            logging.error(f"action: client_closed_connection | result: fail | error: {e}")
        except OSError as e:
            logging.error("action: apuesta_almacenada | result: fail | error: {e}")
            response_message = ResponseStatus.ERROR
            self.__send_message(response_message.value)
        except ValueError as e:
            logging.error(f"action: parse_message | result: fail | error: {e}")
            response_message = ResponseStatus.ERROR
            self.__send_message(response_message.value)
        finally:
            self._client_socket.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        self._client_socket = c

    def __receive_exact_message(self, length_to_read):
        message = self._client_socket.recv(length_to_read)
        while len(message) < length_to_read:
            message_read = self._client_socket.recv(length_to_read - len(message))
            if not message_read: 
                raise EOFError("EOF")
            message += message_read
        return message

    def __receive_message(self):
        """
        Receive a message from the client
        """
        received_message = self.__receive_exact_message(MSG_SIZE_LEN)
        message_size = int.from_bytes(received_message, 'big')

        received_message = self.__receive_exact_message(message_size)
        agency_id = int.from_bytes(received_message[0:AGENCY_ID_LEN], 'big')
        return agency_id, received_message[AGENCY_ID_LEN:]

    def __send_message(self, response_status):
        """
        Send a message to the client
        """
        total_sent = 0
        response_status_bytes = response_status.to_bytes(MSG_TYPE_LEN, 'big')
        while total_sent < len(response_status_bytes):
            try:
                sent = self._client_socket.send(response_status_bytes[total_sent:])
                if sent == 0:
                    raise OSError("Socket connection broken")
                total_sent += sent
            except OSError as e:
                logging.error(f"action: send_message | result: fail | error: {e}")
                break