import csv
import datetime
import logging


""" Bets storage location. """
STORAGE_FILEPATH = "./bets.csv"
""" Simulated winner number in the lottery contest. """
LOTTERY_WINNER_NUMBER = 7574

FIRST_NAME_LENGTH_LEN = 1
LAST_NAME_LENGTH_LEN = 1
DOCUMENT_LEN = 4
BIRTHDATE_LEN = 10
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

    def get_document(self) -> str:
        return self.document

    def get_number(self) -> int:
        return self.number
    
    @staticmethod
    def deserialize(agency_id: int, bet_info: bytes, index: int) -> 'Bet':
        start_index = index
        
        # First name
        first_name_length = int.from_bytes([bet_info[index]], 'big')
        index += FIRST_NAME_LENGTH_LEN
        first_name = bet_info[index:index+first_name_length].decode('utf-8')
        index += first_name_length
        
        # Last name
        last_name_length = int.from_bytes([bet_info[index]], 'big')
        index += LAST_NAME_LENGTH_LEN
        last_name = bet_info[index:index+last_name_length].decode('utf-8')
        index += last_name_length

        # Document
        document = int.from_bytes(bet_info[index:index+DOCUMENT_LEN], 'big')
        index += DOCUMENT_LEN

        # Birthdate
        birthdate = bet_info[index:index+BIRTHDATE_LEN].decode('utf-8')
        index += BIRTHDATE_LEN

        # Number
        number = int.from_bytes(bet_info[index:index+NUMBER_LEN], 'big')
        index += NUMBER_LEN

        bet = Bet(agency_id, first_name, last_name, document, birthdate, number)
        bytes_deserialized = index - start_index
        return bet, bytes_deserialized

""" Checks whether a bet won the prize or not. """
def has_won(bet: Bet) -> bool:
    return bet.number == LOTTERY_WINNER_NUMBER

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

