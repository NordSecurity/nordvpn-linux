import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/pb/daemon/uievent.pbenum.dart';

void main() {
  group('toUIEventItemValue', () {
    test('returns ITEM_VALUE_UNSPECIFIED for empty arguments', () {
      final args = ConnectArguments();
      expect(args.toUIEventItemValue(), UIEvent_ItemValue.ITEM_VALUE_UNSPECIFIED);
    });

    test('returns COUNTRY when only country is specified', () {
      final args = ConnectArguments(country: Country(code: 'US', name: 'United States'));
      expect(args.toUIEventItemValue(), UIEvent_ItemValue.COUNTRY);
    });

    test('returns CITY when city is specified', () {
      final args = ConnectArguments(
        country: Country(code: 'US', name: 'United States'),
        city: City('New York'),
      );
      expect(args.toUIEventItemValue(), UIEvent_ItemValue.CITY);
    });

    test('returns DIP for dedicated IP specialty group', () {
      final args = ConnectArguments(specialtyGroup: ServerType.dedicatedIP);
      expect(args.toUIEventItemValue(), UIEvent_ItemValue.DIP);
    });

    test('returns OBFUSCATED for obfuscated specialty group', () {
      final args = ConnectArguments(specialtyGroup: ServerType.obfuscated);
      expect(args.toUIEventItemValue(), UIEvent_ItemValue.OBFUSCATED);
    });

    test('returns ONION_OVER_VPN for onion over VPN specialty group', () {
      final args = ConnectArguments(specialtyGroup: ServerType.onionOverVpn);
      expect(args.toUIEventItemValue(), UIEvent_ItemValue.ONION_OVER_VPN);
    });

    test('returns DOUBLE_VPN for double VPN specialty group', () {
      final args = ConnectArguments(specialtyGroup: ServerType.doubleVpn);
      expect(args.toUIEventItemValue(), UIEvent_ItemValue.DOUBLE_VPN);
    });

    test('returns P2P for P2P specialty group', () {
      final args = ConnectArguments(specialtyGroup: ServerType.p2p);
      expect(args.toUIEventItemValue(), UIEvent_ItemValue.P2P);
    });

    test('specialty group takes priority over city', () {
      final args = ConnectArguments(
        country: Country(code: 'US', name: 'United States'),
        city: City('New York'),
        specialtyGroup: ServerType.doubleVpn,
      );
      expect(args.toUIEventItemValue(), UIEvent_ItemValue.DOUBLE_VPN);
    });

    test('specialty group takes priority over country', () {
      final args = ConnectArguments(
        country: Country(code: 'US', name: 'United States'),
        specialtyGroup: ServerType.p2p,
      );
      expect(args.toUIEventItemValue(), UIEvent_ItemValue.P2P);
    });

    test('returns ITEM_VALUE_UNSPECIFIED for standard VPN specialty group', () {
      // standardVpn is filtered out by the specialtyGroup getter
      final args = ConnectArguments(specialtyGroup: ServerType.standardVpn);
      expect(args.toUIEventItemValue(), UIEvent_ItemValue.ITEM_VALUE_UNSPECIFIED);
    });
  });
}
