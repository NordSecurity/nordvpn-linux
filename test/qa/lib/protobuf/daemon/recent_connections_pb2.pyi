import server_selection_rule_pb2 as _server_selection_rule_pb2
from config import group_pb2 as _group_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RecentConnectionsResponse(_message.Message):
    __slots__ = ("connections",)
    CONNECTIONS_FIELD_NUMBER: _ClassVar[int]
    connections: _containers.RepeatedCompositeFieldContainer[RecentConnectionModel]
    def __init__(self, connections: _Optional[_Iterable[_Union[RecentConnectionModel, _Mapping]]] = ...) -> None: ...

class RecentConnectionModel(_message.Message):
    __slots__ = ("country", "city", "group", "country_code", "specific_server_name", "specific_server", "connection_type")
    COUNTRY_FIELD_NUMBER: _ClassVar[int]
    CITY_FIELD_NUMBER: _ClassVar[int]
    GROUP_FIELD_NUMBER: _ClassVar[int]
    COUNTRY_CODE_FIELD_NUMBER: _ClassVar[int]
    SPECIFIC_SERVER_NAME_FIELD_NUMBER: _ClassVar[int]
    SPECIFIC_SERVER_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_TYPE_FIELD_NUMBER: _ClassVar[int]
    country: str
    city: str
    group: _group_pb2.ServerGroup
    country_code: str
    specific_server_name: str
    specific_server: str
    connection_type: _server_selection_rule_pb2.ServerSelectionRule
    def __init__(self, country: _Optional[str] = ..., city: _Optional[str] = ..., group: _Optional[_Union[_group_pb2.ServerGroup, str]] = ..., country_code: _Optional[str] = ..., specific_server_name: _Optional[str] = ..., specific_server: _Optional[str] = ..., connection_type: _Optional[_Union[_server_selection_rule_pb2.ServerSelectionRule, str]] = ...) -> None: ...

class RecentConnectionsRequest(_message.Message):
    __slots__ = ("limit",)
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    limit: int
    def __init__(self, limit: _Optional[int] = ...) -> None: ...
