import 'package:flutter/material.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/theme/connection_card_theme.dart';

final class ConnectionCardServerInfo extends StatelessWidget {
  static const textKey = Key("vpnServerInfoText");
  final VpnStatus vpnStatus;

  const ConnectionCardServerInfo({super.key, required this.vpnStatus});

  @override
  Widget build(BuildContext context) {
    final statusCardTheme = context.connectionCardTheme;
    var label = t.ui.connectToVpn;

    if (vpnStatus.isDisconnected()) {
      label = t.ui.fastestServer;
    } else {
      if (vpnStatus.isConnected()) {
        assert(
          vpnStatus.isMeshnetRouting ||
              (vpnStatus.country != null && vpnStatus.city != null),
          "Status is connected, but we don't know to what we connected",
        );
      }

      if (vpnStatus.isMeshnetRouting) {
        label = vpnStatus.hostname ?? vpnStatus.ip ?? "";
      } else if (vpnStatus.country != null) {
        final countryName = vpnStatus.country!.localizedName;
        if (vpnStatus.city != null) {
          label = "${vpnStatus.city!}, $countryName";
        } else {
          label = countryName;
        }
        label += vpnStatus.isVirtualLocation ? " - ${t.ui.virtual}" : "";
      } else {
        label = t.ui.fastestServer;
      }
    }

    return Text(
      label,
      key: ConnectionCardServerInfo.textKey,
      style: statusCardTheme.primaryFont,
      overflow: TextOverflow.ellipsis,
    );
  }
}
