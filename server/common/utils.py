import csv
import datetime
import socket
import time
from typing import Tuple

from common.errors import ShouldReadStreamError

""" Bets storage location. """
STORAGE_FILEPATH = "./bets.csv"
""" Simulated winner number in the lottery contest. """
LOTTERY_WINNER_NUMBER = 7574

"""lengths in bytes"""
AGENCY_LEN = 1
NAME_LEN = 1 # bytes for name lenght
SURNAME_LEN = 1 # bytes for name lenght
ID_LEN = 4
BIRTHDATE_LEN=10
NUMBER_LEN = 2

""" A lottery bet registry. """
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
    def deserialize(agency: int, data: bytes) -> Tuple['Bet', bytes]:
        curr = 0
        try:
            name_len = int.from_bytes([data[curr]], 'little')
            curr+=NAME_LEN
            name = data[curr:name_len+curr].decode()
            curr+=name_len

            surname_len = int.from_bytes([data[curr]], 'little')
            curr+=SURNAME_LEN
            surname = data[curr:surname_len+curr].decode()
            curr+=surname_len
            document=int.from_bytes(data[curr:ID_LEN+curr], 'little')
            curr+=ID_LEN

            birthdate=data[curr:BIRTHDATE_LEN+curr].decode()
            curr+=BIRTHDATE_LEN

            number = int.from_bytes(data[curr:curr+NUMBER_LEN], 'little')
            curr+= NUMBER_LEN
        except (IndexError, UnicodeDecodeError):
            raise ShouldReadStreamError("Must read more bytes from the stream to keep on creating bets")
        
        if len(data) < curr:
            # necessary check, as slicing does not throw error
            raise ShouldReadStreamError("Must read more bytes from the stream to keep on creating bets")

        return Bet(agency, name, surname, document, birthdate, number), data[curr:]
    
    def __repr__(self):
        return (f"Bet(agency={self.agency}, first name={self.first_name}, last name={self.last_name}, "
                f"document={self.document}, birthdate={self.birthdate}, number={self.number})")

    def __is_from_agency(self, agency: int):
        return self.agency == agency
    

""" Checks whether a bet won the prize or not. """
def has_won(bet: Bet) -> bool:
    return bet.number == LOTTERY_WINNER_NUMBER

def get_bet_documents_from_agency(agency: int, all_bets: list[Bet]) -> map:
    return list(map(lambda bet: bet.document, filter(lambda bet: bet.__is_from_agency(agency), all_bets)))

def get_winners() -> list[Bet]:
    return list(filter(has_won, load_bets()))

"""
Persist the information of each bet in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def store_bets(bets: list[Bet]) -> None:
    with open(STORAGE_FILEPATH, 'a+') as file:
        writer = csv.writer(file, quoting=csv.QUOTE_MINIMAL)
        for bet in bets:
            writer.writerow([bet.agency, bet.first_name, bet.last_name,
                             bet.document, bet.birthdate, bet.number])

"""
Loads the information all the bets in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def load_bets() -> list[Bet]:
    with open(STORAGE_FILEPATH, 'r') as file:
        reader = csv.reader(file, quoting=csv.QUOTE_MINIMAL)
        for row in reader:
            yield Bet(row[0], row[1], row[2], row[3], row[4], row[5])

