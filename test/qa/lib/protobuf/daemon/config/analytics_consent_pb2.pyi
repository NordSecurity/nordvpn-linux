from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class ConsentMode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    NONE: _ClassVar[ConsentMode]
    ALLOWED: _ClassVar[ConsentMode]
    FORBIDDEN: _ClassVar[ConsentMode]
NONE: ConsentMode
ALLOWED: ConsentMode
FORBIDDEN: ConsentMode
