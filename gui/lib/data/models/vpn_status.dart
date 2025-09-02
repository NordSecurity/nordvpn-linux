import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';
import 'package:freezed_annotation/freezed_annotation.dart';

part "vpn_status.freezed.dart";

@freezed
abstract class VpnStatus with _$VpnStatus {
  const VpnStatus._();

  const factory VpnStatus({
    required String? ip,
    required String? hostname,
    required City? city,
    required Country? country,
    required ConnectionState status,
    required VpnProtocol protocol,
    required bool isVirtualLocation,
    required ConnectionParameters connectionParameters,
    required bool isMeshnetRouting,
  }) = _VpnStatus;

  factory VpnStatus.fromStatusResponse(StatusResponse status) {
    final vpnStatus = VpnStatus(
      ip: status.ip,
      hostname: status.hostname,
      city: status.city.isNotEmpty ? City(status.city) : null,
      country: status.country.isNotEmpty
          ? Country.fromCode(status.country)
          : null,
      status: status.state,
      protocol: convertToVpnProtocol(status.technology, status.protocol),
      isVirtualLocation: status.virtualLocation,
      connectionParameters: status.parameters,
      isMeshnetRouting: status.isMeshPeer,
    );

    assert(
      vpnStatus.isMeshnetRouting ||
          vpnStatus.isConnecting() ||
          (vpnStatus.isConnected() && vpnStatus.country != null) ||
          (vpnStatus.isDisconnected() && vpnStatus.country == null),
    );

    return vpnStatus;
  }

  bool isConnected() => status == ConnectionState.CONNECTED;
  bool isAutoConnected() =>
      isConnected() && connectionParameters.source == ConnectionSource.AUTO;
  bool isConnecting() => status == ConnectionState.CONNECTING;
  bool isDisconnected() => status == ConnectionState.DISCONNECTED;

  // Don't check all the members because they might not be relevant, e.g. connection duration
  bool isEqualToStatusResponse(StatusResponse statusResponse) {
    return status == statusResponse.state &&
        ip == statusResponse.ip &&
        country?.name == statusResponse.country &&
        city?.name == statusResponse.city &&
        hostname == statusResponse.hostname &&
        protocol ==
            convertToVpnProtocol(
              statusResponse.technology,
              statusResponse.protocol,
            );
  }
}
