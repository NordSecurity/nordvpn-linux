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
import 'package:nordvpn/settings/terms_screen.dart';
import 'package:nordvpn/settings/threat_protection_settings.dart';
import 'package:nordvpn/settings/vpn_connection_settings.dart';
import 'package:nordvpn/vpn/vpn.dart';
import 'package:nordvpn/widgets/widgets_showcase.dart';

final Map<String, RouteMetadata> routeRegistry = {};
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
  settingsTerms,

  // non-direct navigation routes here
  root,
  login,
  loadingScreen,
  errorScreen,
  consentScreen;

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
      AppRoute.settingsTerms => "/settings/terms",
      AppRoute.showcase => "/showcase",
      AppRoute.root => "/root",
      AppRoute.login => "/login",
      AppRoute.loadingScreen => "/loading",
      AppRoute.errorScreen => "/error",
      AppRoute.consentScreen => "/consent",
    };
  }
}

extension GoRouterExt on BuildContext {
  void navigateToRoute(AppRoute route) {
    go(route.toString());
  }
}

List<RouteBase> configureRoutes() {
  return [
    _route(RouteMetadata(
      route: AppRoute.loadingScreen,
      screen: const LoadingScreen(),
      isBlocking: true,
    )),
    _route(RouteMetadata(
      route: AppRoute.errorScreen,
      screen: const ErrorScreen(),
      isBlocking: true,
    )),
    _route(RouteMetadata(
      route: AppRoute.login,
      screen: const LoginScreen(),
      isBlocking: true,
    )),
    _route(RouteMetadata(
      route: AppRoute.consentScreen,
      screen: const ConsentScreen(),
      isBlocking: true,
    )),

    _routeWithAppScaffold(RouteMetadata(
      route: AppRoute.vpn,
      screen: const VpnWidget(),
      isBlocking: false,
    )),

    _routeWithAppScaffold(RouteMetadata(
      route: AppRoute.settings,
      screen: const SettingsHomeScreen(),
      isBlocking: false,
      displayName: t.ui.settings,
      onPressed: (context) => context.navigateToRoute(AppRoute.settings),
    )),

    if (kDebugMode)
      _routeWithAppScaffold(RouteMetadata(
        route: AppRoute.showcase,
        screen: const WidgetsShowcase(),
        isBlocking: false,
      )),

    // settings pages
    _routeWithAppScaffold(RouteMetadata(
      route: AppRoute.settingsGeneral,
      screen: const GeneralSettings(),
      isBlocking: false,
      displayName: t.ui.general,
    )),
    _routeWithAppScaffold(RouteMetadata(
      route: AppRoute.settingsVpnConnection,
      screen: VpnConnectionSettings(),
      isBlocking: false,
      displayName: t.ui.vpnConnection,
      onPressed: (context) =>
          context.navigateToRoute(AppRoute.settingsVpnConnection),
    )),
    _routeWithAppScaffold(RouteMetadata(
      route: AppRoute.settingsAutoconnect,
      screen: AutoconnectSettings(),
      isBlocking: false,
      displayName: t.ui.autoConnect,
    )),
    _routeWithAppScaffold(RouteMetadata(
      route: AppRoute.settingsSecurityAndPrivacy,
      screen: const SecurityAndPrivacySettings(),
      isBlocking: false,
      displayName: t.ui.securityAndPrivacy,
    )),
    _routeWithAppScaffold(RouteMetadata(
      route: AppRoute.settingsAllowList,
      screen: AllowListSettings(),
      isBlocking: false,
      displayName: t.ui.allowlist,
    )),
    _routeWithAppScaffold(RouteMetadata(
      route: AppRoute.settingsCustomDns,
      screen: CustomDns(),
      isBlocking: false,
      displayName: t.ui.customDns,
    )),
    _routeWithAppScaffold(RouteMetadata(
      route: AppRoute.settingsThreatProtection,
      screen: const ThreatProtectionSettings(),
      isBlocking: false,
      displayName: t.ui.threatProtection,
    )),
    _routeWithAppScaffold(RouteMetadata(
      route: AppRoute.settingsTerms,
      screen: const LegalInformation(),
      isBlocking: false,
      displayName: t.ui.terms,
    )),
    _routeWithAppScaffold(RouteMetadata(
      route: AppRoute.settingsAccount,
      screen: const AccountDetailsSettings(),
      isBlocking: false,
      displayName: t.ui.account,
    )),
  ];
}

// Helper function to make a route without scaffold and without navigation.
GoRoute _route(RouteMetadata metadata) {
  routeRegistry[metadata.route.toString()] = metadata;
  return GoRoute(
    path: metadata.route.toString(),
    builder: (_, _) => metadata.screen,
  );
}

// Helper function to make a route with the application scaffold.
GoRoute _routeWithAppScaffold(RouteMetadata metadata) {
  routeRegistry[metadata.route.toString()] = metadata;
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
        child: AppScaffold(child: metadata.screen),
        transitionsBuilder: (context, animation, secondaryAnimation, child) {
          return FadeTransition(
            opacity: CurveTween(curve: Curves.easeInOutCirc).animate(animation),
            child: child,
          );
        },
      );
    },
  );
}
