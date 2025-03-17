import signal
import socket
import logging

from common.exceptions import ClientClosedConnection
from common.message import CONFIRMATION, ERROR, Message
from common.utils import Bet, store_bets

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
            message = self.__receive_message()
            logging.info(f'action: receive_message | result: success | message: {message}')
            bet = self.__parse_message(message)
            store_bets([bet])
            logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.get_document()} | numero: {bet.get_number()}')
            response_message = Message(CONFIRMATION, "Bet received").serialize()
            self.__send_message(response_message)

        except ClientClosedConnection as e:
            logging.error(f"action: client_closed_connection | result: fail | error: {e}")
        except OSError as e:
            logging.error("action: apuesta_almacenada | result: fail | error: {e}")
            response_message = Message(ERROR, str(e)).serialize()
            self.__send_message(response_message)
        except ValueError as e:
            logging.error(f"action: parse_message | result: fail | error: {e}")
            response_message = Message(ERROR, str(e)).serialize()
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

    def __parse_message(self, message):
        """
        Parse the message and return a Bet instance
        """
        try:
            agency_id, bet_info = message.split(" ", 1)
            first_name, last_name, document, birthdate, number = bet_info.split(",")
            return Bet(agency_id, first_name, last_name, document, birthdate, number)
        except ValueError as e:
            raise ValueError(f"Invalid message format: {message}") from e

    def __receive_message(self):
        """
        Receive a message from the client
        """
        message = ''
        while message == '' or message[-1] != '\n':
            received_message = self._client_socket.recv(1024).decode('utf-8')
            if received_message == '':
                raise ClientClosedConnection('Connection closed by client')
            message += received_message
        return message.rstrip()

    def __send_message(self, message):
        """
        Send a message to the client
        """
        total_sent = 0
        message_bytes = message.encode('utf-8')
        while total_sent < len(message_bytes):
            try:
                sent = self._client_socket.send(message_bytes[total_sent:])
                if sent == 0:
                    raise OSError("Socket connection broken")
                total_sent += sent
            except OSError as e:
                logging.error(f"action: send_message | result: fail | error: {e}")
                break