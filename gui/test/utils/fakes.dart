import 'package:faker/faker.dart';
import 'package:fixnum/fixnum.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/pb/daemon/account.pb.dart';
import 'package:nordvpn/pb/daemon/config/group.pbenum.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';

AccountResponse fakeAccount() {
  return AccountResponse(
    type: Int64(DaemonStatusCode.success),
    email: faker.internet.email(),
    subscriptionExpiresAt: daemonDateFormat.format(
      DateTime.now().add(const Duration(days: 1)),
    ),
    createdOn: daemonDateFormat.format(
      DateTime.now().subtract(const Duration(days: 365)),
    ),
    username: faker.internet.userName(),
  );
}

VpnStatus fakeVpnStatus({
  ConnectionState status = ConnectionState.DISCONNECTED,
  VpnProtocol protocol = VpnProtocol.nordlynx,
  ServerGroup group = ServerGroup.UNDEFINED,
  bool isObfuscated = false,
}) {
  return VpnStatus(
    ip: null,
    hostname: null,
    city: null,
    country: null,
    status: status,
    protocol: protocol,
    isVirtualLocation: false,
    isObfuscated: isObfuscated,
    connectionParameters: ConnectionParameters(
      source: ConnectionSource.MANUAL,
      group: group,
    ),
    isMeshnetRouting: false,
  );
}
