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

    return MergeSemantics(
      child: Semantics(
        button: true,
        enabled: enabled,
        onTap: enabled ? open : null,
        child: FocusableActionDetector(
          enabled: enabled,
          mouseCursor: SystemMouseCursors.click,
          shortcuts: const <ShortcutActivator, Intent>{
            SingleActivator(LogicalKeyboardKey.enter): ActivateIntent(),
            SingleActivator(LogicalKeyboardKey.space): ActivateIntent(),
          },
          actions: <Type, Action<Intent>>{
            ActivateIntent: CallbackAction<ActivateIntent>(
              onInvoke: (_) {
                open();
                return null;
              },
            ),
          },
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
        ),
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
    void toggle() {
      if (enabled) unawaited(onChanged(!value));
    }

    return MergeSemantics(
      child: Semantics(
        toggled: value,
        enabled: enabled,
        label: subtitle == null ? title : '$title. $subtitle',
        onTap: enabled ? toggle : null,
        child: FocusableActionDetector(
          enabled: enabled,
          mouseCursor: SystemMouseCursors.click,
          shortcuts: const <ShortcutActivator, Intent>{
            SingleActivator(LogicalKeyboardKey.enter): ActivateIntent(),
            SingleActivator(LogicalKeyboardKey.space): ActivateIntent(),
          },
          actions: <Type, Action<Intent>>{
            ActivateIntent: CallbackAction<ActivateIntent>(
              onInvoke: (_) {
                toggle();
                return null;
              },
            ),
          },
          child: ExcludeSemantics(
            child: SettingsWrapperWidget.buildListItem(
              context,
              title: title,
              subtitle: subtitle,
              trailing: OnOffSwitch(
                key: key,
                value: value,
                onChanged: onChanged,
              ),
            ),
          ),
        ),
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
    final semanticLabel = groupLabel == null ? label : '$groupLabel, $label';

    void select() {
      if (enabled) onChanged(value);
    }

    return MergeSemantics(
      child: Semantics(
        inMutuallyExclusiveGroup: true,
        checked: selected,
        enabled: enabled,
        label: semanticLabel,
        onTap: enabled ? select : null,
        child: FocusableActionDetector(
          enabled: enabled,
          mouseCursor: SystemMouseCursors.click,
          shortcuts: const <ShortcutActivator, Intent>{
            SingleActivator(LogicalKeyboardKey.enter): ActivateIntent(),
            SingleActivator(LogicalKeyboardKey.space): ActivateIntent(),
          },
          actions: <Type, Action<Intent>>{
            ActivateIntent: CallbackAction<ActivateIntent>(
              onInvoke: (_) {
                select();
                return null;
              },
            ),
          },
          child: ExcludeSemantics(
            child: RadioButton(
              key: VpnSettingsWidgetKeys.protocolRadio(value),
              value: value,
              groupValue: groupValue,
              onChanged: (VpnProtocol? v) {
                if (v != null) onChanged(v);
              },
              label: label,
            ),
          ),
        ),
      ),
    );
  }

  void _setProtocol(WidgetRef ref, VpnProtocol value) async {
    await ref
        .read(vpnSettingsControllerProvider.notifier)
        .setVpnProtocol(value);
  }
}
