import signal
import socket
import logging

from common.exceptions import InvalidMessageError
from common.message_type import MessageType
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
            finished = False
            while not finished:
                agency_id, message_type, message = self.__receive_message()
                logging.info(f'action: receive_message | result: success | agency_id: {agency_id} | message_type: {message_type}')
            
                finished = self.__handle_received_message(agency_id, message_type, message)

        except OSError as e:
            logging.error("action: apuesta_recibida | result: fail | error: {e}")
            response_message = ResponseStatus.ERROR
            self.__send_message(response_message.value)
        except InvalidMessageError as e:
            logging.error(f'action: receive_message | result: fail | error: {e}')
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

    def __obtain_bets(self, agency_id, bet_info):
        bets=[]
        total_bytes = len(bet_info)
        total_bytes_deserialized = 0
        while total_bytes_deserialized < total_bytes:
            bet, bytes_deserialized = Bet.deserialize(agency_id, bet_info, total_bytes_deserialized)
            total_bytes_deserialized += bytes_deserialized
            bets.append(bet)
        return bets
        
    def __handle_batch_message(self, agency_id, message):
        """
        Handle a batch message from the client
        """
        bets = self.__obtain_bets(agency_id, message)
        store_bets(bets)
        logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')

        response_message = ResponseStatus.OK
        self.__send_message(response_message.value)

    def __handle_received_message(self, agency_id, message_type, message):
        """
        Handle a received message from the client
        """
        if message_type == MessageType.BATCH:
            self.__handle_batch_message(agency_id, message)
            return False
        elif message_type == MessageType.BATCH_END:
            logging.info(f"action: client_finished_sending_bets | result: success | agency: {agency_id}")
            return True
        
        raise InvalidMessageError(f"Invalid message type: {message_type}")

    def __receive_exact_message(self, length_to_read):
        message = self._client_socket.recv(length_to_read)
        
        while len(message) < length_to_read:
            message_read = self._client_socket.recv(length_to_read)
            if not message_read: 
                return message
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
        message_type = MessageType(int.from_bytes(received_message[AGENCY_ID_LEN:AGENCY_ID_LEN+MSG_TYPE_LEN], 'big'))

        return agency_id, message_type, received_message[AGENCY_ID_LEN+MSG_TYPE_LEN:]
    
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
