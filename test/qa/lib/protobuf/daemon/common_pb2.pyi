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

class ClientID(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UNKNOWN_CLIENT: _ClassVar[ClientID]
    CLI: _ClassVar[ClientID]
    GUI: _ClassVar[ClientID]
    TRAY: _ClassVar[ClientID]

class DiagnosticsErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    DIAGNOSTICS_ERROR_CODE_UNSPECIFIED: _ClassVar[DiagnosticsErrorCode]
    DIAGNOSTICS_ERROR_CODE_INTERNAL: _ClassVar[DiagnosticsErrorCode]
    DIAGNOSTICS_ERROR_CODE_FAILED_TO_CREATE_ZIP: _ClassVar[DiagnosticsErrorCode]
    DIAGNOSTICS_ERROR_CODE_CHOWN_FAILED: _ClassVar[DiagnosticsErrorCode]
    DIAGNOSTICS_ERROR_CODE_ZIP_TOO_LARGE: _ClassVar[DiagnosticsErrorCode]
    DIAGNOSTICS_ERROR_CODE_COLLECTION_FAILED: _ClassVar[DiagnosticsErrorCode]
    DIAGNOSTICS_ERROR_CODE_FAILED_TO_CLOSE_ZIP: _ClassVar[DiagnosticsErrorCode]
    DIAGNOSTICS_ERROR_CODE_NO_DAEMON_LOG_SOURCE: _ClassVar[DiagnosticsErrorCode]
UNKNOWN: TriState
DISABLED: TriState
ENABLED: TriState
UNKNOWN_CLIENT: ClientID
CLI: ClientID
GUI: ClientID
TRAY: ClientID
DIAGNOSTICS_ERROR_CODE_UNSPECIFIED: DiagnosticsErrorCode
DIAGNOSTICS_ERROR_CODE_INTERNAL: DiagnosticsErrorCode
DIAGNOSTICS_ERROR_CODE_FAILED_TO_CREATE_ZIP: DiagnosticsErrorCode
DIAGNOSTICS_ERROR_CODE_CHOWN_FAILED: DiagnosticsErrorCode
DIAGNOSTICS_ERROR_CODE_ZIP_TOO_LARGE: DiagnosticsErrorCode
DIAGNOSTICS_ERROR_CODE_COLLECTION_FAILED: DiagnosticsErrorCode
DIAGNOSTICS_ERROR_CODE_FAILED_TO_CLOSE_ZIP: DiagnosticsErrorCode
DIAGNOSTICS_ERROR_CODE_NO_DAEMON_LOG_SOURCE: DiagnosticsErrorCode

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

class InjectVpnConnectionErrorRequest(_message.Message):
    __slots__ = ("telio_code", "pubkey")
    TELIO_CODE_FIELD_NUMBER: _ClassVar[int]
    PUBKEY_FIELD_NUMBER: _ClassVar[int]
    telio_code: int
    pubkey: str
    def __init__(self, telio_code: _Optional[int] = ..., pubkey: _Optional[str] = ...) -> None: ...

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

class DiagnosticsProgress(_message.Message):
    __slots__ = ("step", "file_path", "error_code")
    STEP_FIELD_NUMBER: _ClassVar[int]
    FILE_PATH_FIELD_NUMBER: _ClassVar[int]
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    step: str
    file_path: str
    error_code: DiagnosticsErrorCode
    def __init__(self, step: _Optional[str] = ..., file_path: _Optional[str] = ..., error_code: _Optional[_Union[DiagnosticsErrorCode, str]] = ...) -> None: ...
