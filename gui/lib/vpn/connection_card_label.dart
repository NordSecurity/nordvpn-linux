import 'package:flutter/material.dart';
import 'package:nordvpn/data/models/server_group_extension.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/i18n/string_translation_extension.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/theme/connection_card_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

final class ConnectionCardLabel extends StatelessWidget {
  static const labelKey = Key("vpnStatusLabelText");

  final VpnStatus vpnStatus;

  const ConnectionCardLabel({super.key, required this.vpnStatus});

  @override
  Widget build(BuildContext context) {
    final labelTheme = context.connectionCardTheme.labelTheme;

    return Row(
      key: labelKey,
      spacing: labelTheme.spacing,
      children: [
        _buildLabel(labelTheme),
        ..._addServerTypeIfNeeded(labelTheme),
      ],
    );
  }

  Widget _buildLabel(ConnectionCardLabelTheme labelTheme) {
    var connectionStatus = t.ui.notSecured;
    if (vpnStatus.isAutoConnected()) {
      connectionStatus = t.ui.autoConnected;
    } else if (vpnStatus.isConnected()) {
      connectionStatus = vpnStatus.isMeshnetRouting
          ? t.ui.connected
          : t.ui.secured;
    } else if (vpnStatus.isConnecting()) {
      connectionStatus = "${t.ui.connecting}...";
    }

    return Text(
      connectionStatus,
      overflow: TextOverflow.ellipsis,
      style: labelTheme.font.copyWith(color: _labelColor(labelTheme)),
    );
  }

  List<Widget> _addServerTypeIfNeeded(ConnectionCardLabelTheme labelTheme) {
    if (!vpnStatus.isConnected()) {
      return [];
    }

    var serverType = "";
    if (vpnStatus.isObfuscated) {
      serverType = labelForServerType(ServerType.obfuscated);
    } else {
      final serverGroup = vpnStatus.connectionParameters.group
          .toSpecialtyType();
      if (serverGroup != null &&
          serverGroup != ServerType.standardVpn &&
          serverGroup != ServerType.p2p) {
        serverType = labelForServerType(serverGroup);
      }
    }

    if (serverType == "") {
      return [];
    }

    return [
      DynamicThemeImage("dot.svg"),
      Text(
        serverType,
        overflow: TextOverflow.ellipsis,
        style: labelTheme.font.copyWith(color: labelTheme.serverTypeColor),
      ),
    ];
  }

  Color _labelColor(ConnectionCardLabelTheme labelTheme) {
    if (vpnStatus.isConnected()) {
      return labelTheme.connectedColor;
    } else if (vpnStatus.isConnecting()) {
      return labelTheme.connectingColor;
    } else {
      return labelTheme.disconnectedColor;
    }
  }
}
