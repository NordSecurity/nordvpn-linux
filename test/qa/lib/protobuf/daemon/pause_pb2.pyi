from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PauseInverval(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PAUSE_5_MIN: _ClassVar[PauseInverval]
    PAUSE_15_MIN: _ClassVar[PauseInverval]
    PAUSE_30_MIN: _ClassVar[PauseInverval]
    PAUSE_1_HOUR: _ClassVar[PauseInverval]
    PAUSE_24_HOURS: _ClassVar[PauseInverval]
PAUSE_5_MIN: PauseInverval
PAUSE_15_MIN: PauseInverval
PAUSE_30_MIN: PauseInverval
PAUSE_1_HOUR: PauseInverval
PAUSE_24_HOURS: PauseInverval

class PauseRequest(_message.Message):
    __slots__ = ("interval",)
    INTERVAL_FIELD_NUMBER: _ClassVar[int]
    interval: PauseInverval
    def __init__(self, interval: _Optional[_Union[PauseInverval, str]] = ...) -> None: ...
