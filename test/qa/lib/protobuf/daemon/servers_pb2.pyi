from config import group_pb2 as _group_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ServersError(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    NO_ERROR: _ClassVar[ServersError]
    GET_CONFIG_ERROR: _ClassVar[ServersError]
    FILTER_SERVERS_ERROR: _ClassVar[ServersError]

class Technology(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UNKNOWN_TECHNLOGY: _ClassVar[Technology]
    NORDLYNX: _ClassVar[Technology]
    OPENVPN_TCP: _ClassVar[Technology]
    OPENVPN_UDP: _ClassVar[Technology]
    OBFUSCATED_OPENVPN_TCP: _ClassVar[Technology]
    OBFUSCATED_OPENVPN_UDP: _ClassVar[Technology]
NO_ERROR: ServersError
GET_CONFIG_ERROR: ServersError
FILTER_SERVERS_ERROR: ServersError
UNKNOWN_TECHNLOGY: Technology
NORDLYNX: Technology
OPENVPN_TCP: Technology
OPENVPN_UDP: Technology
OBFUSCATED_OPENVPN_TCP: Technology
OBFUSCATED_OPENVPN_UDP: Technology

class Server(_message.Message):
    __slots__ = ("id", "host_name", "virtual", "server_groups", "technologies")
    ID_FIELD_NUMBER: _ClassVar[int]
    HOST_NAME_FIELD_NUMBER: _ClassVar[int]
    VIRTUAL_FIELD_NUMBER: _ClassVar[int]
    SERVER_GROUPS_FIELD_NUMBER: _ClassVar[int]
    TECHNOLOGIES_FIELD_NUMBER: _ClassVar[int]
    id: int
    host_name: str
    virtual: bool
    server_groups: _containers.RepeatedScalarFieldContainer[_group_pb2.ServerGroup]
    technologies: _containers.RepeatedScalarFieldContainer[Technology]
    def __init__(self, id: _Optional[int] = ..., host_name: _Optional[str] = ..., virtual: bool = ..., server_groups: _Optional[_Iterable[_Union[_group_pb2.ServerGroup, str]]] = ..., technologies: _Optional[_Iterable[_Union[Technology, str]]] = ...) -> None: ...

class ServerCity(_message.Message):
    __slots__ = ("city_name", "servers")
    CITY_NAME_FIELD_NUMBER: _ClassVar[int]
    SERVERS_FIELD_NUMBER: _ClassVar[int]
    city_name: str
    servers: _containers.RepeatedCompositeFieldContainer[Server]
    def __init__(self, city_name: _Optional[str] = ..., servers: _Optional[_Iterable[_Union[Server, _Mapping]]] = ...) -> None: ...

class ServerCountry(_message.Message):
    __slots__ = ("country_code", "cities")
    COUNTRY_CODE_FIELD_NUMBER: _ClassVar[int]
    CITIES_FIELD_NUMBER: _ClassVar[int]
    country_code: str
    cities: _containers.RepeatedCompositeFieldContainer[ServerCity]
    def __init__(self, country_code: _Optional[str] = ..., cities: _Optional[_Iterable[_Union[ServerCity, _Mapping]]] = ...) -> None: ...

class ServersMap(_message.Message):
    __slots__ = ("servers_by_country",)
    SERVERS_BY_COUNTRY_FIELD_NUMBER: _ClassVar[int]
    servers_by_country: _containers.RepeatedCompositeFieldContainer[ServerCountry]
    def __init__(self, servers_by_country: _Optional[_Iterable[_Union[ServerCountry, _Mapping]]] = ...) -> None: ...

class ServersResponse(_message.Message):
    __slots__ = ("servers", "error")
    SERVERS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    servers: ServersMap
    error: ServersError
    def __init__(self, servers: _Optional[_Union[ServersMap, _Mapping]] = ..., error: _Optional[_Union[ServersError, str]] = ...) -> None: ...
