from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class LoginType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    LoginType_UNKNOWN: _ClassVar[LoginType]
    LoginType_LOGIN: _ClassVar[LoginType]
    LoginType_SIGNUP: _ClassVar[LoginType]

class LoginOAuth2Error(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UNKNOWN_OAUTH2_ERROR: _ClassVar[LoginOAuth2Error]
    NO_NET: _ClassVar[LoginOAuth2Error]
LoginType_UNKNOWN: LoginType
LoginType_LOGIN: LoginType
LoginType_SIGNUP: LoginType
UNKNOWN_OAUTH2_ERROR: LoginOAuth2Error
NO_NET: LoginOAuth2Error

class LoginOAuth2Request(_message.Message):
    __slots__ = ("type",)
    TYPE_FIELD_NUMBER: _ClassVar[int]
    type: LoginType
    def __init__(self, type: _Optional[_Union[LoginType, str]] = ...) -> None: ...

class LoginOAuth2CallbackRequest(_message.Message):
    __slots__ = ("token", "type")
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    token: str
    type: LoginType
    def __init__(self, token: _Optional[str] = ..., type: _Optional[_Union[LoginType, str]] = ...) -> None: ...

class LoginResponse(_message.Message):
    __slots__ = ("type", "url")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    type: int
    url: str
    def __init__(self, type: _Optional[int] = ..., url: _Optional[str] = ...) -> None: ...

class String(_message.Message):
    __slots__ = ("data",)
    DATA_FIELD_NUMBER: _ClassVar[int]
    data: str
    def __init__(self, data: _Optional[str] = ...) -> None: ...

class LoginOAuth2Response(_message.Message):
    __slots__ = ("url", "error")
    URL_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    url: String
    error: LoginOAuth2Error
    def __init__(self, url: _Optional[_Union[String, _Mapping]] = ..., error: _Optional[_Union[LoginOAuth2Error, str]] = ...) -> None: ...
