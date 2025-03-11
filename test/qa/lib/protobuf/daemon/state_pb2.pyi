from protobuf.daemon import settings_pb2 as _settings_pb2
from protobuf.daemon import status_pb2 as _status_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AppStateError(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    FAILED_TO_GET_UID: _ClassVar[AppStateError]

class ConnectionState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    DISCONNECTED: _ClassVar[ConnectionState]
    CONNECTING: _ClassVar[ConnectionState]
    CONNECTED: _ClassVar[ConnectionState]

class LoginEventType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    LOGIN: _ClassVar[LoginEventType]
    LOGOUT: _ClassVar[LoginEventType]

class UpdateEvent(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SERVERS_LIST_UPDATE: _ClassVar[UpdateEvent]
FAILED_TO_GET_UID: AppStateError
DISCONNECTED: ConnectionState
CONNECTING: ConnectionState
CONNECTED: ConnectionState
LOGIN: LoginEventType
LOGOUT: LoginEventType
SERVERS_LIST_UPDATE: UpdateEvent

class ConnectionStatus(_message.Message):
    __slots__ = ("state", "server_ip", "server_country", "server_city", "server_hostname", "server_name", "is_mesh_peer", "by_user", "is_virtual_location", "parameters")
    STATE_FIELD_NUMBER: _ClassVar[int]
    SERVER_IP_FIELD_NUMBER: _ClassVar[int]
    SERVER_COUNTRY_FIELD_NUMBER: _ClassVar[int]
    SERVER_CITY_FIELD_NUMBER: _ClassVar[int]
    SERVER_HOSTNAME_FIELD_NUMBER: _ClassVar[int]
    SERVER_NAME_FIELD_NUMBER: _ClassVar[int]
    IS_MESH_PEER_FIELD_NUMBER: _ClassVar[int]
    BY_USER_FIELD_NUMBER: _ClassVar[int]
    IS_VIRTUAL_LOCATION_FIELD_NUMBER: _ClassVar[int]
    PARAMETERS_FIELD_NUMBER: _ClassVar[int]
    state: ConnectionState
    server_ip: str
    server_country: str
    server_city: str
    server_hostname: str
    server_name: str
    is_mesh_peer: bool
    by_user: bool
    is_virtual_location: bool
    parameters: _status_pb2.ConnectionParameters
    def __init__(self, state: _Optional[_Union[ConnectionState, str]] = ..., server_ip: _Optional[str] = ..., server_country: _Optional[str] = ..., server_city: _Optional[str] = ..., server_hostname: _Optional[str] = ..., server_name: _Optional[str] = ..., is_mesh_peer: bool = ..., by_user: bool = ..., is_virtual_location: bool = ..., parameters: _Optional[_Union[_status_pb2.ConnectionParameters, _Mapping]] = ...) -> None: ...

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

class AppState(_message.Message):
    __slots__ = ("error", "connection_status", "login_event", "settings_change", "update_event", "account_modification")
    ERROR_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_STATUS_FIELD_NUMBER: _ClassVar[int]
    LOGIN_EVENT_FIELD_NUMBER: _ClassVar[int]
    SETTINGS_CHANGE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_EVENT_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_MODIFICATION_FIELD_NUMBER: _ClassVar[int]
    error: AppStateError
    connection_status: ConnectionStatus
    login_event: LoginEvent
    settings_change: _settings_pb2.Settings
    update_event: UpdateEvent
    account_modification: AccountModification
    def __init__(self, error: _Optional[_Union[AppStateError, str]] = ..., connection_status: _Optional[_Union[ConnectionStatus, _Mapping]] = ..., login_event: _Optional[_Union[LoginEvent, _Mapping]] = ..., settings_change: _Optional[_Union[_settings_pb2.Settings, _Mapping]] = ..., update_event: _Optional[_Union[UpdateEvent, str]] = ..., account_modification: _Optional[_Union[AccountModification, _Mapping]] = ...) -> None: ...
