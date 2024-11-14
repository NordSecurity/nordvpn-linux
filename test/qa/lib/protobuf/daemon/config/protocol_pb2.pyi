from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class Protocol(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UNKNOWN_PROTOCOL: _ClassVar[Protocol]
    UDP: _ClassVar[Protocol]
    TCP: _ClassVar[Protocol]
UNKNOWN_PROTOCOL: Protocol
UDP: Protocol
TCP: Protocol
