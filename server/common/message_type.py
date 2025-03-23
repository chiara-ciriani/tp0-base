from enum import Enum

class MessageType(Enum):
    BATCH = 0
    BATCH_END = 1
    WINNERS_REQUEST = 2