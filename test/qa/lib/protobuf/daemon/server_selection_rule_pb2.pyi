from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class ServerSelectionRule(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    NONE: _ClassVar[ServerSelectionRule]
    RECOMMENDED: _ClassVar[ServerSelectionRule]
    CITY: _ClassVar[ServerSelectionRule]
    COUNTRY: _ClassVar[ServerSelectionRule]
    SPECIFIC_SERVER: _ClassVar[ServerSelectionRule]
    GROUP: _ClassVar[ServerSelectionRule]
    COUNTRY_WITH_GROUP: _ClassVar[ServerSelectionRule]
    SPECIFIC_SERVER_WITH_GROUP: _ClassVar[ServerSelectionRule]
NONE: ServerSelectionRule
RECOMMENDED: ServerSelectionRule
CITY: ServerSelectionRule
COUNTRY: ServerSelectionRule
SPECIFIC_SERVER: ServerSelectionRule
GROUP: ServerSelectionRule
COUNTRY_WITH_GROUP: ServerSelectionRule
SPECIFIC_SERVER_WITH_GROUP: ServerSelectionRule
