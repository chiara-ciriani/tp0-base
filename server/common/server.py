import signal
import socket
import logging

from common.exceptions import InvalidMessageError
from common.message_type import MessageType
from common.response_status import ResponseStatus
from common.utils import Bet, store_bets, get_winners

MSG_SIZE_LEN = 2
AGENCY_ID_LEN = 1
MSG_TYPE_LEN = 1

DOCUMENT_LEN = 4
WINNERS_LEN = 2

class Server:
    def __init__(self, port, listen_backlog, total_agencies):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_socket = None
        self._down = False
        self._required_agencies = total_agencies
        self._done_agencies = set()

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
                if self._down or not self._client_socket: 
                    break
                self.__handle_client_connection()
            except OSError as e:
                if not self._client_socket:
                    logging.error(f"action: handle_client_connection | result: client disconnected")
                    break
                else:
                    logging.error(f"action: handle_client_connection | result: fail | error: {e}")
                    self._client_socket.close()
                    self._client_socket = None
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
            logging.error(f"action: apuesta_recibida | result: fail | error: {e}")
            response_status = ResponseStatus.ERROR
            self.__send_message(response_status.value)
        except InvalidMessageError as e:
            logging.error(f'action: receive_message | result: fail | error: {e}')
        except EOFError:
            logging.info("action: client_disconnected | result: success")
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
        logging.info(f'action: apuesta_recibida | result: success | agency_id: {agency_id} | cantidad: {len(bets)}')

        response_status = ResponseStatus.OK
        self.__send_message(response_status.value)

    def __handle_batch_end_message(self, agency_id):
        """
        Handle an end message from the client
        """
        self._done_agencies.add(agency_id)
        logging.info(f"action: client_finished_sending_bets | result: success | agency: {agency_id}")

        if len(self._done_agencies) == self._required_agencies:
            logging.info(f"action: sorteo | result: success")
        
    def __build_winners_message(self, agency_id):
        winners = get_winners(self._required_agencies)
        agency_winners = winners.get(agency_id, [])
        encoded_winners = b''.join([int(winner).to_bytes(DOCUMENT_LEN, 'big') for winner in agency_winners])
        return encoded_winners

    def __send_winners(self, agency_id):
        response_status = ResponseStatus.SEND_WINNERS
        response_message = self.__build_winners_message(agency_id)
        self.__send_message(response_status.value, response_message)
        logging.info(f"action: winners_sent | result: success | agency: {agency_id}")

    def __handle_winners_request(self, agency_id):
        """
        Handle a winners request from the client
        """
        if len(self._done_agencies) == self._required_agencies:
            self.__send_winners(agency_id)
            logging.info(f"action: client_requested_winners | result: success | agency: {agency_id}")
        else:
            response_status = ResponseStatus.BET_NOT_FINISHED
            self.__send_message(response_status.value)
            logging.info(f"action: bet_not_finished_message_sent | result: success | agency: {agency_id}")

    def __handle_received_message(self, agency_id, message_type, message):
        """
        Handle a received message from the client
        """
        if message_type == MessageType.BATCH:
            self.__handle_batch_message(agency_id, message)
            return False
        elif message_type == MessageType.BATCH_END:
            self.__handle_batch_end_message(agency_id)
            return True
        elif message_type == MessageType.WINNERS_REQUEST:
            self.__handle_winners_request(agency_id)
            return True
        
        raise InvalidMessageError(f"Invalid message type: {message_type}")

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
        message_type = MessageType(int.from_bytes(received_message[AGENCY_ID_LEN:AGENCY_ID_LEN+MSG_TYPE_LEN], 'big'))
        return agency_id, message_type, received_message[AGENCY_ID_LEN+MSG_TYPE_LEN:]
    
    def __build_message_to_client(self, response_status, message=None):
        """
        TO DO: documentar
        """
        response_status_bytes = response_status.to_bytes(MSG_TYPE_LEN, 'big')
        message_bytes = message if message else b''
        message_size = MSG_TYPE_LEN + len(message_bytes)
        return message_size.to_bytes(MSG_SIZE_LEN, 'big') + response_status_bytes + message_bytes

    def __send_message(self, response_status, message=None):
        """
        Send a message to the client
        """
        total_sent = 0
        message_bytes = self.__build_message_to_client(response_status, message)
        while total_sent < len(message_bytes):
            try:
                sent = self._client_socket.send(message_bytes[total_sent:])
                if sent == 0:
                    raise OSError("Socket connection broken")
                total_sent += sent
            except OSError as e:
                logging.error(f"action: send_message | result: fail | error: {e}")
                break
