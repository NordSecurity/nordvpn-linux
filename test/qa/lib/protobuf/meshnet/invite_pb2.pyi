from protobuf.meshnet import empty_pb2 as _empty_pb2
from protobuf.meshnet import service_response_pb2 as _service_response_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RespondToInviteErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UNKNOWN: _ClassVar[RespondToInviteErrorCode]
    NO_SUCH_INVITATION: _ClassVar[RespondToInviteErrorCode]
    DEVICE_COUNT: _ClassVar[RespondToInviteErrorCode]

class InviteResponseErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ALREADY_EXISTS: _ClassVar[InviteResponseErrorCode]
    INVALID_EMAIL: _ClassVar[InviteResponseErrorCode]
    SAME_ACCOUNT_EMAIL: _ClassVar[InviteResponseErrorCode]
    LIMIT_REACHED: _ClassVar[InviteResponseErrorCode]
    PEER_COUNT: _ClassVar[InviteResponseErrorCode]
UNKNOWN: RespondToInviteErrorCode
NO_SUCH_INVITATION: RespondToInviteErrorCode
DEVICE_COUNT: RespondToInviteErrorCode
ALREADY_EXISTS: InviteResponseErrorCode
INVALID_EMAIL: InviteResponseErrorCode
SAME_ACCOUNT_EMAIL: InviteResponseErrorCode
LIMIT_REACHED: InviteResponseErrorCode
PEER_COUNT: InviteResponseErrorCode

class GetInvitesResponse(_message.Message):
    __slots__ = ("invites", "service_error_code", "meshnet_error_code")
    INVITES_FIELD_NUMBER: _ClassVar[int]
    SERVICE_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    MESHNET_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    invites: InvitesList
    service_error_code: _service_response_pb2.ServiceErrorCode
    meshnet_error_code: _service_response_pb2.MeshnetErrorCode
    def __init__(self, invites: _Optional[_Union[InvitesList, _Mapping]] = ..., service_error_code: _Optional[_Union[_service_response_pb2.ServiceErrorCode, str]] = ..., meshnet_error_code: _Optional[_Union[_service_response_pb2.MeshnetErrorCode, str]] = ...) -> None: ...

class InvitesList(_message.Message):
    __slots__ = ("sent", "received")
    SENT_FIELD_NUMBER: _ClassVar[int]
    RECEIVED_FIELD_NUMBER: _ClassVar[int]
    sent: _containers.RepeatedCompositeFieldContainer[Invite]
    received: _containers.RepeatedCompositeFieldContainer[Invite]
    def __init__(self, sent: _Optional[_Iterable[_Union[Invite, _Mapping]]] = ..., received: _Optional[_Iterable[_Union[Invite, _Mapping]]] = ...) -> None: ...

class Invite(_message.Message):
    __slots__ = ("email", "expires_at", "os")
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    OS_FIELD_NUMBER: _ClassVar[int]
    email: str
    expires_at: _timestamp_pb2.Timestamp
    os: str
    def __init__(self, email: _Optional[str] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., os: _Optional[str] = ...) -> None: ...

class InviteRequest(_message.Message):
    __slots__ = ("email", "allowIncomingTraffic", "allowTrafficRouting", "allowLocalNetwork", "allowFileshare")
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    ALLOWINCOMINGTRAFFIC_FIELD_NUMBER: _ClassVar[int]
    ALLOWTRAFFICROUTING_FIELD_NUMBER: _ClassVar[int]
    ALLOWLOCALNETWORK_FIELD_NUMBER: _ClassVar[int]
    ALLOWFILESHARE_FIELD_NUMBER: _ClassVar[int]
    email: str
    allowIncomingTraffic: bool
    allowTrafficRouting: bool
    allowLocalNetwork: bool
    allowFileshare: bool
    def __init__(self, email: _Optional[str] = ..., allowIncomingTraffic: bool = ..., allowTrafficRouting: bool = ..., allowLocalNetwork: bool = ..., allowFileshare: bool = ...) -> None: ...

class DenyInviteRequest(_message.Message):
    __slots__ = ("email",)
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    email: str
    def __init__(self, email: _Optional[str] = ...) -> None: ...

class RespondToInviteResponse(_message.Message):
    __slots__ = ("empty", "respond_to_invite_error_code", "service_error_code", "meshnet_error_code")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    RESPOND_TO_INVITE_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    SERVICE_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    MESHNET_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    respond_to_invite_error_code: RespondToInviteErrorCode
    service_error_code: _service_response_pb2.ServiceErrorCode
    meshnet_error_code: _service_response_pb2.MeshnetErrorCode
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., respond_to_invite_error_code: _Optional[_Union[RespondToInviteErrorCode, str]] = ..., service_error_code: _Optional[_Union[_service_response_pb2.ServiceErrorCode, str]] = ..., meshnet_error_code: _Optional[_Union[_service_response_pb2.MeshnetErrorCode, str]] = ...) -> None: ...

class InviteResponse(_message.Message):
    __slots__ = ("empty", "invite_response_error_code", "service_error_code", "meshnet_error_code")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    INVITE_RESPONSE_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    SERVICE_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    MESHNET_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    invite_response_error_code: InviteResponseErrorCode
    service_error_code: _service_response_pb2.ServiceErrorCode
    meshnet_error_code: _service_response_pb2.MeshnetErrorCode
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., invite_response_error_code: _Optional[_Union[InviteResponseErrorCode, str]] = ..., service_error_code: _Optional[_Union[_service_response_pb2.ServiceErrorCode, str]] = ..., meshnet_error_code: _Optional[_Union[_service_response_pb2.MeshnetErrorCode, str]] = ...) -> None: ...
