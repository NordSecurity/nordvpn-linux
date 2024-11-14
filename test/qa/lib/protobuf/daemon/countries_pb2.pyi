from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CountriesResponse(_message.Message):
    __slots__ = ("countries",)
    COUNTRIES_FIELD_NUMBER: _ClassVar[int]
    countries: _containers.RepeatedCompositeFieldContainer[Country]
    def __init__(self, countries: _Optional[_Iterable[_Union[Country, _Mapping]]] = ...) -> None: ...

class Country(_message.Message):
    __slots__ = ("name", "code")
    NAME_FIELD_NUMBER: _ClassVar[int]
    CODE_FIELD_NUMBER: _ClassVar[int]
    name: str
    code: str
    def __init__(self, name: _Optional[str] = ..., code: _Optional[str] = ...) -> None: ...
