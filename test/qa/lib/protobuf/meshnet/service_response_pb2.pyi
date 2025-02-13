import empty_pb2 as _empty_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ServiceErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    NOT_LOGGED_IN: _ClassVar[ServiceErrorCode]
    API_FAILURE: _ClassVar[ServiceErrorCode]
    CONFIG_FAILURE: _ClassVar[ServiceErrorCode]

class MeshnetErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    NOT_REGISTERED: _ClassVar[MeshnetErrorCode]
    LIB_FAILURE: _ClassVar[MeshnetErrorCode]
    ALREADY_ENABLED: _ClassVar[MeshnetErrorCode]
    ALREADY_DISABLED: _ClassVar[MeshnetErrorCode]
    NOT_ENABLED: _ClassVar[MeshnetErrorCode]
    TECH_FAILURE: _ClassVar[MeshnetErrorCode]
    TUNNEL_CLOSED: _ClassVar[MeshnetErrorCode]
    CONFLICT_WITH_PQ: _ClassVar[MeshnetErrorCode]
    CONFLICT_WITH_PQ_SERVER: _ClassVar[MeshnetErrorCode]
NOT_LOGGED_IN: ServiceErrorCode
API_FAILURE: ServiceErrorCode
CONFIG_FAILURE: ServiceErrorCode
NOT_REGISTERED: MeshnetErrorCode
LIB_FAILURE: MeshnetErrorCode
ALREADY_ENABLED: MeshnetErrorCode
ALREADY_DISABLED: MeshnetErrorCode
NOT_ENABLED: MeshnetErrorCode
TECH_FAILURE: MeshnetErrorCode
TUNNEL_CLOSED: MeshnetErrorCode
CONFLICT_WITH_PQ: MeshnetErrorCode
CONFLICT_WITH_PQ_SERVER: MeshnetErrorCode

class MeshnetResponse(_message.Message):
    __slots__ = ("empty", "service_error", "meshnet_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    SERVICE_ERROR_FIELD_NUMBER: _ClassVar[int]
    MESHNET_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    service_error: ServiceErrorCode
    meshnet_error: MeshnetErrorCode
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., service_error: _Optional[_Union[ServiceErrorCode, str]] = ..., meshnet_error: _Optional[_Union[MeshnetErrorCode, str]] = ...) -> None: ...

class ServiceResponse(_message.Message):
    __slots__ = ("empty", "error_code")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    error_code: ServiceErrorCode
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., error_code: _Optional[_Union[ServiceErrorCode, str]] = ...) -> None: ...

class ServiceBoolResponse(_message.Message):
    __slots__ = ("value", "error_code")
    VALUE_FIELD_NUMBER: _ClassVar[int]
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    value: bool
    error_code: ServiceErrorCode
    def __init__(self, value: bool = ..., error_code: _Optional[_Union[ServiceErrorCode, str]] = ...) -> None: ...

class EnabledStatus(_message.Message):
    __slots__ = ("value", "uid")
    VALUE_FIELD_NUMBER: _ClassVar[int]
    UID_FIELD_NUMBER: _ClassVar[int]
    value: bool
    uid: int
    def __init__(self, value: bool = ..., uid: _Optional[int] = ...) -> None: ...

class IsEnabledResponse(_message.Message):
    __slots__ = ("status", "error_code")
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    status: EnabledStatus
    error_code: ServiceErrorCode
    def __init__(self, status: _Optional[_Union[EnabledStatus, _Mapping]] = ..., error_code: _Optional[_Union[ServiceErrorCode, str]] = ...) -> None: ...
