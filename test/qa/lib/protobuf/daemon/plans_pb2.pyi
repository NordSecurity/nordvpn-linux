from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PlansResponse(_message.Message):
    __slots__ = ("type", "plans")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    PLANS_FIELD_NUMBER: _ClassVar[int]
    type: int
    plans: _containers.RepeatedCompositeFieldContainer[Plan]
    def __init__(self, type: _Optional[int] = ..., plans: _Optional[_Iterable[_Union[Plan, _Mapping]]] = ...) -> None: ...

class Plan(_message.Message):
    __slots__ = ("id", "title", "cost", "currency")
    ID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    COST_FIELD_NUMBER: _ClassVar[int]
    CURRENCY_FIELD_NUMBER: _ClassVar[int]
    id: str
    title: str
    cost: str
    currency: str
    def __init__(self, id: _Optional[str] = ..., title: _Optional[str] = ..., cost: _Optional[str] = ..., currency: _Optional[str] = ...) -> None: ...
