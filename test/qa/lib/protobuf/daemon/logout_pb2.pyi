from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class LogoutRequest(_message.Message):
    __slots__ = ("revoke_token",)
    REVOKE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    revoke_token: bool
    def __init__(self, revoke_token: bool = ...) -> None: ...
