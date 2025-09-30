import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nordvpn/analytics/consent_screen.dart';
import 'package:nordvpn/app_scaffold.dart';
import 'package:nordvpn/daemon/error_screen.dart';
import 'package:nordvpn/daemon/loading_screen.dart';
import 'package:nordvpn/daemon/login_screen.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/router/metadata.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/settings/account_details_screen.dart';
import 'package:nordvpn/settings/allow_list/allow_list_settings.dart';
import 'package:nordvpn/settings/autoconnect_settings.dart';
import 'package:nordvpn/settings/custom_dns.dart';
import 'package:nordvpn/settings/general_settings.dart';
import 'package:nordvpn/settings/security_and_privacy_settings.dart';
import 'package:nordvpn/settings/settings_home_screen.dart';
import 'package:nordvpn/settings/threat_protection_settings.dart';
import 'package:nordvpn/settings/vpn_connection_settings.dart';
import 'package:nordvpn/snap/snap_screen.dart';
import 'package:nordvpn/vpn/vpn.dart';
import 'package:nordvpn/widgets/responsive_scaffold.dart';
import 'package:nordvpn/widgets/widgets_showcase.dart';

final Map<String, RouteMetadata> routeToNameMap = {};

enum AppRoute {
  vpn,
  settings,
  showcase, // Debug only

  // settings sub-routes
  settingsGeneral,
  settingsVpnConnection,
  settingsAutoconnect,
  settingsSecurityAndPrivacy,
  settingsAllowList,
  settingsCustomDns,
  settingsThreatProtection,
  settingsAccount,

  // non-direct navigation routes here
  root,
  login,
  loadingScreen,
  errorScreen,
  consentScreen,
  missingSnapPermissions;

  @override
  String toString() {
    return switch (this) {
      AppRoute.vpn => "/vpn",
      AppRoute.settings => "/settings",
      AppRoute.settingsGeneral => "/settings/general",
      AppRoute.settingsVpnConnection => "/settings/vpn-connection",
      AppRoute.settingsAutoconnect => "/settings/vpn-connection/autoconnect",
      AppRoute.settingsSecurityAndPrivacy => "/settings/security-privacy",
      AppRoute.settingsAllowList => "/settings/security-privacy/allow-list",
      AppRoute.settingsCustomDns => "/settings/security-privacy/custom-dns",
      AppRoute.settingsThreatProtection => "/settings/threat-protection",
      AppRoute.settingsAccount => "/settings/account",
      AppRoute.showcase => "/showcase",
      AppRoute.root => "/root",
      AppRoute.login => "/login",
      AppRoute.loadingScreen => "/loading",
      AppRoute.errorScreen => "/error",
      AppRoute.consentScreen => "/consent",
      AppRoute.missingSnapPermissions => "/snap",
    };
  }
}

extension GoRouterExt on BuildContext {
  int currentLocationIdx() {
    String currentLocation = GoRouterState.of(this).uri.toString();
    // If there are multiple / in the path, then get the first part and
    // return the index for it because this is needed for navigation trail
    final childRouteIndex = currentLocation.indexOf("/", 1);

    if (childRouteIndex != -1) {
      currentLocation = currentLocation.substring(0, childRouteIndex);
    }

    final idx = AppRoute.values.indexWhere(
      (e) => e.toString() == currentLocation,
    );

    return idx;
  }

  String locationName(int index) => AppRoute.values[index].toString();

  void navigateToRoute(AppRoute route) {
    go(route.toString());
  }
}

List<RouteBase> configureRoutes() {
  return [
    _route(AppRoute.loadingScreen, const LoadingScreen()),
    _route(AppRoute.errorScreen, const ErrorScreen()),
    _route(AppRoute.login, const LoginScreen()),
    _route(AppRoute.consentScreen, const ConsentScreen()),
    _route(AppRoute.missingSnapPermissions, const SnapScreen()),

    _routeWithAppScaffold(
      RouteMetadata(route: AppRoute.vpn, screen: const VpnWidget()),
    ),

    _routeWithAppScaffold(
      RouteMetadata(
        route: AppRoute.settings,
        screen: const SettingsHomeScreen(),
        displayName: t.ui.settings,
        onPressed: (context) => context.navigateToRoute(AppRoute.settings),
      ),
    ),

    if (kDebugMode)
      _routeWithAppScaffold(
        RouteMetadata(
          route: AppRoute.showcase,
          screen: const WidgetsShowcase(),
        ),
      ),

    // settings pages
    _routeWithAppScaffold(
      RouteMetadata(
        route: AppRoute.settingsGeneral,
        screen: const GeneralSettings(),
        displayName: t.ui.general,
      ),
    ),
    _routeWithAppScaffold(
      RouteMetadata(
        route: AppRoute.settingsVpnConnection,
        screen: VpnConnectionSettings(),
        displayName: t.ui.vpnConnection,
        onPressed: (context) =>
            context.navigateToRoute(AppRoute.settingsVpnConnection),
      ),
    ),
    _routeWithAppScaffold(
      RouteMetadata(
        route: AppRoute.settingsAutoconnect,
        screen: AutoconnectSettings(),
        displayName: t.ui.autoConnect,
      ),
    ),
    _routeWithAppScaffold(
      RouteMetadata(
        route: AppRoute.settingsSecurityAndPrivacy,
        screen: const SecurityAndPrivacySettings(),
        displayName: t.ui.securityAndPrivacy,
      ),
    ),
    _routeWithAppScaffold(
      RouteMetadata(
        route: AppRoute.settingsAllowList,
        screen: AllowListSettings(),
        displayName: t.ui.allowlist,
      ),
    ),
    _routeWithAppScaffold(
      RouteMetadata(
        route: AppRoute.settingsCustomDns,
        screen: CustomDns(),
        displayName: t.ui.customDns,
      ),
    ),
    _routeWithAppScaffold(
      RouteMetadata(
        route: AppRoute.settingsThreatProtection,
        screen: const ThreatProtectionSettings(),
        displayName: t.ui.threatProtection,
      ),
    ),
    _routeWithAppScaffold(
      RouteMetadata(
        route: AppRoute.settingsAccount,
        screen: const AccountDetailsSettings(),
        displayName: t.ui.account,
      ),
    ),
  ];
}

// Helper function to make a blocking route.
// It is a route without scaffold and without navigation.
GoRoute _route(AppRoute route, Widget child) {
  return GoRoute(path: route.toString(), builder: (_, __) => child);
}

// Helper function to make a route for a path and a child widget.
// Routes include the application scaffold.
GoRoute _routeWithAppScaffold(RouteMetadata metadata) {
  if (metadata.displayName != null) {
    routeToNameMap.putIfAbsent(
      Uri.parse(metadata.route.toString()).pathSegments.last,
      () => metadata,
    );
  }
  return GoRoute(
    path: metadata.route.toString(),
    pageBuilder: (context, state) {
      return CustomTransitionPage(
        key: state.pageKey,
        child: ResponsiveScaffold(child: metadata.screen),
        transitionsBuilder: (context, animation, secondaryAnimation, child) {
          return FadeTransition(
            opacity: CurveTween(curve: Curves.easeInOutCirc).animate(animation),
            child: AppScaffold(child: child),
          );
        },
      );
    },
  );
}
