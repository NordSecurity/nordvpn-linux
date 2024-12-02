from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class TokenInfoResponse(_message.Message):
    __slots__ = ("type", "token", "expires_at", "trusted_pass_token", "trusted_pass_owner_id")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    TRUSTED_PASS_TOKEN_FIELD_NUMBER: _ClassVar[int]
    TRUSTED_PASS_OWNER_ID_FIELD_NUMBER: _ClassVar[int]
    type: int
    token: str
    expires_at: str
    trusted_pass_token: str
    trusted_pass_owner_id: str
    def __init__(self, type: _Optional[int] = ..., token: _Optional[str] = ..., expires_at: _Optional[str] = ..., trusted_pass_token: _Optional[str] = ..., trusted_pass_owner_id: _Optional[str] = ...) -> None: ...
