from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class ConnectRequest(_message.Message):
    __slots__ = ("server_tag", "server_group")
    SERVER_TAG_FIELD_NUMBER: _ClassVar[int]
    SERVER_GROUP_FIELD_NUMBER: _ClassVar[int]
    server_tag: str
    server_group: str
    def __init__(self, server_tag: _Optional[str] = ..., server_group: _Optional[str] = ...) -> None: ...
