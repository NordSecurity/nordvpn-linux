from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class CitiesRequest(_message.Message):
    __slots__ = ("country",)
    COUNTRY_FIELD_NUMBER: _ClassVar[int]
    country: str
    def __init__(self, country: _Optional[str] = ...) -> None: ...
