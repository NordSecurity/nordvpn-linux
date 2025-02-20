import empty_pb2 as _empty_pb2
import service_response_pb2 as _service_response_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PeerStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    DISCONNECTED: _ClassVar[PeerStatus]
    CONNECTED: _ClassVar[PeerStatus]

class UpdatePeerErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PEER_NOT_FOUND: _ClassVar[UpdatePeerErrorCode]

class ChangeNicknameErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SAME_NICKNAME: _ClassVar[ChangeNicknameErrorCode]
    NICKNAME_ALREADY_EMPTY: _ClassVar[ChangeNicknameErrorCode]
    DOMAIN_NAME_EXISTS: _ClassVar[ChangeNicknameErrorCode]
    RATE_LIMIT_REACH: _ClassVar[ChangeNicknameErrorCode]
    NICKNAME_TOO_LONG: _ClassVar[ChangeNicknameErrorCode]
    DUPLICATE_NICKNAME: _ClassVar[ChangeNicknameErrorCode]
    CONTAINS_FORBIDDEN_WORD: _ClassVar[ChangeNicknameErrorCode]
    SUFFIX_OR_PREFIX_ARE_INVALID: _ClassVar[ChangeNicknameErrorCode]
    NICKNAME_HAS_DOUBLE_HYPHENS: _ClassVar[ChangeNicknameErrorCode]
    INVALID_CHARS: _ClassVar[ChangeNicknameErrorCode]

class AllowRoutingErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ROUTING_ALREADY_ALLOWED: _ClassVar[AllowRoutingErrorCode]

class DenyRoutingErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ROUTING_ALREADY_DENIED: _ClassVar[DenyRoutingErrorCode]

class AllowIncomingErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    INCOMING_ALREADY_ALLOWED: _ClassVar[AllowIncomingErrorCode]

class DenyIncomingErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    INCOMING_ALREADY_DENIED: _ClassVar[DenyIncomingErrorCode]

class AllowLocalNetworkErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    LOCAL_NETWORK_ALREADY_ALLOWED: _ClassVar[AllowLocalNetworkErrorCode]

class DenyLocalNetworkErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    LOCAL_NETWORK_ALREADY_DENIED: _ClassVar[DenyLocalNetworkErrorCode]

class AllowFileshareErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SEND_ALREADY_ALLOWED: _ClassVar[AllowFileshareErrorCode]

class DenyFileshareErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SEND_ALREADY_DENIED: _ClassVar[DenyFileshareErrorCode]

class EnableAutomaticFileshareErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    AUTOMATIC_FILESHARE_ALREADY_ENABLED: _ClassVar[EnableAutomaticFileshareErrorCode]

class DisableAutomaticFileshareErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    AUTOMATIC_FILESHARE_ALREADY_DISABLED: _ClassVar[DisableAutomaticFileshareErrorCode]

class ConnectErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PEER_DOES_NOT_ALLOW_ROUTING: _ClassVar[ConnectErrorCode]
    ALREADY_CONNECTED: _ClassVar[ConnectErrorCode]
    CONNECT_FAILED: _ClassVar[ConnectErrorCode]
    PEER_NO_IP: _ClassVar[ConnectErrorCode]
    ALREADY_CONNECTING: _ClassVar[ConnectErrorCode]
    CANCELED: _ClassVar[ConnectErrorCode]
