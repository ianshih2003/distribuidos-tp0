import socket
import logging
import signal
from common.utils import process_incoming_message

MAX_MESSAGE_BYTES = 4


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        signal.signal(signal.SIGTERM, lambda signal, frame: self.stop())
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.client_sock = None

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        # TODO: Modify this program to handle signal to graceful shutdown
        # the server
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
        except:
            self.__send_error_message()
            logging.error(
                f"action: receive_message_length | result: failed | error: cant convert to int")
            return 0

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            msg_length = self.__receive_message_length()

            if msg_length == 0:
                return

            msg = self.__safe_receive(msg_length).rstrip()
            addr = self.client_sock.getpeername()
            logging.info(
                f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')

            process_incoming_message(msg)

            self.__send_success_message()
        except OSError as e:
            self.__send_error_message()
            logging.error(
                f"action: receive_message | result: fail | error: {e}")
        finally:
            self._close_client_socket()

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
            self.client_sock.close()
            self.client_sock = None
        logging.info('action: closing client socket | result: success')

    def __send_success_message(self):
        self.__safe_send("suc")
        logging.info('action: send sucess message | result: success')

    def __send_error_message(self):
        self.__safe_send("err")
        logging.info('action: send error message | result: success')

    def __safe_send(self, message):
        total_sent = 0
        bytes_to_send = message.encode()

        while total_sent < len(message):
            n = self.client_sock.send(bytes_to_send[total_sent:])
            total_sent += n
        return

    def __safe_receive(self, buffer_length):

        n = 0

        buffer = bytes()
        while n < buffer_length:
            message = self.client_sock.recv(buffer_length)
            buffer += message
            n += len(message)
        return buffer
