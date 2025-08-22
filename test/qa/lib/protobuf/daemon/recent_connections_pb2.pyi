from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ServerSelectionRule(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SERVER_SELECTION_RULE_NONE: _ClassVar[ServerSelectionRule]
    SERVER_SELECTION_RULE_RECOMMENDED: _ClassVar[ServerSelectionRule]
    SERVER_SELECTION_RULE_CITY: _ClassVar[ServerSelectionRule]
    SERVER_SELECTION_RULE_COUNTRY: _ClassVar[ServerSelectionRule]
    SERVER_SELECTION_RULE_SPECIFIC_SERVER: _ClassVar[ServerSelectionRule]
    SERVER_SELECTION_RULE_GROUP: _ClassVar[ServerSelectionRule]
    SERVER_SELECTION_RULE_COUNTRY_WITH_GROUP: _ClassVar[ServerSelectionRule]
    SERVER_SELECTION_RULE_SPECIFIC_SERVER_WITH_GROUP: _ClassVar[ServerSelectionRule]
SERVER_SELECTION_RULE_NONE: ServerSelectionRule
SERVER_SELECTION_RULE_RECOMMENDED: ServerSelectionRule
SERVER_SELECTION_RULE_CITY: ServerSelectionRule
SERVER_SELECTION_RULE_COUNTRY: ServerSelectionRule
SERVER_SELECTION_RULE_SPECIFIC_SERVER: ServerSelectionRule
SERVER_SELECTION_RULE_GROUP: ServerSelectionRule
SERVER_SELECTION_RULE_COUNTRY_WITH_GROUP: ServerSelectionRule
SERVER_SELECTION_RULE_SPECIFIC_SERVER_WITH_GROUP: ServerSelectionRule

class RecentConnectionsResponse(_message.Message):
    __slots__ = ("connections",)
    CONNECTIONS_FIELD_NUMBER: _ClassVar[int]
    connections: _containers.RepeatedCompositeFieldContainer[RecentConnection]
    def __init__(self, connections: _Optional[_Iterable[_Union[RecentConnection, _Mapping]]] = ...) -> None: ...

class RecentConnection(_message.Message):
    __slots__ = ("connection_model", "display_label")
    CONNECTION_MODEL_FIELD_NUMBER: _ClassVar[int]
    DISPLAY_LABEL_FIELD_NUMBER: _ClassVar[int]
    connection_model: RecentConnectionModel
    display_label: str
    def __init__(self, connection_model: _Optional[_Union[RecentConnectionModel, _Mapping]] = ..., display_label: _Optional[str] = ...) -> None: ...

class RecentConnectionModel(_message.Message):
    __slots__ = ("country", "city", "specific_server", "specific_server_name", "group", "connection_type")
    COUNTRY_FIELD_NUMBER: _ClassVar[int]
    CITY_FIELD_NUMBER: _ClassVar[int]
    SPECIFIC_SERVER_FIELD_NUMBER: _ClassVar[int]
    SPECIFIC_SERVER_NAME_FIELD_NUMBER: _ClassVar[int]
    GROUP_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_TYPE_FIELD_NUMBER: _ClassVar[int]
    country: str
    city: str
    specific_server: str
    specific_server_name: str
    group: str
    connection_type: ServerSelectionRule
    def __init__(self, country: _Optional[str] = ..., city: _Optional[str] = ..., specific_server: _Optional[str] = ..., specific_server_name: _Optional[str] = ..., group: _Optional[str] = ..., connection_type: _Optional[_Union[ServerSelectionRule, str]] = ...) -> None: ...

class RecentConnectionsRequest(_message.Message):
    __slots__ = ("limit",)
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    limit: int
    def __init__(self, limit: _Optional[int] = ...) -> None: ...
