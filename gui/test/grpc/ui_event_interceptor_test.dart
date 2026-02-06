import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/grpc/ui_event_interceptor.dart';
import 'package:nordvpn/pb/daemon/uievent.pbenum.dart';

void main() {
  group('metadata keys', () {
    test('form reference key matches Go constant', () {
      expect(metadataKeyFormReference, 'ui-form-reference');
    });

    test('item name key matches Go constant', () {
      expect(metadataKeyItemName, 'ui-item-name');
    });

    test('item type key matches Go constant', () {
      expect(metadataKeyItemType, 'ui-item-type');
    });

    test('item value key matches Go constant', () {
      expect(metadataKeyItemValue, 'ui-item-value');
    });
  });

  group('createUiEventCallOptions', () {
    test('creates CallOptions with required metadata', () {
      final options = createUiEventCallOptions(
        formReference: UIEvent_FormReference.HOME_SCREEN,
        itemName: UIEvent_ItemName.CONNECT_RECENTS,
      );

      expect(options.metadata, isNotNull);
      expect(
        options.metadata[metadataKeyFormReference],
        UIEvent_FormReference.HOME_SCREEN.value.toString(),
      );
      expect(
        options.metadata[metadataKeyItemName],
        UIEvent_ItemName.CONNECT_RECENTS.value.toString(),
      );
      expect(
        options.metadata[metadataKeyItemType],
        UIEvent_ItemType.CLICK.value.toString(),
      );
      // itemValue should not be present when not specified
      expect(options.metadata.containsKey(metadataKeyItemValue), isFalse);
    });

    test('includes itemValue when specified', () {
      final options = createUiEventCallOptions(
        formReference: UIEvent_FormReference.HOME_SCREEN,
        itemName: UIEvent_ItemName.CONNECT,
        itemValue: UIEvent_ItemValue.COUNTRY,
      );

      expect(
        options.metadata[metadataKeyItemValue],
        UIEvent_ItemValue.COUNTRY.value.toString(),
      );
    });

    test('excludes itemValue when UNSPECIFIED', () {
      final options = createUiEventCallOptions(
        formReference: UIEvent_FormReference.HOME_SCREEN,
        itemName: UIEvent_ItemName.CONNECT,
        itemValue: UIEvent_ItemValue.ITEM_VALUE_UNSPECIFIED,
      );

      expect(options.metadata.containsKey(metadataKeyItemValue), isFalse);
    });

    test('allows custom itemType', () {
      final options = createUiEventCallOptions(
        formReference: UIEvent_FormReference.HOME_SCREEN,
        itemName: UIEvent_ItemName.CONNECT,
        itemType: UIEvent_ItemType.ITEM_TYPE_UNSPECIFIED,
      );

      expect(
        options.metadata[metadataKeyItemType],
        UIEvent_ItemType.ITEM_TYPE_UNSPECIFIED.value.toString(),
      );
    });

    test('creates correct metadata for CONNECT', () {
      final options = createUiEventCallOptions(
        formReference: UIEvent_FormReference.HOME_SCREEN,
        itemName: UIEvent_ItemName.CONNECT,
      );

      expect(
        options.metadata[metadataKeyFormReference],
        '3', // HOME_SCREEN = 3
      );
      expect(
        options.metadata[metadataKeyItemName],
        '1', // CONNECT = 1
      );
      expect(
        options.metadata[metadataKeyItemType],
        '1', // CLICK = 1
      );
    });

    test('creates correct metadata for CONNECT_RECENTS', () {
      final options = createUiEventCallOptions(
        formReference: UIEvent_FormReference.HOME_SCREEN,
        itemName: UIEvent_ItemName.CONNECT_RECENTS,
      );

      expect(
        options.metadata[metadataKeyItemName],
        '2', // CONNECT_RECENTS = 2
      );
    });

    test('creates correct metadata for DISCONNECT', () {
      final options = createUiEventCallOptions(
        formReference: UIEvent_FormReference.HOME_SCREEN,
        itemName: UIEvent_ItemName.DISCONNECT,
      );

      expect(
        options.metadata[metadataKeyItemName],
        '3', // DISCONNECT = 3
      );
    });

    test('creates correct metadata for LOGIN', () {
      final options = createUiEventCallOptions(
        formReference: UIEvent_FormReference.GUI,
        itemName: UIEvent_ItemName.LOGIN,
      );

      expect(
        options.metadata[metadataKeyItemName],
        '4', // LOGIN = 4
      );
    });

    test('creates correct metadata for LOGOUT', () {
      final options = createUiEventCallOptions(
        formReference: UIEvent_FormReference.GUI,
        itemName: UIEvent_ItemName.LOGOUT,
      );

      expect(
        options.metadata[metadataKeyItemName],
        '5', // LOGOUT = 5
      );
    });
  });
}
