import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/vpn_status_card_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

final class ConnectionCardButtons extends ConsumerWidget {
  static const disconnectButtonKey = Key("vpnDisconnectButton");
  static const secureMyConnectionButtonKey = Key("vpnSecureMyConnectionButton");

  final VpnStatus vpnStatus;

  const ConnectionCardButtons({super.key, required this.vpnStatus});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final appTheme = context.appTheme;
    final connectionCardTheme = context.vpnStatusCardTheme;

    return ScalerResponsiveBox(
      maxWidth: connectionCardTheme.maxConnectButtonWidth,
      child: IntrinsicHeight(
        child: Row(
          spacing: appTheme.horizontalSpaceSmall,
          children: _buildButtons(context, ref, appTheme, vpnStatus),
        ),
      ),
    );
  }

  List<Widget> _buildButtons(
    BuildContext context,
    WidgetRef ref,
    AppTheme appTheme,
    VpnStatus status,
  ) {
    final statusCardTheme = context.vpnStatusCardTheme;
    final settings = ref.watch(vpnSettingsControllerProvider).valueOrNull;
    if (status.isConnected()) {
      return [
        Expanded(
          child: OutlinedButton(
            key: ConnectionCardButtons.disconnectButtonKey,
            onPressed: () =>
                ref.read(vpnStatusControllerProvider.notifier).disconnect(),
            child: Text(t.ui.disconnect),
          ),
        ),
        if (!status.isMeshnetRouting)
          OutlinedButton(
            style: OutlinedButton.styleFrom(padding: EdgeInsets.all(0)),
            onPressed: () => _reconnect(ref, status, settings),
            child: DynamicThemeImage("reconnect.svg"),
          ),
      ];
    }

    return [
      Expanded(
        child: ElevatedButton(
          key: ConnectionCardButtons.secureMyConnectionButtonKey,
          onPressed: () async {
            if (status.isDisconnected()) {
              // Quick connect
              ConnectArguments? args;
              if (settings?.obfuscatedServers == true) {
                args = ConnectArguments();
              }
              ref.read(vpnStatusControllerProvider.notifier).connect(args);
            } else if (status.isConnecting()) {
              ref.read(vpnStatusControllerProvider.notifier).cancelConnect();
            }
          },
          style: statusCardTheme.secureMyConnectionButtonStyle,
          child: Text(
            status.isConnecting() ? t.ui.cancel : t.ui.secureMyConnection,
          ),
        ),
      ),
    ];
  }

  Future<void> _reconnect(
    WidgetRef ref,
    VpnStatus status,
    ApplicationSettings? settings,
  ) async {
    if (settings?.obfuscatedServers == true) {
      status.connectionParameters.group = ServerType.obfuscated.toServerGroup();
    }
    ref
        .read(vpnStatusControllerProvider.notifier)
        .reconnect(status.connectionParameters);
  }
}
