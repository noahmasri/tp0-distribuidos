from enum import Enum

"""Defines different message types a client msg can be"""
class MessageCode(Enum):
    BET = 0
    END_BETTING = 1
    REQUEST_WINNERS = 2