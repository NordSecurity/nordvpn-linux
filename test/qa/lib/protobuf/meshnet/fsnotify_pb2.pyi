import empty_pb2 as _empty_pb2
import peer_pb2 as _peer_pb2
import service_response_pb2 as _service_response_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class NewTransferNotification(_message.Message):
    __slots__ = ("identifier", "os", "file_name", "file_count", "transfer_id")
    IDENTIFIER_FIELD_NUMBER: _ClassVar[int]
    OS_FIELD_NUMBER: _ClassVar[int]
    FILE_NAME_FIELD_NUMBER: _ClassVar[int]
    FILE_COUNT_FIELD_NUMBER: _ClassVar[int]
    TRANSFER_ID_FIELD_NUMBER: _ClassVar[int]
    identifier: str
    os: str
    file_name: str
    file_count: int
    transfer_id: str
    def __init__(self, identifier: _Optional[str] = ..., os: _Optional[str] = ..., file_name: _Optional[str] = ..., file_count: _Optional[int] = ..., transfer_id: _Optional[str] = ...) -> None: ...

class NotifyNewTransferResponse(_message.Message):
    __slots__ = ("empty", "update_peer_error_code", "service_error_code", "meshnet_error_code")
    EMPTY_FIELD_NUMBER: _ClassVar[int]
    UPDATE_PEER_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    SERVICE_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    MESHNET_ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    empty: _empty_pb2.Empty
    update_peer_error_code: _peer_pb2.UpdatePeerErrorCode
    service_error_code: _service_response_pb2.ServiceErrorCode
    meshnet_error_code: _service_response_pb2.MeshnetErrorCode
    def __init__(self, empty: _Optional[_Union[_empty_pb2.Empty, _Mapping]] = ..., update_peer_error_code: _Optional[_Union[_peer_pb2.UpdatePeerErrorCode, str]] = ..., service_error_code: _Optional[_Union[_service_response_pb2.ServiceErrorCode, str]] = ..., meshnet_error_code: _Optional[_Union[_service_response_pb2.MeshnetErrorCode, str]] = ...) -> None: ...