DISCONNECTED: PeerStatus
CONNECTED: PeerStatus
PEER_NOT_FOUND: UpdatePeerErrorCode
SAME_NICKNAME: ChangeNicknameErrorCode
NICKNAME_ALREADY_EMPTY: ChangeNicknameErrorCode
DOMAIN_NAME_EXISTS: ChangeNicknameErrorCode
RATE_LIMIT_REACH: ChangeNicknameErrorCode
NICKNAME_TOO_LONG: ChangeNicknameErrorCode
DUPLICATE_NICKNAME: ChangeNicknameErrorCode
CONTAINS_FORBIDDEN_WORD: ChangeNicknameErrorCode
SUFFIX_OR_PREFIX_ARE_INVALID: ChangeNicknameErrorCode
NICKNAME_HAS_DOUBLE_HYPHENS: ChangeNicknameErrorCode
INVALID_CHARS: ChangeNicknameErrorCode
ROUTING_ALREADY_ALLOWED: AllowRoutingErrorCode
ROUTING_ALREADY_DENIED: DenyRoutingErrorCode
INCOMING_ALREADY_ALLOWED: AllowIncomingErrorCode
INCOMING_ALREADY_DENIED: DenyIncomingErrorCode
LOCAL_NETWORK_ALREADY_ALLOWED: AllowLocalNetworkErrorCode
LOCAL_NETWORK_ALREADY_DENIED: DenyLocalNetworkErrorCode
SEND_ALREADY_ALLOWED: AllowFileshareErrorCode
SEND_ALREADY_DENIED: DenyFileshareErrorCode
AUTOMATIC_FILESHARE_ALREADY_ENABLED: EnableAutomaticFileshareErrorCode
AUTOMATIC_FILESHARE_ALREADY_DISABLED: DisableAutomaticFileshareErrorCode
PEER_DOES_NOT_ALLOW_ROUTING: ConnectErrorCode
ALREADY_CONNECTED: ConnectErrorCode
CONNECT_FAILED: ConnectErrorCode
PEER_NO_IP: ConnectErrorCode
ALREADY_CONNECTING: ConnectErrorCode
CANCELED: ConnectErrorCode

class GetPeersResponse(_message.Message):
    __slots__ = ("peers", "error")
    PEERS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    peers: PeerList
    error: Error
    def __init__(self, peers: _Optional[_Union[PeerList, _Mapping]] = ..., error: _Optional[_Union[Error, _Mapping]] = ...) -> None: ...

class PeerList(_message.Message):
    __slots__ = ("self", "local", "external")
    SELF_FIELD_NUMBER: _ClassVar[int]
    LOCAL_FIELD_NUMBER: _ClassVar[int]
    EXTERNAL_FIELD_NUMBER: _ClassVar[int]
    self: Peer
    local: _containers.RepeatedCompositeFieldContainer[Peer]
    external: _containers.RepeatedCompositeFieldContainer[Peer]
    def __init__(self, self_: _Optional[_Union[Peer, _Mapping]] = ..., local: _Optional[_Iterable[_Union[Peer, _Mapping]]] = ..., external: _Optional[_Iterable[_Union[Peer, _Mapping]]] = ...) -> None: ...

class Peer(_message.Message):
    __slots__ = ("identifier", "pubkey", "ip", "endpoints", "os", "os_version", "hostname", "distro", "email", "is_inbound_allowed", "is_routable", "is_local_network_allowed", "is_fileshare_allowed", "do_i_allow_inbound", "do_i_allow_routing", "do_i_allow_local_network", "do_i_allow_fileshare", "always_accept_files", "status", "nickname")
    IDENTIFIER_FIELD_NUMBER: _ClassVar[int]
    PUBKEY_FIELD_NUMBER: _ClassVar[int]
    IP_FIELD_NUMBER: _ClassVar[int]
    ENDPOINTS_FIELD_NUMBER: _ClassVar[int]
    OS_FIELD_NUMBER: _ClassVar[int]
    OS_VERSION_FIELD_NUMBER: _ClassVar[int]
    HOSTNAME_FIELD_NUMBER: _ClassVar[int]
    DISTRO_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    IS_INBOUND_ALLOWED_FIELD_NUMBER: _ClassVar[int]
    IS_ROUTABLE_FIELD_NUMBER: _ClassVar[int]
    IS_LOCAL_NETWORK_ALLOWED_FIELD_NUMBER: _ClassVar[int]
    IS_FILESHARE_ALLOWED_FIELD_NUMBER: _ClassVar[int]
    DO_I_ALLOW_INBOUND_FIELD_NUMBER: _ClassVar[int]
    DO_I_ALLOW_ROUTING_FIELD_NUMBER: _ClassVar[int]
    DO_I_ALLOW_LOCAL_NETWORK_FIELD_NUMBER: _ClassVar[int]
    DO_I_ALLOW_FILESHARE_FIELD_NUMBER: _ClassVar[int]
    ALWAYS_ACCEPT_FILES_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    NICKNAME_FIELD_NUMBER: _ClassVar[int]
    identifier: str
    pubkey: str
    ip: str
    endpoints: _containers.RepeatedScalarFieldContainer[str]
    os: str
    os_version: str
    hostname: str
    distro: str
    email: str
    is_inbound_allowed: bool
    is_routable: bool
    is_local_network_allowed: bool
    is_fileshare_allowed: bool
    do_i_allow_inbound: bool
    do_i_allow_routing: bool
    do_i_allow_local_network: bool
    do_i_allow_fileshare: bool
    always_accept_files: bool
    status: PeerStatus
    nickname: str
    def __init__(self, identifier: _Optional[str] = ..., pubkey: _Optional[str] = ..., ip: _Optional[str] = ..., endpoints: _Optional[_Iterable[str]] = ..., os: _Optional[str] = ..., os_version: _Optional[str] = ..., hostname: _Optional[str] = ..., distro: _Optional[str] = ..., email: _Optional[str] = ..., is_inbound_allowed: bool = ..., is_routable: bool = ..., is_local_network_allowed: bool = ..., is_fileshare_allowed: bool = ..., do_i_allow_inbound: bool = ..., do_i_allow_routing: bool = ..., do_i_allow_local_network: bool = ..., do_i_allow_fileshare: bool = ..., always_accept_files: bool = ..., status: _Optional[_Union[PeerStatus, str]] = ..., nickname: _Optional[str] = ...) -> None: ...

