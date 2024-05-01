from multiprocessing import Lock, Process, Value
import socket
import logging
import signal

from common.client_handler import create_client_handler

MAX_MESSAGE_BYTES = 4
EXIT = "exit"
WINNERS = "winners"
CONFIRMATION_MSG_LENGTH = 3
SUCCESS_MSG = "suc"


class Server:
    def __init__(self, port, listen_backlog, clients):
        # Initialize server socket
        signal.signal(signal.SIGTERM, lambda signal, frame: self.stop())
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

        self.clients = clients

        self.finished_clients = Value('i', 0)

        self.file_lock = Lock()

        self.processes = []

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

                process = Process(target=create_client_handler,
                                  args=(client_sock, self.file_lock, self.finished_clients, self.clients))

                self.processes.append(process)
                process.start()
            except OSError:
                break

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
        self._server_socket.shutdown(socket.SHUT_RDWR)
        self._server_socket.close()
        logging.info('action: closing listening socket | result: success')

        for process in self.processes:
            process.terminate()

        logging.info(
            f'action: receive_termination_signal | result: success')
