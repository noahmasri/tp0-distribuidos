import csv
import signal
import socket
import logging
import multiprocessing
from multiprocessing.pool import ThreadPool
from typing import Tuple
from common.utils import Bet, ShouldReadStreamError, get_bet_documents_from_agency, get_winners, store_bets
from common.response import ResponseStatus
from common.errors import BetBatchError, ClientCannotSendMoreBetsError, NoMessageReceivedError, WrongHeaderError
from common.request import MessageCode

MAX_PROCESSES=5

AGENCY_LEN=1
BATCH_LEN=1
MSG_CODE_LEN=1
AGENCY_WINNERS_LEN=2
# amount of agencies required to close betting
AGENCY_CLOSING_NUMBER=5

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._should_stop = False

        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

        self._agencies_done = set()
        self._agencies_convar = multiprocessing.Condition()

        self.winners = None
        self.file_lock = multiprocessing.Lock()

    def __shutdown_gracefully(self, signum, frame):
        logging.info(f'action: shutdown_gracefully | result: in_progress | msg: received SIGTERM signal')
        self._should_stop = True

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

        with ThreadPool(processes=MAX_PROCESSES) as pool:
            while not self._should_stop:
                try:
                    client_socket = self.__accept_new_connection()
                    pool.apply_async(self.handle_client_connection, args=(client_socket,))
                except OSError as e:
                    if self._should_stop:
                        # ignore error if socket was closed manually
                        logging.info(f'action: shutdown_gracefully | result: success')
                    else:
                        # server exits if there was a problem with the socket
                        logging.info(f'action: error_exiting | result: in_progress | error: {e}')
                    break


    def __obtain_all_batch_bets(self, client_socket: socket.socket, agency: int, data: bytes) -> list[Bet]:
        """
        Receives all bets from stream,

        If it doesn't receive all the information it should,
        then server replies with an error, stating how many
        bets were read appropriately.
        """
        with self._agencies_convar:
            if agency in self._agencies_done:
                raise ClientCannotSendMoreBetsError("Either lottery is closed, or agency anounced he was done before")

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
                msg = client_socket.recv(1024)
                if not msg:
                    logging.info(f'action: apuesta_recibida | result: fail | cantidad: {len(bets)}')
                    raise BetBatchError("There was an error parsing bet batch: couldn't get all required bets")
                data += msg  # Concatenar los datos recibidos al buffer existente
                
        return bets
    
    def __handle_bet_batch_message(self, client_socket: socket.socket, agency: int, data: bytes):
        bets = self.__obtain_all_batch_bets(client_socket, agency, data)
        if len(bets) > 0:
            with self.file_lock:
                store_bets(bets)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
            client_socket.sendall(ResponseStatus.OK.value.to_bytes(1, 'little'))

    def __handle_end_betting_message(self, client_socket: socket.socket, agency: int):
        """ 
        Adds agency to agencies that stated they were done.
        As adding an element to a set is idempotent, I chose to do the operation over and over again,
        and simply keep anouncing client it was already added.
        """
        with self._agencies_convar:
            if agency not in self._agencies_done:
                self._agencies_done.add(agency)
                logging.info(f'action: receive_end_bet | result: success | agencia: {agency}')
                if len(self._agencies_done) == AGENCY_CLOSING_NUMBER:
                    # just got to the num of required agencies, close bet
                    self._agencies_convar.notify_all()
                    logging.info("action: sorteo | result: success")

        client_socket.sendall(ResponseStatus.OK.value.to_bytes(1, 'little'))

    def __handle_request_winners(self, client_socket: socket.socket, agency: int):
        with self._agencies_convar:
            while len(self._agencies_done) < AGENCY_CLOSING_NUMBER:
                logging.info(f'action: wait_agencies | result: in_progress | agencia: {agency}')
                self._agencies_convar.wait()

        with self.file_lock: 
            if self.winners is None:
                self.winners = get_winners()
            winners_from_agency = get_bet_documents_from_agency(agency, self.winners)

        msg = ResponseStatus.SEND_WINNERS.value.to_bytes(1, 'little') + len(winners_from_agency).to_bytes(AGENCY_WINNERS_LEN, 'little')
        for winner in winners_from_agency:
            msg += int(winner).to_bytes(4, 'little')
        client_socket.sendall(msg)

    def __handle_message(self, client_socket: socket.socket, code: MessageCode, agency: int, data: bytes):
        if code == MessageCode.BET:
            self.__handle_bet_batch_message(client_socket, agency, data)
        elif code == MessageCode.END_BETTING:
            self.__handle_end_betting_message(client_socket, agency)
        elif code == MessageCode.REQUEST_WINNERS:
            self.__handle_request_winners(client_socket, agency)
        elif code == MessageCode.END_CONNECTION:
            client_socket.close()


    """ Read agency and message code from a specific client """
    def __read_header(self, client_socket: socket.socket) -> Tuple[MessageCode, int, bytes]:
        curr = 0
        # read as much as i can to avoid too many reads
        msg = client_socket.recv(1024)
        if not msg:
            raise NoMessageReceivedError("No message was received from socket: client probably ended connection before sending anything")
        
        try:
            agency = int.from_bytes([msg[curr]], 'little')
            curr += AGENCY_LEN

            msg_num = int.from_bytes([msg[curr]], 'little')
            msg_code = MessageCode(msg_num)
            curr += MSG_CODE_LEN
        except (IndexError, UnicodeDecodeError, ValueError) as e:
            raise WrongHeaderError(f"Error parsing message header: {e}")


        return msg_code, agency, msg[curr:]
    
    def __announce_error_with_code(self, client_socket: socket.socket, status: ResponseStatus, error: str):
        if self._should_stop:
            # client socket has already been closed somewhere else and should ignore error 
            return
        logging.info(f'action: receive_message | result: fail | error: {error}')

        try:
            client_socket.send(status.value.to_bytes(1, 'little'))
        except:
            logging.info(f'action: send_error | result: fail | error: broken pipe')

    def handle_client_connection(self, client_socket):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        while not self._should_stop:
            try:
                msg_code, agency, data = self.__read_header(client_socket)
                if msg_code == MessageCode.END_CONNECTION:
                    break
                self.__handle_message(client_socket, msg_code, agency, data)
            except (BetBatchError, WrongHeaderError) as e:
                # error was because either client closed connection or because he sent wrong batch information
                self.__announce_error_with_code(ResponseStatus.BAD_REQUEST, e)
                break
            except (OSError, csv.Error, NoMessageReceivedError) as e:
                self.__announce_error_with_code(ResponseStatus.ERROR, e)
                break
            except ClientCannotSendMoreBetsError as e:
                # give client the chance to redeem himself
                self.__announce_error_with_code(ResponseStatus.NO_MORE_BETS_ALLOWED, e)

        logging.info('action: close_connection | result: success')
        client_socket.close()

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
        return c
