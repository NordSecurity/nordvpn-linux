from config import protocol_pb2 as _protocol_pb2
from config import technology_pb2 as _technology_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SetErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    FAILURE: _ClassVar[SetErrorCode]
    CONFIG_ERROR: _ClassVar[SetErrorCode]
    ALREADY_SET: _ClassVar[SetErrorCode]

class SetThreatProtectionLiteStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    TPL_CONFIGURED: _ClassVar[SetThreatProtectionLiteStatus]
    TPL_CONFIGURED_DNS_RESET: _ClassVar[SetThreatProtectionLiteStatus]

class SetDNSStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    DNS_CONFIGURED: _ClassVar[SetDNSStatus]
    DNS_CONFIGURED_TPL_RESET: _ClassVar[SetDNSStatus]
    INVALID_DNS_ADDRESS: _ClassVar[SetDNSStatus]
    TOO_MANY_VALUES: _ClassVar[SetDNSStatus]

class SetProtocolStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PROTOCOL_CONFIGURED: _ClassVar[SetProtocolStatus]
    PROTOCOL_CONFIGURED_VPN_ON: _ClassVar[SetProtocolStatus]
    INVALID_TECHNOLOGY: _ClassVar[SetProtocolStatus]

class SetLANDiscoveryStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    DISCOVERY_CONFIGURED: _ClassVar[SetLANDiscoveryStatus]
    DISCOVERY_CONFIGURED_ALLOWLIST_RESET: _ClassVar[SetLANDiscoveryStatus]
FAILURE: SetErrorCode
CONFIG_ERROR: SetErrorCode
ALREADY_SET: SetErrorCode
TPL_CONFIGURED: SetThreatProtectionLiteStatus
TPL_CONFIGURED_DNS_RESET: SetThreatProtectionLiteStatus
DNS_CONFIGURED: SetDNSStatus
DNS_CONFIGURED_TPL_RESET: SetDNSStatus
INVALID_DNS_ADDRESS: SetDNSStatus
TOO_MANY_VALUES: SetDNSStatus
PROTOCOL_CONFIGURED: SetProtocolStatus
PROTOCOL_CONFIGURED_VPN_ON: SetProtocolStatus
INVALID_TECHNOLOGY: SetProtocolStatus
DISCOVERY_CONFIGURED: SetLANDiscoveryStatus
DISCOVERY_CONFIGURED_ALLOWLIST_RESET: SetLANDiscoveryStatus

class SetAutoconnectRequest(_message.Message):
    __slots__ = ("enabled", "server_tag")
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    SERVER_TAG_FIELD_NUMBER: _ClassVar[int]
    enabled: bool
    server_tag: str
    def __init__(self, enabled: bool = ..., server_tag: _Optional[str] = ...) -> None: ...

class SetGenericRequest(_message.Message):
    __slots__ = ("enabled",)
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    enabled: bool
    def __init__(self, enabled: bool = ...) -> None: ...

