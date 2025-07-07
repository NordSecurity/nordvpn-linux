from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class SetDefaultsRequest(_message.Message):
    __slots__ = ("no_logout", "off_killswitch")
    NO_LOGOUT_FIELD_NUMBER: _ClassVar[int]
    OFF_KILLSWITCH_FIELD_NUMBER: _ClassVar[int]
    no_logout: bool
    off_killswitch: bool
    def __init__(self, no_logout: bool = ..., off_killswitch: bool = ...) -> None: ...
