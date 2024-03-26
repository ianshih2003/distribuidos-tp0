import socket
import logging
import signal
from common.utils import load_bets, process_bets, get_winner_bets_by_agency

MAX_MESSAGE_BYTES = 4
TIMEOUT = 1.0
EXIT = "exit"
WINNERS = "winners"


class Server:
    def __init__(self, port, listen_backlog, clients):
        # Initialize server socket
        signal.signal(signal.SIGTERM, lambda signal, frame: self.stop())
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.client_sock = None

        self.clients = clients

        self.finished_clients = 0

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while True:
            try:
                client_sock = self.__accept_new_connection()
                self.client_sock = client_sock
                self.__handle_client_connection()
            except OSError:
                break

    def __receive_message_length(self):
        try:

            msg_length = int.from_bytes(self.__safe_receive(
                MAX_MESSAGE_BYTES).rstrip(), "little")

            logging.info(
                f"action: receive_message_length | result: success | length: {msg_length}")

            self.__send_success_message()

            return msg_length
        except socket.error as e:
            logging.error(
                f"action: receive_message_length | result: failed | error: client disconnected")
            raise e
        except Exception as e:
            self.__send_error_message()
            logging.error(
                f"action: receive_message_length | result: failed | error: {e}")
            return 0

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        while True:
            try:
                msg_length = self.__receive_message_length()

                if msg_length == 0:
                    return

                msg = self.__safe_receive(msg_length).rstrip()
                self.__log_ip()

                self.__process_message(msg)

            except OSError as e:
                self.__send_error_message()
                logging.error(
                    f"action: receive_message | result: fail | error: {e}")
                break
            except:
                break
        self._close_client_socket()

    def __log_ip(self):
        addr = self.client_sock.getpeername()
        logging.info(
            f'action: receive_message | result: success | ip: {addr[0]}')

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(
            f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def stop(self):
        logging.info(
            'action: receive_termination_signal | result: in_progress')

        logging.info('action: closing listening socket | result: in_progress')
        self._server_socket.close()
        logging.info('action: closing listening socket | result: success')

        self._close_client_socket()

        logging.info(
            f'action: receive_termination_signal | result: success')

    def _close_client_socket(self):
        logging.info('action: closing client socket | result: in_progress')
        if self.client_sock:

            self.finished_clients += 1

            self.client_sock.close()

            self.client_sock = None
        logging.info('action: closing client socket | result: success')

    def __send_success_message(self):
        self.__safe_send("suc".encode())
        logging.info('action: send sucess message | result: success')

    def __send_error_message(self):
        self.__safe_send("err".encode())
        logging.error('action: send error message | result: success')

    def __safe_send(self, message: bytes):
        n = 0
        max_tries = 5

        while n != len(message) and max_tries > 0:
            n = self.client_sock.send(message)
            max_tries -= 1
        return

    def __safe_receive(self, buffer_length: int):
        n = 0

        buffer = bytes()
        while n < buffer_length:
            message = self.client_sock.recv(buffer_length)
            buffer += message
            n += len(message)

        return buffer

    def __process_message(self, message: bytes):
        msg = message.decode()

        split_msg = msg.split(",")
        logging.info(split_msg)
        if msg == EXIT:
            raise socket.error("Client disconnected")
        elif len(split_msg) == 2 and split_msg[0] == WINNERS:
            self.__send_winners(split_msg[1])
        else:
            process_bets(msg)
            self.__send_success_message()

    def __send_winners(self, agency: str):
        if self.finished_clients < self.clients:
            self.__send("waiting".encode())
            return

        bets = load_bets()

        winner_bets = get_winner_bets_by_agency(bets, agency)

        dnis = map(lambda bet: bet.document, winner_bets)

        response = ",".join(dnis)

        self.__send(response.encode())

    def __send(self, message: bytes):
        self.__safe_send(len(message).to_bytes(MAX_MESSAGE_BYTES, 'little'))

        if self.__safe_receive(3).decode() != "suc":
            raise socket.error("rejected")

        self.__safe_send(message)

        if self.__safe_receive(3).decode() != "suc":
            raise socket.error("rejected")
