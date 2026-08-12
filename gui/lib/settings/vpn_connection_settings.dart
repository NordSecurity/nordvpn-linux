import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/settings/autoconnect_settings.dart';
import 'package:nordvpn/settings/settings_wrapper_widget.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/autoconnect_panel_theme.dart';
import 'package:nordvpn/widgets/advanced_list_tile.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/enabled_widget.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';
import 'package:nordvpn/widgets/radio_button.dart';

enum _VpnConnectionItems { autoConnect, killSwitch, protocol }

final class VpnSettingsWidgetKeys {
  VpnSettingsWidgetKeys._();
  static const autoConnectTile = Key("vpnSettingsAutoConnectTile");
  static const autoConnectSwitch = Key("vpnSettingsAutoConnectSwitch");
  static const killSwitch = Key("vpnSettingsKillSwitch");
  static Key protocolRadio(VpnProtocol protocol) =>
      Key("vpnSettingsProtocol_$protocol");
}

final class VpnConnectionSettings extends ConsumerWidget {
  final ImagesManager imagesManager;

  VpnConnectionSettings({super.key, ImagesManager? imagesManager})
    : imagesManager = imagesManager ?? sl();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return ref
        .watch(vpnSettingsControllerProvider)
        .when(
          loading: () => const LoadingIndicator(),
          error: (error, stackTrace) => CustomErrorWidget(message: "$error"),
          data: (settings) => _build(context, ref, settings),
        );
  }

  Widget _build(
    BuildContext context,
    WidgetRef ref,
    ApplicationSettings settings,
  ) {
    return SettingsWrapperWidget(
      itemsCount: _VpnConnectionItems.values.length,
      itemBuilder: (context, index) {
        switch (_VpnConnectionItems.values[index]) {
          case _VpnConnectionItems.autoConnect:
            return _buildAutoConnect(context, ref, settings);
          case _VpnConnectionItems.killSwitch:
            return _accessibleSwitchTile(
              context,
              key: VpnSettingsWidgetKeys.killSwitch,
              title: t.ui.killSwitch,
              subtitle: t.ui.killSwitchDescription,
              value: settings.killSwitch,
              enabled: settings.firewall,
              onChanged: (value) => ref
                  .read(vpnSettingsControllerProvider.notifier)
                  .setKillSwitch(value),
            );
          case _VpnConnectionItems.protocol:
            return _buildProtocolsList(context, ref, settings);
        }
      },
    );
  }

  Widget _buildAutoConnect(
    BuildContext context,
    WidgetRef ref,
    ApplicationSettings settings,
  ) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        _accessibleSwitchTile(
          context,
          key: VpnSettingsWidgetKeys.autoConnectSwitch,
          title: t.ui.autoConnect,
          subtitle: t.ui.autoConnectDescription,
          value: settings.autoConnect,
          onChanged: (value) => ref
              .read(vpnSettingsControllerProvider.notifier)
              .setAutoConnect(value, null),
        ),
        _buildAutoConnectTarget(context, ref, settings),
      ],
    );
  }

  Widget _buildAutoConnectTarget(
    BuildContext context,
    WidgetRef ref,
    ApplicationSettings settings,
  ) {
    final panelTheme = context.autoconnectPanelTheme;
    final enabled = settings.autoConnect;

    void open() {
      if (enabled) context.navigateToRoute(AppRoute.settingsAutoconnect);
    }

    return _AccessibleItem(
      enabled: enabled,
      button: true,
      excludeChildSemantics: false,
      onActivate: open,
      child: SettingsWrapperWidget.buildListItem(
        context,
        key: VpnSettingsWidgetKeys.autoConnectTile,
        title: "${t.ui.autoConnectTo}:",
        titleStyle: panelTheme.primaryFont,
        center: _buildCenter(context, ref, settings),
        trailingLocation: TrailingLocation.center,
        trailing: DynamicThemeImage("right_arrow.svg"),
        onTap: open,
        enabled: enabled,
      ),
    );
  }

  Widget _buildCenter(
    BuildContext context,
    WidgetRef ref,
    ApplicationSettings settings,
  ) {
    final vpnStatusProvider = ref.watch(vpnStatusControllerProvider);
    final appTheme = context.appTheme;

    return vpnStatusProvider.when(
      error: (error, _) => CustomErrorWidget(message: "$error"),
      loading: () => LoadingIndicator(),
      data: (vpnStatus) {
        return Expanded(
          child: Padding(
            padding: EdgeInsets.only(left: appTheme.outerPadding),
            child: AutoconnectSelectionStatus(
              vpnStatus: vpnStatus,
              savedLocation: settings.autoConnectLocation,
            ),
          ),
        );
      },
    );
  }

  Widget _buildProtocolsList(
    BuildContext context,
    WidgetRef ref,
    ApplicationSettings settings,
  ) {
    final appTheme = context.appTheme;
    final vpnStatus = ref.watch(vpnStatusControllerProvider);
    final isConnecting =
        vpnStatus.whenOrNull(data: (status) => status.isConnecting()) ?? false;

    List<({String label, VpnProtocol value})> protocols = [
      (label: t.ui.nordLynx, value: VpnProtocol.nordlynx),
      (label: t.ui.nordWhisper, value: VpnProtocol.nordWhisper),
      (label: t.ui.openVpnTcp, value: VpnProtocol.openVpnTcp),
      (label: t.ui.openVpnUdp, value: VpnProtocol.openVpnUdp),
    ];

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        ExcludeSemantics(
          child: SettingsWrapperWidget.buildListItem(
            context,
            title: t.ui.vpnProtocol,
          ),
        ),
        Padding(
          padding: EdgeInsets.symmetric(horizontal: appTheme.outerPadding),
          child: EnabledWidget(
            enabled: !isConnecting,
            disabledOpacity: appTheme.disabledOpacity,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                for (final item in protocols)
                  _buildAccessibleRadio(
                    context,
                    groupLabel: t.ui.vpnProtocol,
                    label: item.label,
                    value: item.value,
                    groupValue: settings.protocol,
                    enabled: !isConnecting,
                    onChanged: (value) => _setProtocol(ref, value),
                  ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _accessibleSwitchTile(
    BuildContext context, {
    Key? key,
    required String title,
    String? subtitle,
    required bool value,
    required Future<void> Function(bool) onChanged,
    bool enabled = true,
  }) {
    return _AccessibleItem(
      enabled: enabled,
      toggled: value,
      label: subtitle == null ? title : '$title. $subtitle',
      onActivate: () => unawaited(onChanged(!value)),
      child: SettingsWrapperWidget.buildListItem(
        context,
        title: title,
        subtitle: subtitle,
        trailing: OnOffSwitch(key: key, value: value, onChanged: onChanged),
      ),
    );
  }

  Widget _buildAccessibleRadio(
    BuildContext context, {
    required String label,
    required VpnProtocol value,
    required VpnProtocol groupValue,
    required ValueChanged<VpnProtocol> onChanged,
    String? groupLabel,
    bool enabled = true,
  }) {
    final selected = value == groupValue;
    final base = groupLabel == null ? label : '$groupLabel, $label';

    return _AccessibleItem(
      enabled: enabled,
      inMutuallyExclusiveGroup: true,
      checked: selected,
      label: _spokenLabel(base),
      onActivate: () => onChanged(value),
      child: RadioButton(
        key: VpnSettingsWidgetKeys.protocolRadio(value),
        value: value,
        groupValue: groupValue,
        onChanged: (VpnProtocol? v) {
          if (v != null) onChanged(v);
        },
        label: label,
      ),
    );
  }

  // Remove parenthesis from the protocol name
  String _spokenLabel(String raw) => raw
      .replaceAll(RegExp(r'[()]'), '')
      .replaceAll(RegExp(r'\s+'), ' ')
      .trim();

  void _setProtocol(WidgetRef ref, VpnProtocol value) async {
    await ref
        .read(vpnSettingsControllerProvider.notifier)
        .setVpnProtocol(value);
  }
}

class _AccessibleItem extends StatefulWidget {
  const _AccessibleItem({
    required this.child,
    required this.onActivate,
    this.label,
    this.enabled = true,
    this.excludeChildSemantics = true,
    this.button = false,
    this.toggled,
    this.checked,
    this.inMutuallyExclusiveGroup = false,
  });

  final Widget child;

  final VoidCallback onActivate;

  final String? label;
  final bool enabled;

  final bool excludeChildSemantics;

  final bool button;
  final bool? toggled;
  final bool? checked;
  final bool inMutuallyExclusiveGroup;

  @override
  State<_AccessibleItem> createState() => _AccessibleItemState();
}

class _AccessibleItemState extends State<_AccessibleItem> {
  bool _focused = false;

  @override
  Widget build(BuildContext context) {
    final onActivate = widget.enabled ? widget.onActivate : null;

    Widget visual = widget.child;
    if (widget.excludeChildSemantics) {
      visual = ExcludeSemantics(child: visual);
    }

    visual = ColoredBox(
      color: _focused ? Theme.of(context).hoverColor : Colors.transparent,
      child: visual,
    );

    return MergeSemantics(
      child: Semantics(
        button: widget.button,
        toggled: widget.toggled,
        checked: widget.checked,
        inMutuallyExclusiveGroup: widget.inMutuallyExclusiveGroup,
        enabled: widget.enabled,
        label: widget.label,
        onTap: onActivate,
        child: FocusableActionDetector(
          enabled: widget.enabled,
          mouseCursor: widget.enabled
              ? SystemMouseCursors.click
              : SystemMouseCursors.basic,
          onShowFocusHighlight: (value) {
            if (value != _focused) setState(() => _focused = value);
          },
          shortcuts: const <ShortcutActivator, Intent>{
            SingleActivator(LogicalKeyboardKey.enter): ActivateIntent(),
            SingleActivator(LogicalKeyboardKey.space): ActivateIntent(),
          },
          actions: <Type, Action<Intent>>{
            ActivateIntent: CallbackAction<ActivateIntent>(
              onInvoke: (_) {
                onActivate?.call();
                return null;
              },
            ),
          },
          child: visual,
        ),
      ),
    );
  }
}
