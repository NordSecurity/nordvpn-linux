from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class Technology(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UNKNOWN_TECHNOLOGY: _ClassVar[Technology]
    OPENVPN: _ClassVar[Technology]
    NORDLYNX: _ClassVar[Technology]
    NORDWHISPER: _ClassVar[Technology]
UNKNOWN_TECHNOLOGY: Technology
OPENVPN: Technology
NORDLYNX: Technology
NORDWHISPER: Technology
