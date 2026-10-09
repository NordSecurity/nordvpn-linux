from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class VPNProtocol(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    VPN_PROTOCOL_UNSPECIFIED: _ClassVar[VPNProtocol]
    VPN_PROTOCOL_NORDLYNX: _ClassVar[VPNProtocol]
    VPN_PROTOCOL_OPENVPN_UDP: _ClassVar[VPNProtocol]
    VPN_PROTOCOL_OPENVPN_TCP: _ClassVar[VPNProtocol]
    VPN_PROTOCOL_NORDWHISPER: _ClassVar[VPNProtocol]
VPN_PROTOCOL_UNSPECIFIED: VPNProtocol
VPN_PROTOCOL_NORDLYNX: VPNProtocol
VPN_PROTOCOL_OPENVPN_UDP: VPNProtocol
VPN_PROTOCOL_OPENVPN_TCP: VPNProtocol
VPN_PROTOCOL_NORDWHISPER: VPNProtocol
