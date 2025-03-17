CONFIRMATION = "CONFIRMATION"
ERROR = "ERROR"

MESSAGE_DELIMITER=":"

class Message:
    def __init__(self, message_type: str, content: str):
        self.message_type = message_type
        self.content = content

    def serialize(self) -> str:
        """
        Serialize the message to a string format.
        """
        return f"{self.message_type}{MESSAGE_DELIMITER}{self.content}\n"

    @staticmethod
    def deserialize(message: str):
        """
        Deserialize a string to a Message object.
        """
        try:
            message_type, content = message.split(MESSAGE_DELIMITER, 1)
            return Message(message_type, content)
        except ValueError as e:
            raise ValueError(f"Invalid message format: {message}") from e