class SetUint32Request(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: int
    def __init__(self, value: _Optional[int] = ...) -> None: ...

class SetThreatProtectionLiteRequest(_message.Message):
    __slots__ = ("threat_protection_lite",)
    THREAT_PROTECTION_LITE_FIELD_NUMBER: _ClassVar[int]
    threat_protection_lite: bool
    def __init__(self, threat_protection_lite: bool = ...) -> None: ...

class SetThreatProtectionLiteResponse(_message.Message):
    __slots__ = ("error_code", "set_threat_protection_lite_status")
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    SET_THREAT_PROTECTION_LITE_STATUS_FIELD_NUMBER: _ClassVar[int]
    error_code: SetErrorCode
    set_threat_protection_lite_status: SetThreatProtectionLiteStatus
    def __init__(self, error_code: _Optional[_Union[SetErrorCode, str]] = ..., set_threat_protection_lite_status: _Optional[_Union[SetThreatProtectionLiteStatus, str]] = ...) -> None: ...

class SetDNSRequest(_message.Message):
    __slots__ = ("dns", "threat_protection_lite")
    DNS_FIELD_NUMBER: _ClassVar[int]
    THREAT_PROTECTION_LITE_FIELD_NUMBER: _ClassVar[int]
    dns: _containers.RepeatedScalarFieldContainer[str]
    threat_protection_lite: bool
    def __init__(self, dns: _Optional[_Iterable[str]] = ..., threat_protection_lite: bool = ...) -> None: ...

class SetDNSResponse(_message.Message):
    __slots__ = ("error_code", "set_dns_status")
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    SET_DNS_STATUS_FIELD_NUMBER: _ClassVar[int]
    error_code: SetErrorCode
    set_dns_status: SetDNSStatus
    def __init__(self, error_code: _Optional[_Union[SetErrorCode, str]] = ..., set_dns_status: _Optional[_Union[SetDNSStatus, str]] = ...) -> None: ...

class SetKillSwitchRequest(_message.Message):
    __slots__ = ("kill_switch",)
    KILL_SWITCH_FIELD_NUMBER: _ClassVar[int]
    kill_switch: bool
    def __init__(self, kill_switch: bool = ...) -> None: ...

class SetNotifyRequest(_message.Message):
    __slots__ = ("uid", "notify")
    UID_FIELD_NUMBER: _ClassVar[int]
    NOTIFY_FIELD_NUMBER: _ClassVar[int]
    uid: int
    notify: bool
    def __init__(self, uid: _Optional[int] = ..., notify: bool = ...) -> None: ...

class SetTrayRequest(_message.Message):
    __slots__ = ("uid", "tray")
    UID_FIELD_NUMBER: _ClassVar[int]
    TRAY_FIELD_NUMBER: _ClassVar[int]
    uid: int
    tray: bool
    def __init__(self, uid: _Optional[int] = ..., tray: bool = ...) -> None: ...

class SetProtocolRequest(_message.Message):
    __slots__ = ("protocol",)
    PROTOCOL_FIELD_NUMBER: _ClassVar[int]
    protocol: _protocol_pb2.Protocol
    def __init__(self, protocol: _Optional[_Union[_protocol_pb2.Protocol, str]] = ...) -> None: ...

class SetProtocolResponse(_message.Message):
    __slots__ = ("error_code", "set_protocol_status")
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    SET_PROTOCOL_STATUS_FIELD_NUMBER: _ClassVar[int]
    error_code: SetErrorCode
    set_protocol_status: SetProtocolStatus
    def __init__(self, error_code: _Optional[_Union[SetErrorCode, str]] = ..., set_protocol_status: _Optional[_Union[SetProtocolStatus, str]] = ...) -> None: ...

class SetTechnologyRequest(_message.Message):
    __slots__ = ("technology",)
    TECHNOLOGY_FIELD_NUMBER: _ClassVar[int]
    technology: _technology_pb2.Technology
    def __init__(self, technology: _Optional[_Union[_technology_pb2.Technology, str]] = ...) -> None: ...

class PortRange(_message.Message):
    __slots__ = ("start_port", "end_port")
    START_PORT_FIELD_NUMBER: _ClassVar[int]
    END_PORT_FIELD_NUMBER: _ClassVar[int]
    start_port: int
    end_port: int
    def __init__(self, start_port: _Optional[int] = ..., end_port: _Optional[int] = ...) -> None: ...

class SetAllowlistSubnetRequest(_message.Message):
    __slots__ = ("subnet",)
    SUBNET_FIELD_NUMBER: _ClassVar[int]
    subnet: str
    def __init__(self, subnet: _Optional[str] = ...) -> None: ...

class SetAllowlistPortsRequest(_message.Message):
    __slots__ = ("is_udp", "is_tcp", "port_range")
    IS_UDP_FIELD_NUMBER: _ClassVar[int]
    IS_TCP_FIELD_NUMBER: _ClassVar[int]
    PORT_RANGE_FIELD_NUMBER: _ClassVar[int]
    is_udp: bool
    is_tcp: bool
    port_range: PortRange
    def __init__(self, is_udp: bool = ..., is_tcp: bool = ..., port_range: _Optional[_Union[PortRange, _Mapping]] = ...) -> None: ...

class SetAllowlistRequest(_message.Message):
    __slots__ = ("set_allowlist_subnet_request", "set_allowlist_ports_request")
    SET_ALLOWLIST_SUBNET_REQUEST_FIELD_NUMBER: _ClassVar[int]
    SET_ALLOWLIST_PORTS_REQUEST_FIELD_NUMBER: _ClassVar[int]
    set_allowlist_subnet_request: SetAllowlistSubnetRequest
    set_allowlist_ports_request: SetAllowlistPortsRequest
    def __init__(self, set_allowlist_subnet_request: _Optional[_Union[SetAllowlistSubnetRequest, _Mapping]] = ..., set_allowlist_ports_request: _Optional[_Union[SetAllowlistPortsRequest, _Mapping]] = ...) -> None: ...

class SetLANDiscoveryRequest(_message.Message):
    __slots__ = ("enabled",)
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    enabled: bool
    def __init__(self, enabled: bool = ...) -> None: ...

class SetLANDiscoveryResponse(_message.Message):
    __slots__ = ("error_code", "set_lan_discovery_status")
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    SET_LAN_DISCOVERY_STATUS_FIELD_NUMBER: _ClassVar[int]
    error_code: SetErrorCode
    set_lan_discovery_status: SetLANDiscoveryStatus
    def __init__(self, error_code: _Optional[_Union[SetErrorCode, str]] = ..., set_lan_discovery_status: _Optional[_Union[SetLANDiscoveryStatus, str]] = ...) -> None: ...
