import common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DedidcatedIPService(_message.Message):
    __slots__ = ("server_ids", "dedicated_ip_expires_at")
    SERVER_IDS_FIELD_NUMBER: _ClassVar[int]
    DEDICATED_IP_EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    server_ids: _containers.RepeatedScalarFieldContainer[int]
    dedicated_ip_expires_at: str
    def __init__(self, server_ids: _Optional[_Iterable[int]] = ..., dedicated_ip_expires_at: _Optional[str] = ...) -> None: ...

class AccountResponse(_message.Message):
    __slots__ = ("type", "username", "email", "expires_at", "dedicated_ip_status", "last_dedicated_ip_expires_at", "dedicated_ip_services", "mfa_status")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    USERNAME_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    DEDICATED_IP_STATUS_FIELD_NUMBER: _ClassVar[int]
    LAST_DEDICATED_IP_EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    DEDICATED_IP_SERVICES_FIELD_NUMBER: _ClassVar[int]
    MFA_STATUS_FIELD_NUMBER: _ClassVar[int]
    type: int
    username: str
    email: str
    expires_at: str
    dedicated_ip_status: int
    last_dedicated_ip_expires_at: str
    dedicated_ip_services: _containers.RepeatedCompositeFieldContainer[DedidcatedIPService]
    mfa_status: _common_pb2.TriState
    def __init__(self, type: _Optional[int] = ..., username: _Optional[str] = ..., email: _Optional[str] = ..., expires_at: _Optional[str] = ..., dedicated_ip_status: _Optional[int] = ..., last_dedicated_ip_expires_at: _Optional[str] = ..., dedicated_ip_services: _Optional[_Iterable[_Union[DedidcatedIPService, _Mapping]]] = ..., mfa_status: _Optional[_Union[_common_pb2.TriState, str]] = ...) -> None: ...

class AccountRequest(_message.Message):
    __slots__ = ("full",)
    FULL_FIELD_NUMBER: _ClassVar[int]
    full: bool
    def __init__(self, full: bool = ...) -> None: ...
