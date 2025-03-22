class ClientClosedConnection(Exception):
    """Exception raised when the client closes the connection."""
    def __init__(self, message="Connection closed by client"):
        self.message = message
        super().__init__(self.message)

class InvalidMessageError(Exception):
    """Exception raised for invalid message types."""
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)