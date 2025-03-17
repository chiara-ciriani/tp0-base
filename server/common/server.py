import signal
import socket
import logging

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
            # TODO: Modify the receive to avoid short-reads
            msg = self._client_socket.recv(1024).rstrip().decode('utf-8')
            addr = self._client_socket.getpeername()
            logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')
            # TODO: Modify the send to avoid short-writes
            self._client_socket.send("{}\n".format(msg).encode('utf-8'))
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
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
