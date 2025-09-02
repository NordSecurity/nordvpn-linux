import 'package:freezed_annotation/freezed_annotation.dart';
import 'package:nordvpn/data/models/servers_list.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/internal/dates_util.dart';
import 'package:nordvpn/pb/daemon/account.pb.dart';

part 'user_account.freezed.dart';

@freezed
abstract class UserAccount with _$UserAccount {
  const UserAccount._();

  const factory UserAccount({
    required bool hasDipSubscription,
    required String name,
    required String email,
    required DateTime? vpnExpirationDate,
    required List<CountryServersGroup>? dedicatedIpServers,
  }) = _UserAccount;

  factory UserAccount.fromResponse(AccountResponse response) {
    return UserAccount.from(response, null);
  }

  factory UserAccount.from(
    AccountResponse response,
    List<CountryServersGroup>? dedicatedIpServers,
  ) {
    return UserAccount(
      hasDipSubscription:
          response.dedicatedIpStatus == DaemonStatusCode.success,
      name: response.username,
      email: response.email,
      dedicatedIpServers: dedicatedIpServers,
      vpnExpirationDate: parseDate(response.expiresAt),
    );
  }

  bool get hasDipServers => dedicatedIpServers?.isNotEmpty ?? false;

  bool get isExpired => vpnExpirationDate?.isBefore(DateTime.now()) ?? false;
}
