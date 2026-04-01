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
import 'package:nordvpn/theme/connection_card_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

final class ConnectionCardButtons extends ConsumerWidget {
  static const secureMyConnectionButtonKey = Key("vpnSecureMyConnectionButton");
  static const cancelButtonKey = Key("vpnCancelButton");
  static const disconnectButtonKey = Key("vpnDisconnectButton");

  final VpnStatus vpnStatus;

  const ConnectionCardButtons({super.key, required this.vpnStatus});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final appTheme = context.appTheme;
    final buttonTheme = context.connectionCardTheme.buttonTheme;

    return ScalerResponsiveBox(
      maxWidth: buttonTheme.maxConnectButtonWidth,
      child: IntrinsicHeight(
        child: Row(
          spacing: appTheme.horizontalSpaceSmall,
          children: _buildButtons(
            context,
            ref,
            appTheme,
            buttonTheme,
            vpnStatus,
          ),
        ),
      ),
    );
  }

  List<Widget> _buildButtons(
    BuildContext context,
    WidgetRef ref,
    AppTheme appTheme,
    ConnectionCardButtonTheme buttonTheme,
    VpnStatus status,
  ) {
    final settings = ref.watch(vpnSettingsControllerProvider).valueOrNull;
    if (status.isConnected()) {
      return [
        Expanded(
          child: OutlinedButton(
            key: ConnectionCardButtons.disconnectButtonKey,
            style: buttonTheme.cancelButtonStyle,
            onPressed: () async => await ref
                .read(vpnStatusControllerProvider.notifier)
                .disconnect(),
            child: Text(t.ui.disconnect),
          ),
        ),
        if (!status.isMeshnetRouting)
          OutlinedButton(
            style: OutlinedButton.styleFrom(padding: EdgeInsets.all(0)),
            onPressed: () async => await _reconnect(ref, status, settings),
            child: DynamicThemeImage("reconnect.svg"),
          ),
      ];
    }

    if (status.isConnecting()) {
      return [_buildConnectingStateButton(ref, buttonTheme)];
    }

    return [_buildDisconnectedStateButton(ref, buttonTheme, settings)];
  }

  Widget _buildDisconnectedStateButton(
    WidgetRef ref,
    ConnectionCardButtonTheme buttonTheme,
    ApplicationSettings? settings,
  ) {
    return Expanded(
      child: ElevatedButton(
        key: ConnectionCardButtons.secureMyConnectionButtonKey,
        onPressed: () async {
          // Quick connect
          ConnectArguments? args;
          if (settings?.obfuscatedServers == true) {
            args = ConnectArguments();
          }
          await ref.read(vpnStatusControllerProvider.notifier).connect(args);
        },
        style: buttonTheme.secureMyConnectionButtonStyle,
        child: Text(t.ui.secureMyConnection),
      ),
    );
  }

  Widget _buildConnectingStateButton(
    WidgetRef ref,
    ConnectionCardButtonTheme buttonTheme,
  ) {
    return Expanded(
      child: ElevatedButton(
        key: ConnectionCardButtons.cancelButtonKey,
        onPressed: () async {
          await ref.read(vpnStatusControllerProvider.notifier).cancelConnect();
        },
        style: buttonTheme.cancelButtonStyle,
        child: Text(t.ui.cancel),
      ),
    );
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
