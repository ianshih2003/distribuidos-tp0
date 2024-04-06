""" A lottery bet registry. """


import datetime
BET_MESSAGE_FIELD_SEPARATOR = "|"


class Bet:
    def __init__(self, agency: str, first_name: str, last_name: str, document: str, birthdate: str, number: str):
        """
        agency must be passed with integer format.
        birthdate must be passed with format: 'YYYY-MM-DD'.
        number must be passed with integer format.
        """
        self.agency = int(agency)
        self.first_name = first_name
        self.last_name = last_name
        self.document = document
        self.birthdate = datetime.date.fromisoformat(birthdate)
        self.number = int(number)

    @staticmethod
    def deserialize(message: bytes):
        bet_string = message.decode()

        agency, first_name, last_name, document, birthdate, number = bet_string.split(
            BET_MESSAGE_FIELD_SEPARATOR)

        return Bet(agency, first_name, last_name, document, birthdate, number)
