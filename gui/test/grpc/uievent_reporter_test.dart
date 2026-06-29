import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/grpc/uievent_reporter.dart';
import 'package:nordvpn/pb/daemon/uievent.pb.dart';

void main() {
  group('buildUIEvent', () {
    test('sets required fields', () {
      final e = buildUIEvent(
        formReference: UIEvent_FormReference.HOME_SCREEN,
        itemName: UIEvent_ItemName.CONNECT,
      );

      expect(e.formReference, UIEvent_FormReference.HOME_SCREEN);
      expect(e.itemName, UIEvent_ItemName.CONNECT);
      expect(e.itemType, UIEvent_ItemType.CLICK);
    });

    test('includes itemValue when specified', () {
      final e = buildUIEvent(
        formReference: UIEvent_FormReference.HOME_SCREEN,
        itemName: UIEvent_ItemName.CONNECT,
        itemValue: UIEvent_ItemValue.COUNTRY,
      );

      expect(e.itemValue, UIEvent_ItemValue.COUNTRY);
    });

    test('excludes itemValue when UNSPECIFIED', () {
      final e = buildUIEvent(
        formReference: UIEvent_FormReference.HOME_SCREEN,
        itemName: UIEvent_ItemName.CONNECT,
        itemValue: UIEvent_ItemValue.ITEM_VALUE_UNSPECIFIED,
      );

      expect(e.itemValue, UIEvent_ItemValue.ITEM_VALUE_UNSPECIFIED);
    });

    test('defaults itemValue to UNSPECIFIED when null', () {
      final e = buildUIEvent(
        formReference: UIEvent_FormReference.HOME_SCREEN,
        itemName: UIEvent_ItemName.CONNECT,
      );

      expect(e.itemValue, UIEvent_ItemValue.ITEM_VALUE_UNSPECIFIED);
    });

    test('allows custom itemType', () {
      final e = buildUIEvent(
        formReference: UIEvent_FormReference.HOME_SCREEN,
        itemName: UIEvent_ItemName.CONNECT,
        itemType: UIEvent_ItemType.ITEM_TYPE_UNSPECIFIED,
      );

      expect(e.itemType, UIEvent_ItemType.ITEM_TYPE_UNSPECIFIED);
    });
  });

  // Inventory tests matching exact repository call sites.
  group('GUI event inventory — repository call sites', () {
    void expectEvent(
      String description, {
      required UIEvent_FormReference formReference,
      required UIEvent_ItemName itemName,
      UIEvent_ItemValue? itemValue,
    }) {
      test(description, () {
        final e = buildUIEvent(
          formReference: formReference,
          itemName: itemName,
          itemValue: itemValue,
        );

        expect(e.formReference, formReference);
        expect(e.itemName, itemName);
        expect(e.itemType, UIEvent_ItemType.CLICK);
        if (itemValue != null &&
            itemValue != UIEvent_ItemValue.ITEM_VALUE_UNSPECIFIED) {
          expect(e.itemValue, itemValue);
        }
      });
    }

    // vpn_repository.dart: connect()
    expectEvent(
      'connect — HOME_SCREEN / CONNECT / COUNTRY',
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: UIEvent_ItemName.CONNECT,
      itemValue: UIEvent_ItemValue.COUNTRY,
    );

    // vpn_repository.dart: reconnect()
    expectEvent(
      'reconnect — CONNECTION_INFO / RECONNECT',
      formReference: UIEvent_FormReference.CONNECTION_INFO,
      itemName: UIEvent_ItemName.RECONNECT,
    );

    // vpn_repository.dart: changeSettings()
    expectEvent(
      'changeSettings — CONNECTION_INFO / CHANGE_SETTINGS',
      formReference: UIEvent_FormReference.CONNECTION_INFO,
      itemName: UIEvent_ItemName.CHANGE_SETTINGS,
    );

    // vpn_repository.dart: getHelp()
    expectEvent(
      'getHelp — CONNECTION_INFO / GET_HELP',
      formReference: UIEvent_FormReference.CONNECTION_INFO,
      itemName: UIEvent_ItemName.GET_HELP,
    );

    // vpn_repository.dart: disconnect()
    expectEvent(
      'disconnect — HOME_SCREEN / PAUSE / PAUSE_DISCONNECT',
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: UIEvent_ItemName.PAUSE,
      itemValue: UIEvent_ItemValue.PAUSE_DISCONNECT,
    );

    // vpn_repository.dart: pauseConnection(PauseLength.mins5)
    expectEvent(
      'pause 5 min — HOME_SCREEN / PAUSE / PAUSE_5_MIN',
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: UIEvent_ItemName.PAUSE,
      itemValue: UIEvent_ItemValue.PAUSE_5_MIN,
    );

    // vpn_repository.dart: pauseConnection(PauseLength.mins15)
    expectEvent(
      'pause 15 min — HOME_SCREEN / PAUSE / PAUSE_15_MIN',
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: UIEvent_ItemName.PAUSE,
      itemValue: UIEvent_ItemValue.PAUSE_15_MIN,
    );

    // vpn_repository.dart: pauseConnection(PauseLength.mins30)
    expectEvent(
      'pause 30 min — HOME_SCREEN / PAUSE / PAUSE_30_MIN',
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: UIEvent_ItemName.PAUSE,
      itemValue: UIEvent_ItemValue.PAUSE_30_MIN,
    );

    // vpn_repository.dart: pauseConnection(PauseLength.hour1)
    expectEvent(
      'pause 1 hour — HOME_SCREEN / PAUSE / PAUSE_1_HOUR',
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: UIEvent_ItemName.PAUSE,
      itemValue: UIEvent_ItemValue.PAUSE_1_HOUR,
    );

    // account_repository.dart: _doLogin()
    expectEvent(
      'login — GUI / LOGIN',
      formReference: UIEvent_FormReference.GUI,
      itemName: UIEvent_ItemName.LOGIN,
    );

    // account_repository.dart: logout()
    expectEvent(
      'logout — GUI / LOGOUT',
      formReference: UIEvent_FormReference.GUI,
      itemName: UIEvent_ItemName.LOGOUT,
    );
  });
}
