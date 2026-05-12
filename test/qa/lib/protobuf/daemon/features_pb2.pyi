from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class FeatureToggles(_message.Message):
    __slots__ = ("meshnet_enabled", "dedicatedservers_enabled")
    MESHNET_ENABLED_FIELD_NUMBER: _ClassVar[int]
    DEDICATEDSERVERS_ENABLED_FIELD_NUMBER: _ClassVar[int]
    meshnet_enabled: bool
    dedicatedservers_enabled: bool
    def __init__(self, meshnet_enabled: bool = ..., dedicatedservers_enabled: bool = ...) -> None: ...
