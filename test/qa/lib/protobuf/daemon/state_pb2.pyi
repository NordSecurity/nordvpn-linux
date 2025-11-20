import settings_pb2 as _settings_pb2
import status_pb2 as _status_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AppStateError(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    FAILED_TO_GET_UID: _ClassVar[AppStateError]

class LoginEventType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    LOGIN: _ClassVar[LoginEventType]
    LOGOUT: _ClassVar[LoginEventType]

class UpdateEvent(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SERVERS_LIST_UPDATE: _ClassVar[UpdateEvent]
    RECENTS_LIST_UPDATE: _ClassVar[UpdateEvent]
FAILED_TO_GET_UID: AppStateError
LOGIN: LoginEventType
LOGOUT: LoginEventType
SERVERS_LIST_UPDATE: UpdateEvent
RECENTS_LIST_UPDATE: UpdateEvent

class LoginEvent(_message.Message):
    __slots__ = ("type",)
    TYPE_FIELD_NUMBER: _ClassVar[int]
    type: LoginEventType
    def __init__(self, type: _Optional[_Union[LoginEventType, str]] = ...) -> None: ...

class AccountModification(_message.Message):
    __slots__ = ("expires_at",)
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    expires_at: str
    def __init__(self, expires_at: _Optional[str] = ...) -> None: ...

class VersionHealthStatus(_message.Message):
    __slots__ = ("status_code",)
    STATUS_CODE_FIELD_NUMBER: _ClassVar[int]
    status_code: int
    def __init__(self, status_code: _Optional[int] = ...) -> None: ...

class AppState(_message.Message):
    __slots__ = ("error", "connection_status", "login_event", "settings_change", "update_event", "account_modification", "version_health")
    ERROR_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_STATUS_FIELD_NUMBER: _ClassVar[int]
    LOGIN_EVENT_FIELD_NUMBER: _ClassVar[int]
    SETTINGS_CHANGE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_EVENT_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_MODIFICATION_FIELD_NUMBER: _ClassVar[int]
    VERSION_HEALTH_FIELD_NUMBER: _ClassVar[int]
    error: AppStateError
    connection_status: _status_pb2.StatusResponse
    login_event: LoginEvent
    settings_change: _settings_pb2.Settings
    update_event: UpdateEvent
    account_modification: AccountModification
    version_health: VersionHealthStatus
    def __init__(self, error: _Optional[_Union[AppStateError, str]] = ..., connection_status: _Optional[_Union[_status_pb2.StatusResponse, _Mapping]] = ..., login_event: _Optional[_Union[LoginEvent, _Mapping]] = ..., settings_change: _Optional[_Union[_settings_pb2.Settings, _Mapping]] = ..., update_event: _Optional[_Union[UpdateEvent, str]] = ..., account_modification: _Optional[_Union[AccountModification, _Mapping]] = ..., version_health: _Optional[_Union[VersionHealthStatus, _Mapping]] = ...) -> None: ...
