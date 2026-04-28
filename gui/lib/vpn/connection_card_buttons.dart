import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/pause.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/connection_card_theme.dart';
import 'package:nordvpn/widgets/context_menu/context_menu.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

final class ConnectionCardButtons extends ConsumerWidget {
  static const secureMyConnectionButtonKey = Key("vpnSecureMyConnectionButton");
  static const cancelButtonKey = Key("vpnCancelButton");
  static const pauseConnectionButtonKey = Key("pauseConnectionButton");
  static const disconnectMenuItemKey = Key("disconnectMenuItem");
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
      if (status.isMeshnetRouting) {
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
        ];
      }
      final pauseOptions = <({String label, PauseLength pause})>[
        (label: t.ui.pauseFor5Min, pause: PauseLength.mins5),
        (label: t.ui.pauseFor15Min, pause: PauseLength.mins15),
        (label: t.ui.pauseFor30Min, pause: PauseLength.mins30),
        (label: t.ui.pauseFor1Hour, pause: PauseLength.hour1),
        (label: t.ui.pauseFor24Hours, pause: PauseLength.hours24),
      ];
      return [
        Expanded(
          child: ContextMenu(
            key: ConnectionCardButtons.pauseConnectionButtonKey,
            matchAnchorWidth: true,
            items: [
              ...pauseOptions.map(
                (opt) => ContextMenuItem(
                  label: opt.label,
                  onTap: () async => await _pauseConnection(ref, opt.pause),
                ),
              ),
              ContextMenuItem(
                key: ConnectionCardButtons.disconnectMenuItemKey,
                label: t.ui.disconnect,
                labelColor: context.appTheme.textErrorColor,
                onTap: () async => await ref
                    .read(vpnStatusControllerProvider.notifier)
                    .disconnect(),
              ),
            ],
            anchorBuilder: (toggleMenu) => ElevatedButton(
              style: buttonTheme.pauseConnectionButtonStyle,
              onPressed: toggleMenu,
              child: Text(t.ui.pauseConnection),
            ),
          ),
        ),
        IntrinsicWidth(
          child: ContextMenu(
            items: [
              ContextMenuItem(
                label: t.ui.reconnect,
                onTap: () async => await _reconnect(ref, status, settings),
              ),
              ContextMenuItem(
                label: t.ui.changeVPNsettings,
                onTap: () =>
                    context.navigateToRoute(AppRoute.settingsVpnConnection),
              ),
              ContextMenuItem(
                label: t.ui.getHelp,
                onTap: () => getHelpUrl.launch(),
              ),
            ],
            anchorBuilder: (toggleMenu) => ElevatedButton(
              style: buttonTheme.connectionDetailsButtonStyle,
              onPressed: toggleMenu,
              child: DynamicThemeImage("connection_details.svg"),
            ),
          ),
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
    await ref
        .read(vpnStatusControllerProvider.notifier)
        .reconnect(status.connectionParameters);
  }

  Future<void> _pauseConnection(WidgetRef ref, PauseLength pauseLength) async {
    ref.read(vpnStatusControllerProvider.notifier).pauseConnection(pauseLength);
  }
}
