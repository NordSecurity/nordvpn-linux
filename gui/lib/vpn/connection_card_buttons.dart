import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/pause.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/data/repository/uievent_repository.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/connection_card_theme.dart';
import 'package:nordvpn/widgets/context_menu/context_menu.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

final class ConnectionCardButtons extends ConsumerStatefulWidget {
  static const secureMyConnectionButtonKey = Key("vpnSecureMyConnectionButton");
  static const cancelButtonKey = Key("vpnCancelButton");
  static const pauseConnectionButtonKey = Key("pauseConnectionButton");
  static const disconnectMenuItemKey = Key("disconnectMenuItem");
  static const disconnectButtonKey = Key("vpnDisconnectButton");

  final VpnStatus vpnStatus;

  const ConnectionCardButtons({super.key, required this.vpnStatus});

  @override
  ConsumerState<ConnectionCardButtons> createState() =>
      _ConnectionCardButtonsState();
}

class _ConnectionCardButtonsState extends ConsumerState<ConnectionCardButtons> {
  static const _pauseLengths = [
    PauseLength.mins5,
    PauseLength.mins15,
    PauseLength.mins30,
    PauseLength.hour1,
    PauseLength.hours24,
  ];

  final _primaryButtonFocus = FocusNode(debugLabel: "connectionCardPrimary");

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) _primaryButtonFocus.requestFocus();
    });
  }

  @override
  void didUpdateWidget(covariant ConnectionCardButtons old) {
    super.didUpdateWidget(old);
    if (_stateOf(old.vpnStatus) != _stateOf(widget.vpnStatus)) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && !_primaryButtonFocus.hasFocus) {
          _primaryButtonFocus.requestFocus();
        }
      });
    }
  }

  int _stateOf(VpnStatus s) => s.isConnected()
      ? 2
      : s.isConnecting()
          ? 1
          : 0;

  @override
  void dispose() {
    _primaryButtonFocus.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;
    final buttonTheme = context.connectionCardTheme.buttonTheme;

    return ScalerResponsiveBox(
      maxWidth: buttonTheme.maxConnectButtonWidth,
      child: IntrinsicHeight(
        child: FocusTraversalGroup(
          child: Row(
            spacing: appTheme.horizontalSpaceSmall,
            children: _buildButtons(
              context,
              ref,
              appTheme,
              buttonTheme,
              widget.vpnStatus,
            ),
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
          _buildConnectionDetailsButton(context, ref, buttonTheme, status),
        ];
      }
      return [
        Expanded(
          child: FocusTraversalGroup(
            child: ContextMenu(
              key: ConnectionCardButtons.pauseConnectionButtonKey,
              matchAnchorWidth: true,
              items: [
                ..._pauseLengths.map(
                  (pause) => ContextMenuItem(
                    label: _pauseLabel(pause),
                    onTap: () async => await _pauseConnection(ref, pause),
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
              anchorBuilder: (toggleMenu) => OutlinedButton(
                style: buttonTheme.pauseConnectionButtonStyle,
                focusNode: _primaryButtonFocus,
                onPressed: toggleMenu,
                child: Semantics(
                  label:
                      "${_buildSemanticsText(status)} ${t.ui.pauseConnection}",
                  button: true,
                  enabled: true,
                  excludeSemantics: true,
                  child: Text(t.ui.pauseConnection),
                ),
              ),
            ),
          ),
        ),
        _buildConnectionDetailsButton(
          context,
          ref,
          buttonTheme,
          status,
          extraItems: [
            ContextMenuItem(
              label: t.ui.reconnect,
              onTap: () async => await _reconnect(ref, status, settings),
            ),
          ],
        ),
      ];
    }

    if (status.isConnecting()) {
      return [_buildConnectingStateButton(ref, buttonTheme, status)];
    }

    return [_buildDisconnectedStateButton(ref, buttonTheme, settings, status)];
  }

  Widget _buildDisconnectedStateButton(
    WidgetRef ref,
    ConnectionCardButtonTheme buttonTheme,
    ApplicationSettings? settings,
    VpnStatus status,
  ) {
    return Expanded(
      child: OutlinedButton(
        key: ConnectionCardButtons.secureMyConnectionButtonKey,
        focusNode: _primaryButtonFocus,
        onPressed: () async {
          // Quick connect
          ConnectArguments? args;
          if (settings?.obfuscatedServers == true) {
            args = ConnectArguments();
          }
          await ref.read(vpnStatusControllerProvider.notifier).connect(args);
        },
        style: buttonTheme.secureMyConnectionButtonStyle,
        child: Semantics(
          label: "${_buildSemanticsText(status)} ${t.ui.secureMyConnection}",
          enabled: true,
          button: true,
          excludeSemantics: true,
          child: Text(t.ui.secureMyConnection),
        ),
      ),
    );
  }

  Widget _buildConnectingStateButton(
    WidgetRef ref,
    ConnectionCardButtonTheme buttonTheme,
    VpnStatus status,
  ) {
    return Expanded(
      child: OutlinedButton(
        key: ConnectionCardButtons.cancelButtonKey,
        focusNode: _primaryButtonFocus,
        onPressed: () async {
          await ref.read(vpnStatusControllerProvider.notifier).cancelConnect();
        },
        style: buttonTheme.cancelButtonStyle,
        child: Semantics(
          label: "${_buildSemanticsText(status)} ${t.ui.cancel}",
          enabled: true,
          button: true,
          excludeSemantics: true,
          child: Text(t.ui.cancel),
        ),
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

  void _changeSettings(BuildContext context, WidgetRef ref) {
    context.navigateToRoute(AppRoute.settingsVpnConnection);
    ref.read(uiEventRepositoryProvider).reportChangeSettings();
  }

  void _getHelp(WidgetRef ref) {
    getHelpUrl.launch();
    ref.read(uiEventRepositoryProvider).reportGetHelp();
  }

  static String _pauseLabel(PauseLength pause) => switch (pause) {
    PauseLength.mins5 => t.ui.pauseFor5Min,
    PauseLength.mins15 => t.ui.pauseFor15Min,
    PauseLength.mins30 => t.ui.pauseFor30Min,
    PauseLength.hour1 => t.ui.pauseFor1Hour,
    PauseLength.hours24 => t.ui.pauseFor24Hours,
  };

  Widget _buildConnectionDetailsButton(
    BuildContext context,
    WidgetRef ref,
    ConnectionCardButtonTheme buttonTheme,
    VpnStatus status, {
    List<ContextMenuItem> extraItems = const [],
  }) {
    return IntrinsicWidth(
      child: FocusTraversalGroup(
        child: ContextMenu(
          items: [
            ...extraItems,
            ContextMenuItem(
              label: t.ui.changeVPNsettings,
              onTap: () => _changeSettings(context, ref),
            ),
            ContextMenuItem(label: t.ui.getHelp, onTap: () => _getHelp(ref)),
          ],
          anchorBuilder: (toggleMenu) => OutlinedButton(
            style: buttonTheme.connectionDetailsButtonStyle,
            onPressed: toggleMenu,
            child: Semantics(
              label: "${_buildSemanticsText(status)} ${t.ui.more}",
              button: true,
              enabled: true,
              excludeSemantics: true,
              child: DynamicThemeImage("connection_details.svg"),
            ),
          ),
        ),
      ),
    );
  }

  String _buildSemanticsText(VpnStatus vpnStatus) {
    // VPN Panel. Preferred location: Fastest Server. Not secured. Secure my connection push button.
    // VPN Panel. Connecting to Fastest Server. Cancel push button.
    // VPN Panel. Connected to [City], [Country]. Pause menu push button.

    var vpnPanel = "${t.ui.vpnPanel}. ";
    if (vpnStatus.isDisconnected()) {
      return "$vpnPanel ${t.ui.preferredLocation} ${t.ui.fastestServer}. ${t.ui.notSecured}";
    }

    if (vpnStatus.isConnecting()) {
      return "$vpnPanel ${t.ui.connecting} to ${t.ui.fastestServer}.";
    }

    if (vpnStatus.isConnected()) {
      return "$vpnPanel ${t.ui.connected} to ${_buildCityAndCountryText(vpnStatus)}.";
    }

    return "$vpnPanel ${t.ui.loading}";
  }

  String _buildCityAndCountryText(VpnStatus vpnStatus) {
    if (vpnStatus.isMeshnetRouting) {
      return vpnStatus.hostname ?? vpnStatus.ip ?? "";
    }

    if (vpnStatus.country == null) return t.ui.fastestServer;

    final city = vpnStatus.city != null ? "${vpnStatus.city!}, " : "";
    final virtual = vpnStatus.isVirtualLocation ? " ${t.ui.virtual}" : "";
    return "$city${vpnStatus.country!.localizedName}$virtual";
  }
}
