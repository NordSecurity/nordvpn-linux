import 'package:flutter/material.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/settings/settings_wrapper_widget.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/advanced_list_tile.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:package_info_plus/package_info_plus.dart';

// Main page for settings, this contains only the settings categories
final class SettingsHomeScreen extends StatelessWidget {
  const SettingsHomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return SettingsWrapperWidget(
      breadcrumbsSubtitle: _buildAppVersion(context),
      itemsCount: _SettingsGroups.values.length,
      itemBuilder: (context, index) {
        final trailing = DynamicThemeImage("right_arrow.svg");
        switch (_SettingsGroups.values[index]) {
          case _SettingsGroups.general:
            return SettingsWrapperWidget.buildListItem(
              context,
              iconName: "settings_navigation.svg",
              title: t.ui.general,
              subtitle: t.ui.generalSettingsSubtitle,
              trailing: trailing,
              trailingLocation: TrailingLocation.center,
              onTap: () => context.navigateToRoute(AppRoute.settingsGeneral),
            );
          case _SettingsGroups.vpnConnection:
            return SettingsWrapperWidget.buildListItem(
              context,
              iconName: "settings_connection.svg",
              title: t.ui.vpnConnection,
              subtitle: t.ui.vpnConnectionSubtitle,
              trailing: trailing,
              trailingLocation: TrailingLocation.center,
              onTap: () =>
                  context.navigateToRoute(AppRoute.settingsVpnConnection),
            );
          case _SettingsGroups.securityAndPrivacy:
            return SettingsWrapperWidget.buildListItem(
              context,
              iconName: "advanced_security.svg",
              title: t.ui.securityAndPrivacy,
              subtitle: t.ui.securityAndPrivacySubtitle,
              trailing: trailing,
              trailingLocation: TrailingLocation.center,
              onTap: () =>
                  context.navigateToRoute(AppRoute.settingsSecurityAndPrivacy),
            );
          case _SettingsGroups.threatProtection:
            return SettingsWrapperWidget.buildListItem(
              context,
              iconName: "threat_protection.svg",
              title: t.ui.threatProtection,
              subtitle: t.ui.threatProtectionSubtitle,
              trailing: trailing,
              trailingLocation: TrailingLocation.center,
              onTap: () =>
                  context.navigateToRoute(AppRoute.settingsThreatProtection),
            );
          case _SettingsGroups.account:
            return SettingsWrapperWidget.buildListItem(
              context,
              iconName: "account.svg",
              title: t.ui.account,
              subtitle: t.ui.accountSubtitle,
              trailing: trailing,
              trailingLocation: TrailingLocation.center,
              onTap: () => context.navigateToRoute(AppRoute.settingsAccount),
            );
        }
      },
    );
  }

  Widget _buildAppVersion(BuildContext context) {
    return FutureBuilder<PackageInfo>(
      future: PackageInfo.fromPlatform(),
      builder: (context, snapshot) {
        if (snapshot.hasError) {
          logger.e("Failed to get app version ${snapshot.error}");
        } else if (snapshot.hasData) {
          final packageInfo = snapshot.data!;
          return Align(
            alignment: Alignment.centerLeft,
            child: Text(
              "${t.ui.nordVpn} ${packageInfo.version}",
              style: context.appTheme.caption,
            ),
          );
        }
        return const SizedBox.shrink();
      },
    );
  }
}

// Contains all the settings groups
enum _SettingsGroups {
  vpnConnection,
  securityAndPrivacy,
  threatProtection,
  general,
  account,
}
