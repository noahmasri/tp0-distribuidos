import signal
import socket
import logging
import time

from common.utils import Bet, read_exact


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._should_stop = False
        self._client_socket = None 
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

    def __shutdown_gracefully(self, signum, frame):
        logging.info(f'action: shutdown_gracefully | result: in_progress | msg: received SIGTERM signal')
        self._should_stop = True

        if self._client_socket is not None:
            self._client_socket.close()
            logging.info(f'action: close_client_socket | result: success | msg: received SIGTERM signal')

        self._server_socket.close()
        logging.info(f'action: close_server_socket | result: success | msg: received SIGTERM signal')

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        # TODO: Modify this program to handle signal to graceful shutdown
        # the server
        signal.signal(signal.SIGTERM, self.__shutdown_gracefully)

        while not self._should_stop:
            try:
                self.__accept_new_connection()
                self.__handle_client_connection()
            except OSError as e:
                if self._should_stop:
                    # ignore error if socket was closed manually
                    logging.info(f'action: shutdown_gracefully | result: success')
                else:
                    # server exits if there was a problem with the socket
                    logging.info(f'action: error_exiting | result: in_progress | error: {e}')
                break        

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            # TODO: Modify the receive to avoid short-reads
            agency_bytes = read_exact(self._client_socket, 1)
            agency = int.from_bytes(agency_bytes, 'big')
            bet_len_bytes = read_exact(self._client_socket, 2)
            bet_len = int.from_bytes(bet_len_bytes, 'little')
            bet_bytes = read_exact(self._client_socket, bet_len)
            bet = Bet.deserialize(agency, bet_bytes)
            """
            addr = self._client_socket.getpeername()
            logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')
            # TODO: Modify the send to avoid short-writes
            self._client_socket.send("{}\n".format(bet).encode('utf-8'))
            """
            self._client_socket.send("{}\n".format(bet).encode('utf-8'))
        except OSError as e:
            if self._should_stop:
                # client socket has already been closed somewhere else and should ignore err 
                return
            logging.error("action: receive_message | result: fail | error: {e}")
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
