import 'package:flutter/material.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/theme/connection_card_theme.dart';

final class ConnectionCardServerInfo extends StatelessWidget {
  static const textKey = Key("vpnServerInfoText");
  final VpnStatus vpnStatus;

  const ConnectionCardServerInfo({super.key, required this.vpnStatus});

  @override
  Widget build(BuildContext context) {
    final statusCardTheme = context.connectionCardTheme;

    return Text(
      _buildServerInfoLabel(),
      key: ConnectionCardServerInfo.textKey,
      style: statusCardTheme.primaryFont,
      overflow: TextOverflow.ellipsis,
    );
  }

  String _buildServerInfoLabel() {
    if (vpnStatus.isDisconnected()) {
      return t.ui.fastestServer;
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
}
