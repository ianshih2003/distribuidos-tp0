import socket
import logging
import signal
from common.utils import decode_utf8, encode_string_utf8, load_bets, process_bets, get_winner_bets_by_agency


MAX_MESSAGE_BYTES = 4
EXIT = "exit"
WINNERS = "winners"
CONFIRMATION_MSG_LENGTH = 3
SUCCESS_MSG = "suc"
ERROR_MSG = "err"
WAITING_MSG = "waiting"


class ClientHandler:
    def __init__(self, client_sock, file_lock, finished_clients, clients):
        # Initialize server socket
        signal.signal(signal.SIGTERM, lambda signal, frame: self.stop())

        self.file_lock = file_lock

        self.client_sock = client_sock

        self.finished_clients = finished_clients

        self.clients = clients

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

    def handle_client_connection(self):
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
            except Exception as e:
                logging.error(
                    f"action: any | result: fail | error: {e}")
                break
        self._close_client_socket()

    def __log_ip(self):
        addr = self.client_sock.getpeername()
        logging.info(
            f'action: receive_message | result: success | ip: {addr[0]}')

    def stop(self):
        logging.info(
            'action: receive_termination_signal | result: in_progress')
        self._close_client_socket()
        logging.info(
            f'action: receive_termination_signal | result: success')

    def _close_client_socket(self):
        logging.info('action: closing client socket | result: in_progress')
        with self.finished_clients.get_lock():
            self.finished_clients.value += 1

        self.client_sock.close()
        logging.info('action: closing client socket | result: success')

    def __send_success_message(self):
        self.__safe_send(encode_string_utf8(SUCCESS_MSG))
        logging.info('action: send sucess message | result: success')

    def __send_error_message(self):
        self.__safe_send(encode_string_utf8(ERROR_MSG))
        logging.error('action: send error message | result: success')

    def __safe_send(self, bytes_to_send):
        total_sent = 0

        while total_sent < len(bytes_to_send):
            n = self.client_sock.send(bytes_to_send[total_sent:])
            total_sent += n
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
        msg = decode_utf8(message)

        split_msg = msg.split(",")
        if msg == EXIT:
            raise socket.error("Client disconnected")
        elif len(split_msg) == 2 and split_msg[0] == WINNERS:
            self.__send_winners(split_msg[1])
        else:
            process_bets(msg)
            self.__send_success_message()

    def __send_winners(self, agency: str):
        if self.finished_clients.value < self.clients:
            self.__send_and_wait_confirmation(encode_string_utf8(WAITING_MSG))
            return

        bets = load_bets()

        winner_bets = get_winner_bets_by_agency(bets, agency)

        dnis = map(lambda bet: bet.document, winner_bets)

        response = ",".join(dnis)

        self.__send_and_wait_confirmation(encode_string_utf8(response))

    def __send_and_wait_confirmation(self, message: bytes):

        self.__safe_send(len(message).to_bytes(MAX_MESSAGE_BYTES, 'little'))

        if decode_utf8(self.__safe_receive(CONFIRMATION_MSG_LENGTH)) != SUCCESS_MSG:
            raise socket.error("rejected")

        self.__safe_send(message)

        if decode_utf8(self.__safe_receive(CONFIRMATION_MSG_LENGTH)) != SUCCESS_MSG:
            raise socket.error("rejected")


def create_client_handler(client_socket, file_lock, finished_clients, clients):
    c_handler = ClientHandler(client_socket, file_lock,
                              finished_clients, clients)

    c_handler.handle_client_connection()