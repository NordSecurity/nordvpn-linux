import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/recommended_server_provider.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/servers.pb.dart';
import 'package:nordvpn/theme/connection_card_theme.dart';

final class ConnectionCardServerInfo extends ConsumerWidget {
  static const textKey = Key("vpnServerInfoText");
  final VpnStatus vpnStatus;

  const ConnectionCardServerInfo({super.key, required this.vpnStatus});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final statusCardTheme = context.connectionCardTheme;

    final recommendedServerLocation = ref
        .watch(recommendedServerProvider)
        .when(
          data: (fastest) => fastest,
          error: (_, _) => null,
          loading: () => null,
        );

    return Text(
      _buildServerInfoLabel(recommendedServerLocation),
      key: ConnectionCardServerInfo.textKey,
      style: statusCardTheme.primaryFont,
      overflow: TextOverflow.ellipsis,
    );
  }

  String _buildServerInfoLabel(
    RecommendedServerLocation? fastestServerLocation,
  ) {
    if (vpnStatus.isDisconnected()) {
      return _buildDisconnectedServerInfo(fastestServerLocation);
    }

    if (vpnStatus.isConnected()) {
      logger.w("Status is connected, but we don't know to what we connected");
      assert(
        vpnStatus.isMeshnetRouting ||
            (vpnStatus.country != null && vpnStatus.city != null),
        "Status is connected, but we don't know to what we connected",
      );
    }

    if (vpnStatus.isMeshnetRouting) {
      return vpnStatus.hostname ?? vpnStatus.ip ?? "";
    }

    if (vpnStatus.country == null) return t.ui.fastestServer;

    final city = vpnStatus.city != null ? "${vpnStatus.city!}, " : "";
    final virtual = vpnStatus.isVirtualLocation ? " - ${t.ui.virtual}" : "";
    return "$city${vpnStatus.country!.localizedName}$virtual";
  }

  String _buildDisconnectedServerInfo(
    RecommendedServerLocation? recommendedServerLocation,
  ) {
    final country = recommendedServerLocation?.countryName ?? "";
    final city = recommendedServerLocation?.cityName ?? "";
    if (country != "" && city != "") {
      return "$city, $country";
    }

    return t.ui.fastestServer;
  }
}
