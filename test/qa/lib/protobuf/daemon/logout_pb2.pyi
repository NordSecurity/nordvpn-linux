from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class LogoutRequest(_message.Message):
    __slots__ = ("persist_token",)
    PERSIST_TOKEN_FIELD_NUMBER: _ClassVar[int]
    persist_token: bool
    def __init__(self, persist_token: bool = ...) -> None: ...
