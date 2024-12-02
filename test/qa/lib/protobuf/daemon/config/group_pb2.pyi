from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class ServerGroup(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UNDEFINED: _ClassVar[ServerGroup]
    DoubleVPN: _ClassVar[ServerGroup]
    ONION_OVER_VPN: _ClassVar[ServerGroup]
    ULTRA_FAST_TV: _ClassVar[ServerGroup]
    ANTI_DDOS: _ClassVar[ServerGroup]
    DEDICATED_IP: _ClassVar[ServerGroup]
    STANDARD_VPN_SERVERS: _ClassVar[ServerGroup]
    NETFLIX_USA: _ClassVar[ServerGroup]
    P2P: _ClassVar[ServerGroup]
    OBFUSCATED: _ClassVar[ServerGroup]
    EUROPE: _ClassVar[ServerGroup]
    THE_AMERICAS: _ClassVar[ServerGroup]
    ASIA_PACIFIC: _ClassVar[ServerGroup]
    AFRICA_MIDDLE_EAST_INDIA: _ClassVar[ServerGroup]
UNDEFINED: ServerGroup
DoubleVPN: ServerGroup
ONION_OVER_VPN: ServerGroup
ULTRA_FAST_TV: ServerGroup
ANTI_DDOS: ServerGroup
DEDICATED_IP: ServerGroup
STANDARD_VPN_SERVERS: ServerGroup
NETFLIX_USA: ServerGroup
P2P: ServerGroup
OBFUSCATED: ServerGroup
EUROPE: ServerGroup
THE_AMERICAS: ServerGroup
ASIA_PACIFIC: ServerGroup
AFRICA_MIDDLE_EAST_INDIA: ServerGroup
