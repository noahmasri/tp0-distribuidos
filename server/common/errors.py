class ShouldReadStreamError(Exception):
    def __init__(self, message):
        super().__init__(message)

class BetBatchError(Exception):
    def __init__(self, message):
        super().__init__(message)