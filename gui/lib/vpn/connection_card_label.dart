import 'package:flutter/material.dart';
import 'package:nordvpn/data/models/server_group_extension.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/i18n/string_translation_extension.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/theme/vpn_status_card_theme.dart';

final class ConnectionCardLabel extends StatelessWidget {
  static const labelKey = Key("vpnStatusLabelText");

  final VpnStatus vpnStatus;

  const ConnectionCardLabel({super.key, required this.vpnStatus});

  @override
  Widget build(BuildContext context) {
    final connectionCardTheme = context.vpnStatusCardTheme;

    return Text(
      _constructLabel(),
      key: ConnectionCardLabel.labelKey,
      overflow: TextOverflow.ellipsis,
      style: connectionCardTheme.secondaryFont.copyWith(
        color: _labelColor(connectionCardTheme),
      ),
    );
  }

  String _constructLabel() {
    var connectionStatus = t.ui.notSecured;
    if (vpnStatus.isAutoConnected()) {
      connectionStatus = t.ui.autoConnected;
    } else if (vpnStatus.isConnected()) {
      connectionStatus = vpnStatus.isMeshnetRouting
          ? t.ui.meshnet
          : t.ui.connected;
    } else if (vpnStatus.isConnecting()) {
      return "${t.ui.connecting}...";
    }

    // Show obfuscated label if connection is obfuscated
    if (vpnStatus.isObfuscated) {
      connectionStatus +=
          " ${t.ui.to} ${labelForServerType(ServerType.obfuscated)}";
    } else {
      // Otherwise show other specialty server types
      final serverGroup = vpnStatus.connectionParameters.group
          .toSpecialtyType();
      // `standardVpn` is a regular VPN connection - no special label for it.
      if (serverGroup != null && serverGroup != ServerType.standardVpn) {
        connectionStatus += " ${t.ui.to} ${labelForServerType(serverGroup)}";
      }
    }

    return connectionStatus;
  }

  Color _labelColor(VpnStatusCardTheme theme) {
    if (vpnStatus.isConnected()) {
      return theme.labelStyle.connectedColor;
    } else if (vpnStatus.isConnecting()) {
      return theme.labelStyle.connectingColor;
    } else {
      return theme.labelStyle.disconnectedColor;
    }
  }
}
