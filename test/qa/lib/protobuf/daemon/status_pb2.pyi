from config import protocol_pb2 as _protocol_pb2
from config import technology_pb2 as _technology_pb2
from config import group_pb2 as _group_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ConnectionSource(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UNKNOWN_SOURCE: _ClassVar[ConnectionSource]
    MANUAL: _ClassVar[ConnectionSource]
    AUTO: _ClassVar[ConnectionSource]

class ConnectionState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    DISCONNECTED: _ClassVar[ConnectionState]
    CONNECTING: _ClassVar[ConnectionState]
    CONNECTED: _ClassVar[ConnectionState]
UNKNOWN_SOURCE: ConnectionSource
MANUAL: ConnectionSource
AUTO: ConnectionSource
DISCONNECTED: ConnectionState
CONNECTING: ConnectionState
CONNECTED: ConnectionState

class ConnectionParameters(_message.Message):
    __slots__ = ("source", "country", "city", "group")
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    COUNTRY_FIELD_NUMBER: _ClassVar[int]
    CITY_FIELD_NUMBER: _ClassVar[int]
    GROUP_FIELD_NUMBER: _ClassVar[int]
    source: ConnectionSource
    country: str
    city: str
    group: _group_pb2.ServerGroup
    def __init__(self, source: _Optional[_Union[ConnectionSource, str]] = ..., country: _Optional[str] = ..., city: _Optional[str] = ..., group: _Optional[_Union[_group_pb2.ServerGroup, str]] = ...) -> None: ...

class StatusResponse(_message.Message):
    __slots__ = ("state", "technology", "protocol", "ip", "hostname", "country", "city", "download", "upload", "uptime", "name", "virtualLocation", "parameters", "postQuantum", "is_mesh_peer", "by_user")
    STATE_FIELD_NUMBER: _ClassVar[int]
    TECHNOLOGY_FIELD_NUMBER: _ClassVar[int]
    PROTOCOL_FIELD_NUMBER: _ClassVar[int]
    IP_FIELD_NUMBER: _ClassVar[int]
    HOSTNAME_FIELD_NUMBER: _ClassVar[int]
    COUNTRY_FIELD_NUMBER: _ClassVar[int]
    CITY_FIELD_NUMBER: _ClassVar[int]
    DOWNLOAD_FIELD_NUMBER: _ClassVar[int]
    UPLOAD_FIELD_NUMBER: _ClassVar[int]
    UPTIME_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VIRTUALLOCATION_FIELD_NUMBER: _ClassVar[int]
    PARAMETERS_FIELD_NUMBER: _ClassVar[int]
    POSTQUANTUM_FIELD_NUMBER: _ClassVar[int]
    IS_MESH_PEER_FIELD_NUMBER: _ClassVar[int]
    BY_USER_FIELD_NUMBER: _ClassVar[int]
    state: ConnectionState
    technology: _technology_pb2.Technology
    protocol: _protocol_pb2.Protocol
    ip: str
    hostname: str
    country: str
    city: str
    download: int
    upload: int
    uptime: int
    name: str
    virtualLocation: bool
    parameters: ConnectionParameters
    postQuantum: bool
    is_mesh_peer: bool
    by_user: bool
    def __init__(self, state: _Optional[_Union[ConnectionState, str]] = ..., technology: _Optional[_Union[_technology_pb2.Technology, str]] = ..., protocol: _Optional[_Union[_protocol_pb2.Protocol, str]] = ..., ip: _Optional[str] = ..., hostname: _Optional[str] = ..., country: _Optional[str] = ..., city: _Optional[str] = ..., download: _Optional[int] = ..., upload: _Optional[int] = ..., uptime: _Optional[int] = ..., name: _Optional[str] = ..., virtualLocation: bool = ..., parameters: _Optional[_Union[ConnectionParameters, _Mapping]] = ..., postQuantum: bool = ..., is_mesh_peer: bool = ..., by_user: bool = ...) -> None: ...
