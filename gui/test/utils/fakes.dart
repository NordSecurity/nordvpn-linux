import 'package:faker/faker.dart';
import 'package:fixnum/fixnum.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/pb/daemon/account.pb.dart';

AccountResponse fakeAccount() {
  return AccountResponse(
    type: Int64(DaemonStatusCode.success),
    email: faker.internet.email(),
    expiresAt: daemonDateFormat.format(
      DateTime.now().add(const Duration(days: 1)),
    ),
    username: faker.internet.userName(),
  );
}
