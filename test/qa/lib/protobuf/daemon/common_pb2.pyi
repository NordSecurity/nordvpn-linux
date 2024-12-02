from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class TriState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UNKNOWN: _ClassVar[TriState]
    DISABLED: _ClassVar[TriState]
    ENABLED: _ClassVar[TriState]
UNKNOWN: TriState
DISABLED: TriState
ENABLED: TriState

class Empty(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class Bool(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bool
    def __init__(self, value: bool = ...) -> None: ...

class Payload(_message.Message):
    __slots__ = ("type", "data")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    type: int
    data: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, type: _Optional[int] = ..., data: _Optional[_Iterable[str]] = ...) -> None: ...

class Allowlist(_message.Message):
    __slots__ = ("ports", "subnets")
    PORTS_FIELD_NUMBER: _ClassVar[int]
    SUBNETS_FIELD_NUMBER: _ClassVar[int]
    ports: Ports
    subnets: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, ports: _Optional[_Union[Ports, _Mapping]] = ..., subnets: _Optional[_Iterable[str]] = ...) -> None: ...

class Ports(_message.Message):
    __slots__ = ("udp", "tcp")
    UDP_FIELD_NUMBER: _ClassVar[int]
    TCP_FIELD_NUMBER: _ClassVar[int]
    udp: _containers.RepeatedScalarFieldContainer[int]
    tcp: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, udp: _Optional[_Iterable[int]] = ..., tcp: _Optional[_Iterable[int]] = ...) -> None: ...

class ServerGroup(_message.Message):
    __slots__ = ("name", "virtualLocation")
    NAME_FIELD_NUMBER: _ClassVar[int]
    VIRTUALLOCATION_FIELD_NUMBER: _ClassVar[int]
    name: str
    virtualLocation: bool
    def __init__(self, name: _Optional[str] = ..., virtualLocation: bool = ...) -> None: ...

class ServerGroupsList(_message.Message):
    __slots__ = ("type", "servers")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    SERVERS_FIELD_NUMBER: _ClassVar[int]
    type: int
    servers: _containers.RepeatedCompositeFieldContainer[ServerGroup]
    def __init__(self, type: _Optional[int] = ..., servers: _Optional[_Iterable[_Union[ServerGroup, _Mapping]]] = ...) -> None: ...
