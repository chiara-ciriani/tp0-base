from enum import Enum

class ResponseStatus(Enum):
    OK = 0
    ERROR = 1
    SEND_WINNERS = 2
    BET_NOT_FINISHED = 3