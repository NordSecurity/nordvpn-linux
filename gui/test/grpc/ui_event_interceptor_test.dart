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

    test('creates correct metadata for PAUSE', () {
      final options = createUiEventCallOptions(
        formReference: UIEvent_FormReference.GUI,
        itemName: UIEvent_ItemName.PAUSE,
      );

      expect(
        options.metadata[metadataKeyItemName],
        '9', // PAUSE = 9
      );
    });
  });

  // To run regression tests:
  // cd gui && flutter test test/grpc/ui_event_interceptor_test.dart 2>&1

  // Inventory tests matching exact repository call sites.
  // Run before and after UI event refactoring to verify equivalence.
  group('GUI event inventory — repository call sites', () {
    void expectMetadata(
      String description, {
      required UIEvent_FormReference formReference,
      required UIEvent_ItemName itemName,
      UIEvent_ItemValue? itemValue,
      required String expectedFormRef,
      required String expectedItemName,
      String? expectedItemValue,
    }) {
      test(description, () {
        final options = createUiEventCallOptions(
          formReference: formReference,
          itemName: itemName,
          itemValue: itemValue,
        );

        expect(options.metadata[metadataKeyFormReference], expectedFormRef);
        expect(options.metadata[metadataKeyItemName], expectedItemName);
        expect(
          options.metadata[metadataKeyItemType],
          '1', // CLICK
        );
        if (expectedItemValue != null) {
          expect(options.metadata[metadataKeyItemValue], expectedItemValue);
        } else {
          expect(options.metadata.containsKey(metadataKeyItemValue), isFalse);
        }
      });
    }

    // vpn_repository.dart: reconnect()
    expectMetadata(
      'reconnect — CONNECTION_INFO / RECONNECT',
      formReference: UIEvent_FormReference.CONNECTION_INFO,
      itemName: UIEvent_ItemName.RECONNECT,
      expectedFormRef: '5',
      expectedItemName: '10',
    );

    // vpn_repository.dart: changeSettings()
    expectMetadata(
      'changeSettings — CONNECTION_INFO / CHANGE_SETTINGS',
      formReference: UIEvent_FormReference.CONNECTION_INFO,
      itemName: UIEvent_ItemName.CHANGE_SETTINGS,
      expectedFormRef: '5',
      expectedItemName: '11',
    );

    // vpn_repository.dart: getHelp()
    expectMetadata(
      'getHelp — CONNECTION_INFO / GET_HELP',
      formReference: UIEvent_FormReference.CONNECTION_INFO,
      itemName: UIEvent_ItemName.GET_HELP,
      expectedFormRef: '5',
      expectedItemName: '12',
    );

    // vpn_repository.dart: disconnect()
    expectMetadata(
      'disconnect — HOME_SCREEN / PAUSE / PAUSE_DISCONNECT',
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: UIEvent_ItemName.PAUSE,
      itemValue: UIEvent_ItemValue.PAUSE_DISCONNECT,
      expectedFormRef: '3',
      expectedItemName: '9',
      expectedItemValue: '14',
    );

    // vpn_repository.dart: pauseConnection(PauseLength.mins5)
    expectMetadata(
      'pause 5 min — HOME_SCREEN / PAUSE / PAUSE_5_MIN',
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: UIEvent_ItemName.PAUSE,
      itemValue: UIEvent_ItemValue.PAUSE_5_MIN,
      expectedFormRef: '3',
      expectedItemName: '9',
      expectedItemValue: '9',
    );

    // vpn_repository.dart: pauseConnection(PauseLength.mins15)
    expectMetadata(
      'pause 15 min — HOME_SCREEN / PAUSE / PAUSE_15_MIN',
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: UIEvent_ItemName.PAUSE,
      itemValue: UIEvent_ItemValue.PAUSE_15_MIN,
      expectedFormRef: '3',
      expectedItemName: '9',
      expectedItemValue: '10',
    );

    // vpn_repository.dart: pauseConnection(PauseLength.mins30)
    expectMetadata(
      'pause 30 min — HOME_SCREEN / PAUSE / PAUSE_30_MIN',
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: UIEvent_ItemName.PAUSE,
      itemValue: UIEvent_ItemValue.PAUSE_30_MIN,
      expectedFormRef: '3',
      expectedItemName: '9',
      expectedItemValue: '11',
    );

    // vpn_repository.dart: pauseConnection(PauseLength.hour1)
    expectMetadata(
      'pause 1 hour — HOME_SCREEN / PAUSE / PAUSE_1_HOUR',
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: UIEvent_ItemName.PAUSE,
      itemValue: UIEvent_ItemValue.PAUSE_1_HOUR,
      expectedFormRef: '3',
      expectedItemName: '9',
      expectedItemValue: '12',
    );
  });
}
