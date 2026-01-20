from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class UIFormReference(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UI_FORM_REFERENCE_UNSPECIFIED: _ClassVar[UIFormReference]
    UI_FORM_REFERENCE_CLI: _ClassVar[UIFormReference]
    UI_FORM_REFERENCE_TRAY: _ClassVar[UIFormReference]
    UI_FORM_REFERENCE_HOME_SCREEN: _ClassVar[UIFormReference]

class UIItemName(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UI_ITEM_NAME_UNSPECIFIED: _ClassVar[UIItemName]
    UI_ITEM_NAME_CONNECT: _ClassVar[UIItemName]
    UI_ITEM_NAME_CONNECT_RECENTS: _ClassVar[UIItemName]
    UI_ITEM_NAME_DISCONNECT: _ClassVar[UIItemName]
    UI_ITEM_NAME_LOGIN: _ClassVar[UIItemName]
    UI_ITEM_NAME_LOGOUT: _ClassVar[UIItemName]
    UI_ITEM_NAME_RATE_CONNECTION: _ClassVar[UIItemName]
    UI_ITEM_NAME_MESHNET_INVITE_SEND: _ClassVar[UIItemName]

class UIItemType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UI_ITEM_TYPE_UNSPECIFIED: _ClassVar[UIItemType]
    UI_ITEM_TYPE_CLICK: _ClassVar[UIItemType]
    UI_ITEM_TYPE_SHOW: _ClassVar[UIItemType]

class UIItemValue(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UI_ITEM_VALUE_CONNECTION_UNSPECIFIED: _ClassVar[UIItemValue]
    UI_ITEM_VALUE_CONNECTION_COUNTRY: _ClassVar[UIItemValue]
    UI_ITEM_VALUE_CONNECTION_CITY: _ClassVar[UIItemValue]
    UI_ITEM_VALUE_CONNECTION_DIP: _ClassVar[UIItemValue]
    UI_ITEM_VALUE_CONNECTION_MESHNET: _ClassVar[UIItemValue]
    UI_ITEM_VALUE_CONNECTION_OBFUSCATED: _ClassVar[UIItemValue]
    UI_ITEM_VALUE_CONNECTION_ONION_OVER_VPN: _ClassVar[UIItemValue]
    UI_ITEM_VALUE_CONNECTION_DOUBLE_VPN: _ClassVar[UIItemValue]
    UI_ITEM_VALUE_CONNECTION_P2P: _ClassVar[UIItemValue]
UI_FORM_REFERENCE_UNSPECIFIED: UIFormReference
UI_FORM_REFERENCE_CLI: UIFormReference
UI_FORM_REFERENCE_TRAY: UIFormReference
UI_FORM_REFERENCE_HOME_SCREEN: UIFormReference
UI_ITEM_NAME_UNSPECIFIED: UIItemName
UI_ITEM_NAME_CONNECT: UIItemName
UI_ITEM_NAME_CONNECT_RECENTS: UIItemName
UI_ITEM_NAME_DISCONNECT: UIItemName
UI_ITEM_NAME_LOGIN: UIItemName
UI_ITEM_NAME_LOGOUT: UIItemName
UI_ITEM_NAME_RATE_CONNECTION: UIItemName
UI_ITEM_NAME_MESHNET_INVITE_SEND: UIItemName
UI_ITEM_TYPE_UNSPECIFIED: UIItemType
UI_ITEM_TYPE_CLICK: UIItemType
UI_ITEM_TYPE_SHOW: UIItemType
UI_ITEM_VALUE_CONNECTION_UNSPECIFIED: UIItemValue
UI_ITEM_VALUE_CONNECTION_COUNTRY: UIItemValue
UI_ITEM_VALUE_CONNECTION_CITY: UIItemValue
UI_ITEM_VALUE_CONNECTION_DIP: UIItemValue
UI_ITEM_VALUE_CONNECTION_MESHNET: UIItemValue
UI_ITEM_VALUE_CONNECTION_OBFUSCATED: UIItemValue
UI_ITEM_VALUE_CONNECTION_ONION_OVER_VPN: UIItemValue
UI_ITEM_VALUE_CONNECTION_DOUBLE_VPN: UIItemValue
UI_ITEM_VALUE_CONNECTION_P2P: UIItemValue
