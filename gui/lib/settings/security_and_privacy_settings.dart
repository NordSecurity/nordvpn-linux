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
            return SettingsWrapperWidget.buildListItem(
              context,
              title: t.ui.postQuantumVpn,
              subtitle: t.ui.postQuantumDescription,
              trailing: OnOffSwitch(
                value: settings.postQuantum,
                onChanged: (value) async {
                  await ref
                      .read(vpnSettingsControllerProvider.notifier)
                      .setPostQuantum(value);
                },
              ),
            );
          case _SecurityAndPrivacySettingsItems.obfuscated:
            return SettingsWrapperWidget.buildListItem(
              context,
              title: t.ui.obfuscation,
              subtitle: t.ui.obfuscationDescription,
              trailing: OnOffSwitch(
                value: settings.obfuscatedServers,
                onChanged: (value) async {
                  await ref
                      .read(vpnSettingsControllerProvider.notifier)
                      .setObfuscated(value);
                },
              ),
            );
          case _SecurityAndPrivacySettingsItems.firewall:
            return SettingsWrapperWidget.buildListItem(
              context,
              title: t.ui.firewall,
              subtitle: t.ui.firewallDescription,
              trailing: OnOffSwitch(
                key: firewallKey,
                value: settings.firewall,
                onChanged: (value) async {
                  await ref
                      .read(vpnSettingsControllerProvider.notifier)
                      .setFirewall(value);
                },
              ),
              padding: settingsTheme.itemPadding.copyWith(bottom: 0),
            );
          case _SecurityAndPrivacySettingsItems.firewallMark:
            return _buildFirewallMark(context, settings);
        }
      },
    );
  }

  Widget _buildAllowListItem(BuildContext context) {
    return SettingsWrapperWidget.buildListItem(
      context,
      title: t.ui.allowlist,
      subtitle: t.ui.allowListDescription,
      trailingLocation: TrailingLocation.center,
      trailing: DynamicThemeImage("right_arrow.svg"),
      onTap: () => context.navigateToRoute(AppRoute.settingsAllowList),
    );
  }

  Widget _buildLanDiscovery(
    BuildContext context,
    ApplicationSettings settings,
  ) {
    return SettingsWrapperWidget.buildListItem(
      context,
      title: t.ui.lanDiscovery,
      subtitle: t.ui.lanDiscoveryDescription,
      trailing: OnOffSwitch(
        value: settings.lanDiscovery,
        shouldChange: (toValue) => _canChange(settings, toValue),
        onChanged: (value) => _toggleLanDiscovery(value),
      ),
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
    return SettingsWrapperWidget.buildListItem(
      context,
      title: t.ui.customDns,
      subtitle: t.ui.customDnsDescription,
      trailingLocation: TrailingLocation.center,
      trailing: DynamicThemeImage("right_arrow.svg"),
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
