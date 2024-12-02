from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class RateRequest(_message.Message):
    __slots__ = ("rating",)
    RATING_FIELD_NUMBER: _ClassVar[int]
    rating: int
    def __init__(self, rating: _Optional[int] = ...) -> None: ...
