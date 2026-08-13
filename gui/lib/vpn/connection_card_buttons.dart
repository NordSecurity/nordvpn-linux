import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/pause.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/recommended_server_provider.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/data/repository/uievent_repository.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/pb/daemon/servers.pb.dart';
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

  static const _pauseLengths = [
    PauseLength.mins5,
    PauseLength.mins15,
    PauseLength.mins30,
    PauseLength.hour1,
    PauseLength.hours24,
  ];

  final VpnStatus vpnStatus;

  const ConnectionCardButtons({super.key, required this.vpnStatus});
  @override
  ConsumerState<ConnectionCardButtons> createState() =>
      _ConnectionCardButtonsState();
}

final class _ConnectionCardButtonsState
    extends ConsumerState<ConnectionCardButtons> {
  late final FocusNode _buttonFocusNode;

  @override
  void initState() {
    super.initState();
    _buttonFocusNode = FocusNode();
  }

  @override
  void dispose() {
    _buttonFocusNode.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;
    final buttonTheme = context.connectionCardTheme.buttonTheme;
    final recommendedServerLocation = ref
        .watch(recommendedServerProvider)
        .when(
          data: (fastest) => fastest,
          error: (_, _) => null,
          loading: () => null,
        );
    return ScalerResponsiveBox(
      maxWidth: buttonTheme.maxConnectButtonWidth,
      child: IntrinsicHeight(
        child: FocusTraversalGroup(
          child: Row(
            spacing: appTheme.horizontalSpaceSmall,
            children: _buildButtons(context, recommendedServerLocation),
          ),
        ),
      ),
    );
  }

  List<Widget> _buildButtons(
    BuildContext context,
    RecommendedServerLocation? recommendedServerLocation,
  ) {
    final settings = ref.watch(vpnSettingsControllerProvider).value;
    final buttonTheme = context.connectionCardTheme.buttonTheme;

    if (widget.vpnStatus.isConnected()) {
      if (widget.vpnStatus.isMeshnetRouting) {
        return [
          Expanded(
            child: OutlinedButton(
              key: ConnectionCardButtons.disconnectButtonKey,
              focusNode: _buttonFocusNode,
              style: buttonTheme.cancelButtonStyle,
              onPressed: () async => await ref
                  .read(vpnStatusControllerProvider.notifier)
                  .disconnect(),
              child: Semantics(
                label: "${_buildSemanticsText()} ${t.ui.disconnect}",
                button: true,
                enabled: true,
                excludeSemantics: true,
                child: Text(t.ui.disconnect),
              ),
            ),
          ),
          _buildConnectionDetailsButton(context),
        ];
      }
      return [
        Expanded(
          child: FocusTraversalGroup(
            child: ContextMenu(
              key: ConnectionCardButtons.pauseConnectionButtonKey,
              matchAnchorWidth: true,
              items: [
                ...ConnectionCardButtons._pauseLengths.map(
                  (pause) => ContextMenuItem(
                    label: _pauseLabel(pause),
                    onTap: () async => await _pauseConnection(pause),
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
                focusNode: _buttonFocusNode,
                style: buttonTheme.pauseConnectionButtonStyle,
                onPressed: toggleMenu,
                child: Semantics(
                  label: "${_buildSemanticsText()} ${t.ui.pauseConnection}",
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
          extraItems: [
            ContextMenuItem(
              label: t.ui.reconnect,
              onTap: () async => await _reconnect(settings),
            ),
          ],
        ),
      ];
    }

    if (widget.vpnStatus.isConnecting()) {
      return [_buildConnectingStateButton(context)];
    }

    return [
      _buildDisconnectedStateButton(
        context,
        settings,
        recommendedServerLocation,
      ),
    ];
  }

  Widget _buildDisconnectedStateButton(
    BuildContext context,
    ApplicationSettings? settings,
    RecommendedServerLocation? recommendedServerLocation,
  ) {
    final buttonTheme = context.connectionCardTheme.buttonTheme;
    return Expanded(
      child: OutlinedButton(
        key: ConnectionCardButtons.secureMyConnectionButtonKey,
        focusNode: _buttonFocusNode,
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
          label:
              "${_buildSemanticsText(recommendedServerLocation: recommendedServerLocation)} ${t.ui.secureMyConnection}",
          enabled: true,
          button: true,
          excludeSemantics: true,
          child: Text(t.ui.secureMyConnection),
        ),
      ),
    );
  }

  Widget _buildConnectingStateButton(BuildContext context) {
    final buttonTheme = context.connectionCardTheme.buttonTheme;
    return Expanded(
      child: OutlinedButton(
        key: ConnectionCardButtons.cancelButtonKey,
        focusNode: _buttonFocusNode,
        onPressed: () async {
          await ref.read(vpnStatusControllerProvider.notifier).cancelConnect();
        },
        style: buttonTheme.cancelButtonStyle,
        child: Semantics(
          label: "${_buildSemanticsText()} ${t.ui.cancel}",
          enabled: true,
          button: true,
          excludeSemantics: true,
          child: Text(t.ui.cancel),
        ),
      ),
    );
  }

  Future<void> _reconnect(ApplicationSettings? settings) async {
    if (settings?.obfuscatedServers == true) {
      widget.vpnStatus.connectionParameters.group = ServerType.obfuscated
          .toServerGroup();
    }
    await ref
        .read(vpnStatusControllerProvider.notifier)
        .reconnect(widget.vpnStatus.connectionParameters);
  }

  Future<void> _pauseConnection(PauseLength pauseLength) async {
    ref.read(vpnStatusControllerProvider.notifier).pauseConnection(pauseLength);
  }

  void _changeSettings(BuildContext context) {
    context.navigateToRoute(AppRoute.settingsVpnConnection);
    ref.read(uiEventRepositoryProvider).reportChangeSettings();
  }

  void _getHelp() {
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
    BuildContext context, {
    List<ContextMenuItem> extraItems = const [],
  }) {
    final buttonTheme = context.connectionCardTheme.buttonTheme;
    return IntrinsicWidth(
      child: FocusTraversalGroup(
        child: ContextMenu(
          items: [
            ...extraItems,
            ContextMenuItem(
              label: t.ui.changeVPNsettings,
              onTap: () => _changeSettings(context),
            ),
            ContextMenuItem(label: t.ui.getHelp, onTap: () => _getHelp()),
          ],
          anchorBuilder: (toggleMenu) => OutlinedButton(
            style: buttonTheme.connectionDetailsButtonStyle,
            onPressed: toggleMenu,
            child: Semantics(
              label: t.a11y.moreOptions,
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

  String _buildSemanticsText({
    RecommendedServerLocation? recommendedServerLocation,
  }) {
    // VPN Panel. Preferred location Fastest Server/[City], [Country]. Not secured.
    // VPN Panel. Connecting to Fastest Server/[City], [Country].
    // VPN Panel. Connected to [City], [Country].
    if (widget.vpnStatus.isDisconnected() || widget.vpnStatus.isPaused()) {
      return t.a11y.VPNPanelDisconnected(
        location: _buildRecommendedCityAndCountryText(
          recommendedServerLocation,
        ),
      );
    }

    if (widget.vpnStatus.isConnecting()) {
      return t.a11y.VPNPanelConnecting(location: _buildCityAndCountryText());
    }

    if (widget.vpnStatus.isConnected()) {
      return t.a11y.VPNPanelConnected(location: _buildCityAndCountryText());
    }

    return t.a11y.VPNPanelLoading;
  }

  String _buildCityAndCountryText() {
    if (widget.vpnStatus.isMeshnetRouting) {
      return widget.vpnStatus.hostname ?? widget.vpnStatus.ip ?? "";
    }

    if (widget.vpnStatus.country == null) return t.ui.fastestServer;

    final city = widget.vpnStatus.city != null
        ? "${widget.vpnStatus.city!}, "
        : "";
    final virtual = widget.vpnStatus.isVirtualLocation
        ? " ${t.ui.virtual}"
        : "";
    return "$city${widget.vpnStatus.country!.localizedName}$virtual";
  }

  String _buildRecommendedCityAndCountryText(
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
