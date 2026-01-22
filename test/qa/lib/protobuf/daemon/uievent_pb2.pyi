from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class UIEvent(_message.Message):
    __slots__ = ()
    class FormReference(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        FORM_REFERENCE_UNSPECIFIED: _ClassVar[UIEvent.FormReference]
        CLI: _ClassVar[UIEvent.FormReference]
        TRAY: _ClassVar[UIEvent.FormReference]
        HOME_SCREEN: _ClassVar[UIEvent.FormReference]
        GUI: _ClassVar[UIEvent.FormReference]
    FORM_REFERENCE_UNSPECIFIED: UIEvent.FormReference
    CLI: UIEvent.FormReference
    TRAY: UIEvent.FormReference
    HOME_SCREEN: UIEvent.FormReference
    GUI: UIEvent.FormReference
    class ItemName(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        ITEM_NAME_UNSPECIFIED: _ClassVar[UIEvent.ItemName]
        CONNECT: _ClassVar[UIEvent.ItemName]
        CONNECT_RECENTS: _ClassVar[UIEvent.ItemName]
        DISCONNECT: _ClassVar[UIEvent.ItemName]
        LOGIN: _ClassVar[UIEvent.ItemName]
        LOGOUT: _ClassVar[UIEvent.ItemName]
        RATE_CONNECTION: _ClassVar[UIEvent.ItemName]
        MESHNET_INVITE_SEND: _ClassVar[UIEvent.ItemName]
        LOGIN_TOKEN: _ClassVar[UIEvent.ItemName]
    ITEM_NAME_UNSPECIFIED: UIEvent.ItemName
    CONNECT: UIEvent.ItemName
    CONNECT_RECENTS: UIEvent.ItemName
    DISCONNECT: UIEvent.ItemName
    LOGIN: UIEvent.ItemName
    LOGOUT: UIEvent.ItemName
    RATE_CONNECTION: UIEvent.ItemName
    MESHNET_INVITE_SEND: UIEvent.ItemName
    LOGIN_TOKEN: UIEvent.ItemName
    class ItemType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        ITEM_TYPE_UNSPECIFIED: _ClassVar[UIEvent.ItemType]
        CLICK: _ClassVar[UIEvent.ItemType]
    ITEM_TYPE_UNSPECIFIED: UIEvent.ItemType
    CLICK: UIEvent.ItemType
    class ItemValue(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        ITEM_VALUE_UNSPECIFIED: _ClassVar[UIEvent.ItemValue]
        COUNTRY: _ClassVar[UIEvent.ItemValue]
        CITY: _ClassVar[UIEvent.ItemValue]
        DIP: _ClassVar[UIEvent.ItemValue]
        MESHNET: _ClassVar[UIEvent.ItemValue]
        OBFUSCATED: _ClassVar[UIEvent.ItemValue]
        ONION_OVER_VPN: _ClassVar[UIEvent.ItemValue]
        DOUBLE_VPN: _ClassVar[UIEvent.ItemValue]
        P2P: _ClassVar[UIEvent.ItemValue]
    ITEM_VALUE_UNSPECIFIED: UIEvent.ItemValue
    COUNTRY: UIEvent.ItemValue
    CITY: UIEvent.ItemValue
    DIP: UIEvent.ItemValue
    MESHNET: UIEvent.ItemValue
    OBFUSCATED: UIEvent.ItemValue
    ONION_OVER_VPN: UIEvent.ItemValue
    DOUBLE_VPN: UIEvent.ItemValue
    P2P: UIEvent.ItemValue
    def __init__(self) -> None: ...
