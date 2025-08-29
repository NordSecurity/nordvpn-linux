import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/user_account.dart';
import 'package:nordvpn/pb/daemon/account.pb.dart';

void main() {
  group('Date is parsed correctly from AccountResponse', () {
    test('Works with 1 digit for month and day', () async {
      final userAccount = UserAccount.from(
        AccountResponse(expiresAt: "2022-2-5 08:44:20"),
        [],
      );
      expect(userAccount.vpnExpirationDate!.day, 5);
      expect(userAccount.vpnExpirationDate!.month, 2);
      expect(userAccount.vpnExpirationDate!.year, 2022);
    });

    test('Works with 1 digit for month and 2 digits for day', () async {
      final userAccount = UserAccount.from(
        AccountResponse(expiresAt: "2022-9-15 23:1:0"),
        [],
      );
      expect(userAccount.vpnExpirationDate!.day, 15);
      expect(userAccount.vpnExpirationDate!.month, 9);
      expect(userAccount.vpnExpirationDate!.year, 2022);
    });

    test('Works with 2 digits for month and 2 for day', () async {
      final userAccount = UserAccount.from(
        AccountResponse(expiresAt: "2022-12-25 1:1:1"),
        [],
      );
      expect(userAccount.vpnExpirationDate!.day, 25);
      expect(userAccount.vpnExpirationDate!.month, 12);
      expect(userAccount.vpnExpirationDate!.year, 2022);
    });

    test('Works with 2 digits for month and 1 for day', () async {
      final userAccount = UserAccount.from(
        AccountResponse(expiresAt: "2022-12-2 1:1:1"),
        [],
      );
      expect(userAccount.vpnExpirationDate!.day, 2);
      expect(userAccount.vpnExpirationDate!.month, 12);
      expect(userAccount.vpnExpirationDate!.year, 2022);
    });

    test('Works with 2 digits for month and 2 for day', () async {
      final userAccount = UserAccount.from(
        AccountResponse(expiresAt: "2022-12-25 8:1:2"),
        [],
      );
      expect(userAccount.vpnExpirationDate!.day, 25);
      expect(userAccount.vpnExpirationDate!.month, 12);
      expect(userAccount.vpnExpirationDate!.year, 2022);
    });
  });

  test('Fails to parse invalid date', () async {
    final userAccount = UserAccount.from(
      AccountResponse(expiresAt: "2022-25-40 1:2:3"),
      [],
    );
    expect(userAccount.vpnExpirationDate, isNull);
  });
}
