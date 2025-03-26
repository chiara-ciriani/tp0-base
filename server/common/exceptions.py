class InvalidMessageError(Exception):
    """Exception raised for invalid message types."""
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)