class UpdatePeerRequest(_message.Message):
    __slots__ = ("identifier",)
    IDENTIFIER_FIELD_NUMBER: _ClassVar[int]
    identifier: str
    def __init__(self, identifier: _Optional[str] = ...) -> None: ...

class Error(_message.Message):
    __slots__ = ("service_error_code", "meshnet_error_code")
    SERVICE_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    MESHNET_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    service_error_code: _service_response_pb2.ServiceErrorCode
    meshnet_error_code: _service_response_pb2.MeshnetErrorCode
    def __init__(self, service_error_code: _Optional[_Union[_service_response_pb2.ServiceErrorCode, str]] = ..., meshnet_error_code: _Optional[_Union[_service_response_pb2.MeshnetErrorCode, str]] = ...) -> None: ...

class UpdatePeerError(_message.Message):
    __slots__ = ("general_error", "update_peer_error_code")
    GENERAL_ERROR_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    general_error: Error
    update_peer_error_code: UpdatePeerErrorCode
    def __init__(self, general_error: _Optional[_Union[Error, _Mapping]] = ..., update_peer_error_code: _Optional[_Union[UpdatePeerErrorCode, str]] = ...) -> None: ...

class RemovePeerResponse(_message.Message):
    __slots__ = ("empty", "update_peer_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    update_peer_error: UpdatePeerError
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ...) -> None: ...

class ChangePeerNicknameRequest(_message.Message):
    __slots__ = ("identifier", "nickname")
    IDENTIFIER_FIELD_NUMBER: _ClassVar[int]
    NICKNAME_FIELD_NUMBER: _ClassVar[int]
    identifier: str
    nickname: str
    def __init__(self, identifier: _Optional[str] = ..., nickname: _Optional[str] = ...) -> None: ...

class ChangeMachineNicknameRequest(_message.Message):
    __slots__ = ("nickname",)
    NICKNAME_FIELD_NUMBER: _ClassVar[int]
    nickname: str
    def __init__(self, nickname: _Optional[str] = ...) -> None: ...

