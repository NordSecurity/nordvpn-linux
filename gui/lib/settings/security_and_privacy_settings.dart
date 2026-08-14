import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/popup_codes.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/settings/settings_wrapper_widget.dart';
import 'package:nordvpn/theme/settings_theme.dart';
import 'package:nordvpn/widgets/accessible_item.dart';
import 'package:nordvpn/widgets/advanced_list_tile.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/input.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';

final fwMarkKey = UniqueKey();
final firewallKey = UniqueKey();

final class SecurityAndPrivacySettings extends ConsumerStatefulWidget {
  const SecurityAndPrivacySettings({super.key});

  @override
  ConsumerState<SecurityAndPrivacySettings> createState() =>
      _SecurityAndPrivacySettingsState();
}

// Settings types displayed into the screen
enum _SecurityAndPrivacySettingsItems {
  allowList,
  lanDiscovery,
  customDns,
  postQuantum,
  obfuscated,
  firewall,
  firewallMark,
}

class _SecurityAndPrivacySettingsState
    extends ConsumerState<SecurityAndPrivacySettings> {
  @override
  Widget build(BuildContext context) {
    return ref
        .watch(vpnSettingsControllerProvider)
        .when(
          loading: () => const LoadingIndicator(),
          error: (error, stackTrace) => CustomErrorWidget(message: "$error"),
          data: (settings) => _build(context, settings),
        );
  }

  Widget _build(BuildContext context, ApplicationSettings settings) {
    final settingsTheme = context.settingsTheme;
    final items = <_SecurityAndPrivacySettingsItems>[
      _SecurityAndPrivacySettingsItems.allowList,
      _SecurityAndPrivacySettingsItems.customDns,
      _SecurityAndPrivacySettingsItems.lanDiscovery,
      if (settings.protocol.isOpenVpn())
        _SecurityAndPrivacySettingsItems.obfuscated,
      _SecurityAndPrivacySettingsItems.firewall,
      _SecurityAndPrivacySettingsItems.firewallMark,
      if (settings.protocol == VpnProtocol.nordlynx)
        _SecurityAndPrivacySettingsItems.postQuantum,
    ];

    return SettingsWrapperWidget(
      itemsCount: items.length,
      itemBuilder: (context, index) {
        switch (items[index]) {
          case _SecurityAndPrivacySettingsItems.allowList:
            return _buildAllowListItem(context);
          case _SecurityAndPrivacySettingsItems.lanDiscovery:
            return _buildLanDiscovery(context, settings);
          case _SecurityAndPrivacySettingsItems.customDns:
            return _buildCustomDns(context, settings);
          case _SecurityAndPrivacySettingsItems.postQuantum:
            return _accessibleSwitchTile(
              context,
              title: t.ui.postQuantumVpn,
              subtitle: t.ui.postQuantumDescription,
              value: settings.postQuantum,
              onChanged: (value) async {
                await ref
                    .read(vpnSettingsControllerProvider.notifier)
                    .setPostQuantum(value);
              },
            );
          case _SecurityAndPrivacySettingsItems.obfuscated:
            return _accessibleSwitchTile(
              context,
              title: t.ui.obfuscation,
              subtitle: t.ui.obfuscationDescription,
              value: settings.obfuscatedServers,
              onChanged: (value) async {
                await ref
                    .read(vpnSettingsControllerProvider.notifier)
                    .setObfuscated(value);
              },
            );
          case _SecurityAndPrivacySettingsItems.firewall:
            return _accessibleSwitchTile(
              context,
              key: firewallKey,
              title: t.ui.firewall,
              subtitle: t.ui.firewallDescription,
              value: settings.firewall,
              onChanged: (value) async {
                await ref
                    .read(vpnSettingsControllerProvider.notifier)
                    .setFirewall(value);
              },
              padding: settingsTheme.itemPadding.copyWith(bottom: 0),
            );
          case _SecurityAndPrivacySettingsItems.firewallMark:
            return _buildFirewallMark(context, settings);
        }
      },
    );
  }

  Widget _accessibleSwitchTile(
    BuildContext context, {
    Key? key,
    required String title,
    String? subtitle,
    required bool value,
    required Future<void> Function(bool) onChanged,
    Future<bool> Function(bool toValue)? shouldChange,
    bool enabled = true,
    EdgeInsets? padding,
  }) {
    return AccessibleItem(
      enabled: enabled,
      toggled: value,
      label: subtitle == null ? title : '$title. $subtitle',
      onActivate: () =>
          unawaited(_activateSwitch(value, onChanged, shouldChange)),
      child: SettingsWrapperWidget.buildListItem(
        context,
        title: title,
        subtitle: subtitle,
        enabled: enabled,
        padding: padding,
        trailing: OnOffSwitch(
          key: key,
          value: value,
          shouldChange: shouldChange,
          onChanged: onChanged,
        ),
      ),
    );
  }

  Future<void> _activateSwitch(
    bool value,
    Future<void> Function(bool) onChanged,
    Future<bool> Function(bool toValue)? shouldChange,
  ) async {
    final toValue = !value;
    if (shouldChange != null && !(await shouldChange(toValue))) return;
    await onChanged(toValue);
  }

  // Navigation row -> announced as a button; title + subtitle are read as part
  // of it (excludeChildSemantics: false keeps them, the arrow image is silent).
  Widget _accessibleNavTile(
    BuildContext context, {
    required String title,
    String? subtitle,
    required VoidCallback onTap,
    bool enabled = true,
  }) {
    return AccessibleItem(
      enabled: enabled,
      button: true,
      excludeChildSemantics: false,
      onActivate: onTap,
      child: SettingsWrapperWidget.buildListItem(
        context,
        title: title,
        subtitle: subtitle,
        enabled: enabled,
        trailingLocation: TrailingLocation.center,
        trailing: DynamicThemeImage("right_arrow.svg"),
        onTap: onTap,
      ),
    );
  }

  Widget _buildAllowListItem(BuildContext context) {
    return _accessibleNavTile(
      context,
      title: t.ui.allowlist,
      subtitle: t.ui.useAllowListSettingDescription,
      onTap: () => context.navigateToRoute(AppRoute.settingsAllowList),
    );
  }

  Widget _buildLanDiscovery(
    BuildContext context,
    ApplicationSettings settings,
  ) {
    return _accessibleSwitchTile(
      context,
      title: t.ui.lanDiscovery,
      subtitle: t.ui.lanDiscoveryDescription,
      value: settings.lanDiscovery,
      shouldChange: (toValue) => _canChange(settings, toValue),
      onChanged: (value) => _toggleLanDiscovery(value),
    );
  }

  Future<bool> _canChange(ApplicationSettings settings, bool toValue) async {
    // when user tries to enable it and Allowlist contains private subnets, show
    // popup with warning and don't allow to switch to on here (it will be done
    // in popup)
    if (toValue && settings.allowListData.hasPrivateSubnets) {
      ref
          .read(popupsProvider.notifier)
          .show(PopupCodes.removePrivateSubnetsFromAllowlist);
      return false;
    }

    // allow to switch only when Allowlist does not contain any subnets
    return true;
  }

  Future<void> _toggleLanDiscovery(bool value) async {
    ref.read(vpnSettingsControllerProvider.notifier).setLanDiscovery(value);
  }

  Widget _buildCustomDns(BuildContext context, ApplicationSettings settings) {
    return _accessibleNavTile(
      context,
      title: t.ui.customDns,
      subtitle: t.ui.customDnsDescription,
      onTap: () => context.navigateToRoute(AppRoute.settingsCustomDns),
    );
  }

  Widget _buildFirewallMark(
    BuildContext context,
    ApplicationSettings settings,
  ) {
    final settingsTheme = context.settingsTheme;

    return SettingsWrapperWidget.buildListItem(
      context,
      enabled: settings.firewall,
      title: t.ui.firewallMark,
      titleStyle: settingsTheme.itemSubtitleStyle,
      trailingLocation: TrailingLocation.top,
      padding: settingsTheme.itemPadding.copyWith(top: 0),
      trailing: SizedBox(
        width: settingsTheme.fwMarkInputSize,
        child: Input(
          key: fwMarkKey,
          submitDisplay: SubmitDisplay.always,
          submitText: t.ui.save,
          text: "0x${settings.firewallMark.toRadixString(16)}",
          onSubmitted: (value) async {
            if (!_isFirewallMarkValid(value, settings.firewallMark)) {
              return;
            }
            final fwMark = int.tryParse(value.substring(2), radix: 16);
            if ((fwMark != null) && (fwMark <= maxInt32)) {
              await ref
                  .read(vpnSettingsControllerProvider.notifier)
                  .setFirewallMark(fwMark);
            }
          },
          errorMessage: t.ui.invalidFormat,
          validateInput: (value) =>
              _isFirewallMarkValid(value, settings.firewallMark),
        ),
      ),
    );
  }

  bool _isFirewallMarkValid(String value, int currentFirewallMark) {
    if (value.isEmpty) {
      return true;
    }
    if (!RegExp(r'^0x[0-9a-fA-F]+$').hasMatch(value)) {
      return false;
    }
    final fwMark = int.tryParse(value.substring(2), radix: 16);
    if ((fwMark == null) || (fwMark > maxInt32)) {
      return false;
    }

    return fwMark != currentFirewallMark;
  }
}
