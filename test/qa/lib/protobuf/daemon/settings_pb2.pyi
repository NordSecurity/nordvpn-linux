import common_pb2 as _common_pb2
from config import technology_pb2 as _technology_pb2
from config import analytics_consent_pb2 as _analytics_consent_pb2
from config import protocol_pb2 as _protocol_pb2
from config import group_pb2 as _group_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SettingsResponse(_message.Message):
    __slots__ = ("type", "data")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    type: int
    data: Settings
    def __init__(self, type: _Optional[int] = ..., data: _Optional[_Union[Settings, _Mapping]] = ...) -> None: ...

class AutoconnectData(_message.Message):
    __slots__ = ("enabled", "country", "city", "server_group")
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    COUNTRY_FIELD_NUMBER: _ClassVar[int]
    CITY_FIELD_NUMBER: _ClassVar[int]
    SERVER_GROUP_FIELD_NUMBER: _ClassVar[int]
    enabled: bool
    country: str
    city: str
    server_group: _group_pb2.ServerGroup
    def __init__(self, enabled: bool = ..., country: _Optional[str] = ..., city: _Optional[str] = ..., server_group: _Optional[_Union[_group_pb2.ServerGroup, str]] = ...) -> None: ...

class Settings(_message.Message):
    __slots__ = ("technology", "firewall", "kill_switch", "auto_connect_data", "meshnet", "routing", "fwmark", "analytics_consent", "dns", "threat_protection_lite", "protocol", "lan_discovery", "allowlist", "obfuscate", "virtualLocation", "postquantum_vpn", "user_settings")
    TECHNOLOGY_FIELD_NUMBER: _ClassVar[int]
    FIREWALL_FIELD_NUMBER: _ClassVar[int]
    KILL_SWITCH_FIELD_NUMBER: _ClassVar[int]
    AUTO_CONNECT_DATA_FIELD_NUMBER: _ClassVar[int]
    MESHNET_FIELD_NUMBER: _ClassVar[int]
    ROUTING_FIELD_NUMBER: _ClassVar[int]
    FWMARK_FIELD_NUMBER: _ClassVar[int]
    ANALYTICS_CONSENT_FIELD_NUMBER: _ClassVar[int]
    DNS_FIELD_NUMBER: _ClassVar[int]
    THREAT_PROTECTION_LITE_FIELD_NUMBER: _ClassVar[int]
    PROTOCOL_FIELD_NUMBER: _ClassVar[int]
    LAN_DISCOVERY_FIELD_NUMBER: _ClassVar[int]
    ALLOWLIST_FIELD_NUMBER: _ClassVar[int]
    OBFUSCATE_FIELD_NUMBER: _ClassVar[int]
    VIRTUALLOCATION_FIELD_NUMBER: _ClassVar[int]
    POSTQUANTUM_VPN_FIELD_NUMBER: _ClassVar[int]
    USER_SETTINGS_FIELD_NUMBER: _ClassVar[int]
    technology: _technology_pb2.Technology
    firewall: bool
    kill_switch: bool
    auto_connect_data: AutoconnectData
    meshnet: bool
    routing: bool
    fwmark: int
    analytics_consent: _analytics_consent_pb2.ConsentMode
    dns: _containers.RepeatedScalarFieldContainer[str]
    threat_protection_lite: bool
    protocol: _protocol_pb2.Protocol
    lan_discovery: bool
    allowlist: _common_pb2.Allowlist
    obfuscate: bool
    virtualLocation: bool
    postquantum_vpn: bool
    user_settings: UserSpecificSettings
    def __init__(self, technology: _Optional[_Union[_technology_pb2.Technology, str]] = ..., firewall: bool = ..., kill_switch: bool = ..., auto_connect_data: _Optional[_Union[AutoconnectData, _Mapping]] = ..., meshnet: bool = ..., routing: bool = ..., fwmark: _Optional[int] = ..., analytics_consent: _Optional[_Union[_analytics_consent_pb2.ConsentMode, str]] = ..., dns: _Optional[_Iterable[str]] = ..., threat_protection_lite: bool = ..., protocol: _Optional[_Union[_protocol_pb2.Protocol, str]] = ..., lan_discovery: bool = ..., allowlist: _Optional[_Union[_common_pb2.Allowlist, _Mapping]] = ..., obfuscate: bool = ..., virtualLocation: bool = ..., postquantum_vpn: bool = ..., user_settings: _Optional[_Union[UserSpecificSettings, _Mapping]] = ...) -> None: ...

class UserSpecificSettings(_message.Message):
    __slots__ = ("uid", "notify", "tray")
    UID_FIELD_NUMBER: _ClassVar[int]
    NOTIFY_FIELD_NUMBER: _ClassVar[int]
    TRAY_FIELD_NUMBER: _ClassVar[int]
    uid: int
    notify: bool
    tray: bool
    def __init__(self, uid: _Optional[int] = ..., notify: bool = ..., tray: bool = ...) -> None: ...