class ChangeNicknameResponse(_message.Message):
    __slots__ = ("empty", "change_nickname_error_code", "update_peer_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    CHANGE_NICKNAME_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    change_nickname_error_code: ChangeNicknameErrorCode
    update_peer_error: UpdatePeerError
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., change_nickname_error_code: _Optional[_Union[ChangeNicknameErrorCode, str]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ...) -> None: ...

class AllowRoutingResponse(_message.Message):
    __slots__ = ("empty", "allow_routing_error_code", "update_peer_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    ALLOW_ROUTING_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    allow_routing_error_code: AllowRoutingErrorCode
    update_peer_error: UpdatePeerError
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., allow_routing_error_code: _Optional[_Union[AllowRoutingErrorCode, str]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ...) -> None: ...

class DenyRoutingResponse(_message.Message):
    __slots__ = ("empty", "deny_routing_error_code", "update_peer_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    DENY_ROUTING_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    deny_routing_error_code: DenyRoutingErrorCode
    update_peer_error: UpdatePeerError
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., deny_routing_error_code: _Optional[_Union[DenyRoutingErrorCode, str]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ...) -> None: ...

class AllowIncomingResponse(_message.Message):
    __slots__ = ("empty", "allow_incoming_error_code", "update_peer_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    ALLOW_INCOMING_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    allow_incoming_error_code: AllowIncomingErrorCode
    update_peer_error: UpdatePeerError
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., allow_incoming_error_code: _Optional[_Union[AllowIncomingErrorCode, str]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ...) -> None: ...

class DenyIncomingResponse(_message.Message):
    __slots__ = ("empty", "deny_incoming_error_code", "update_peer_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    DENY_INCOMING_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    deny_incoming_error_code: DenyIncomingErrorCode
    update_peer_error: UpdatePeerError
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., deny_incoming_error_code: _Optional[_Union[DenyIncomingErrorCode, str]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ...) -> None: ...

class AllowLocalNetworkResponse(_message.Message):
    __slots__ = ("empty", "allow_local_network_error_code", "update_peer_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    ALLOW_LOCAL_NETWORK_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    allow_local_network_error_code: AllowLocalNetworkErrorCode
    update_peer_error: UpdatePeerError
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., allow_local_network_error_code: _Optional[_Union[AllowLocalNetworkErrorCode, str]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ...) -> None: ...

class DenyLocalNetworkResponse(_message.Message):
    __slots__ = ("empty", "deny_local_network_error_code", "update_peer_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    DENY_LOCAL_NETWORK_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    deny_local_network_error_code: DenyLocalNetworkErrorCode
    update_peer_error: UpdatePeerError
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., deny_local_network_error_code: _Optional[_Union[DenyLocalNetworkErrorCode, str]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ...) -> None: ...

class AllowFileshareResponse(_message.Message):
    __slots__ = ("empty", "allow_send_error_code", "update_peer_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    ALLOW_SEND_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    allow_send_error_code: AllowFileshareErrorCode
    update_peer_error: UpdatePeerError
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., allow_send_error_code: _Optional[_Union[AllowFileshareErrorCode, str]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ...) -> None: ...

class DenyFileshareResponse(_message.Message):
    __slots__ = ("empty", "deny_send_error_code", "update_peer_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    DENY_SEND_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    deny_send_error_code: DenyFileshareErrorCode
    update_peer_error: UpdatePeerError
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., deny_send_error_code: _Optional[_Union[DenyFileshareErrorCode, str]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ...) -> None: ...

class EnableAutomaticFileshareResponse(_message.Message):
    __slots__ = ("empty", "enable_automatic_fileshare_error_code", "update_peer_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    ENABLE_AUTOMATIC_FILESHARE_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    enable_automatic_fileshare_error_code: EnableAutomaticFileshareErrorCode
    update_peer_error: UpdatePeerError
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., enable_automatic_fileshare_error_code: _Optional[_Union[EnableAutomaticFileshareErrorCode, str]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ...) -> None: ...

class DisableAutomaticFileshareResponse(_message.Message):
    __slots__ = ("empty", "disable_automatic_fileshare_error_code", "update_peer_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    DISABLE_AUTOMATIC_FILESHARE_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    disable_automatic_fileshare_error_code: DisableAutomaticFileshareErrorCode
    update_peer_error: UpdatePeerError
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., disable_automatic_fileshare_error_code: _Optional[_Union[DisableAutomaticFileshareErrorCode, str]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ...) -> None: ...

class ConnectResponse(_message.Message):
    __slots__ = ("empty", "connect_error_code", "update_peer_error", "meshnet_error")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    CONNECT_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_FIELD_NUMBER: _ClassVar[int]
    MESHNET_ERROR_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    connect_error_code: ConnectErrorCode
    update_peer_error: UpdatePeerError
    meshnet_error: _service_response_pb2.MeshnetErrorCode
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., connect_error_code: _Optional[_Union[ConnectErrorCode, str]] = ..., update_peer_error: _Optional[_Union[UpdatePeerError, _Mapping]] = ..., meshnet_error: _Optional[_Union[_service_response_pb2.MeshnetErrorCode, str]] = ...) -> None: ...

class PrivateKeyResponse(_message.Message):
    __slots__ = ("private_key", "service_error_code")
    PRIVATE_KEY_FIELD_NUMBER: _ClassVar[int]
    SERVICE_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    private_key: str
    service_error_code: _service_response_pb2.ServiceErrorCode
    def __init__(self, private_key: _Optional[str] = ..., service_error_code: _Optional[_Union[_service_response_pb2.ServiceErrorCode, str]] = ...) -> None: ...
