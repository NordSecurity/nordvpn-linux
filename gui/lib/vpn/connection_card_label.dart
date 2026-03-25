import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/server_group_extension.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/i18n/string_translation_extension.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/vpn_status_card_theme.dart';

final class ConnectionCardLabel extends ConsumerWidget {
  static const labelKey = Key("vpnStatusLabelText");

  final VpnStatus vpnStatus;

  const ConnectionCardLabel({super.key, required this.vpnStatus});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final appTheme = context.appTheme;
    final statusCardTheme = context.vpnStatusCardTheme;
    final settings = ref.watch(vpnSettingsControllerProvider).valueOrNull;

    return Text(
      _constructLabel(settings),
      key: ConnectionCardLabel.labelKey,
      overflow: TextOverflow.ellipsis,
      style: statusCardTheme.secondaryFont.copyWith(
        color: vpnStatus.isDisconnected() || vpnStatus.isConnecting()
            ? appTheme.textErrorColor
            : appTheme.successColor,
      ),
    );
  }

  String _constructLabel(ApplicationSettings? settings) {
    var connectionStatus = t.ui.notSecured;
    if (vpnStatus.isAutoConnected()) {
      connectionStatus = t.ui.autoConnected;
    } else if (vpnStatus.isConnected()) {
      connectionStatus = vpnStatus.isMeshnetRouting
          ? t.ui.meshnet
          : t.ui.connected;
    } else if (vpnStatus.isConnecting()) {
      connectionStatus = t.ui.connecting;
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

    if (vpnStatus.isConnecting()) {
      connectionStatus += "...";
    }

    return connectionStatus;
  }
}
