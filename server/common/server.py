import csv
import signal
import socket
import logging
import time
from typing import List

from common.utils import Bet, ShouldReadStreamError, store_bets
from common.response import ResponseStatus
from common.errors import BetBatchError
from common.request import MessageCode

AGENCY_LEN=1
BATCH_LEN=1
MSG_CODE_LEN=1

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
            self._client_socket.send(ResponseStatus(2).value.to_bytes(1, 'little'))
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

    def __obtain_all_batch_bets(self, agency: int, data: bytes) -> List[Bet]:
        """
        Receives all <number> of bets from stream,

        If it doesn't receive all the information it should,
        then server replies with an error, stating how many
        bets were read appropriately.
        """
        batch_num = int.from_bytes([data[0]], 'little')
        data=data[1:]

        bets=[]
        while len(bets) < batch_num:
            try:
                # Intentar deserializar una apuesta
                bet, data = Bet.deserialize(agency, data)
                bets.append(bet)
            except ShouldReadStreamError:
                # Leer más datos del socket si ocurre un error de deserialización
                msg = self._client_socket.recv(1024)
                if not msg:
                    raise BetBatchError("There was an error parsing bet batch: couldnt get all required bets")
                data += msg  # Concatenar los datos recibidos al buffer existente
                
        return bets

    def __handle_message(self, code: MessageCode, agency: int, data: bytes):
        if code == MessageCode.BET:
            print("handle bets")
            return self.__obtain_all_batch_bets(agency, data)
        else:
            print("unimplemented message", code)

    def __receive_bet_batch(self) -> List[Bet]:
        curr = 0
        # read as much as i can to avoid too many reads
        msg = self._client_socket.recv(1024)
        if not msg:
            raise BetBatchError("Client ended the connection")
        agency = int.from_bytes([msg[curr]], 'little')
        curr += AGENCY_LEN

        msg_code = int.from_bytes([msg[curr]], 'little')
        curr += MSG_CODE_LEN
    
        return self.__handle_message(MessageCode(msg_code), agency, msg[curr:])

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        status = ResponseStatus(0)
        try:
            bets = self.__receive_bet_batch()
            store_bets(bets)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
            self._client_socket.sendall(status.value.to_bytes(1, 'little'))
        except BetBatchError as e:
            # error was because either client closed connection or because he sent wrong batch information
            logging.error(f'action: receive_message | result: fail | error: {e}')
        except (OSError, csv.Error) as e:
            if self._should_stop:
                # client socket has already been closed somewhere else and should ignore err 
                return
            logging.error(f'action: receive_message | result: fail | error: {e}')
            status = ResponseStatus(1)
            self._client_socket.send(status.value.to_bytes(1, 'little'))
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
