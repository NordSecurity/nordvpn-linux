from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class PingResponse(_message.Message):
    __slots__ = ("type", "major", "minor", "patch", "metadata")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    MAJOR_FIELD_NUMBER: _ClassVar[int]
    MINOR_FIELD_NUMBER: _ClassVar[int]
    PATCH_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    type: int
    major: int
    minor: int
    patch: int
    metadata: str
    def __init__(self, type: _Optional[int] = ..., major: _Optional[int] = ..., minor: _Optional[int] = ..., patch: _Optional[int] = ..., metadata: _Optional[str] = ...) -> None: ...
