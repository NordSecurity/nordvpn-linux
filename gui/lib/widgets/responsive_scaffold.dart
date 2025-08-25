import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_adaptive_scaffold/flutter_adaptive_scaffold.dart';
import 'package:go_router/go_router.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/theme/breakpoints.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

final class ResponsiveScaffold extends StatelessWidget {
  final Widget child;

  const ResponsiveScaffold({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Expanded(
          child: AdaptiveScaffold(
            smallBreakpoint: AppBreakpoints.small,
            mediumBreakpoint: AppBreakpoints.medium,
            largeBreakpoint: AppBreakpoints.large,
            body: (context) => child,
            useDrawer: false,
            selectedIndex: context.currentLocationIdx(),
            internalAnimations: false,
            navigationRailWidth: 38,
            extendedNavigationRailWidth: 220,
            destinations: [
              _vpnDestination(context),
              _settingsDestination(context),
              if (kDebugMode) _widgetsShowcaseDestination(context),
            ],
            onSelectedIndexChange: (index) {
              final currentLocation = context.locationName(index);
              context.go(currentLocation);
            },
          ),
        ),
      ],
    );
  }

  NavigationDestination _vpnDestination(BuildContext context) {
    return NavigationDestination(
      icon: DynamicThemeImage("vpn_sidebar_off.svg"),
      label: context.isMediumScreen() ? "" : "VPN",
    );
  }

  NavigationDestination _settingsDestination(BuildContext context) {
    return NavigationDestination(
      icon: DynamicThemeImage("settings_navigation.svg"),
      label: context.isMediumScreen() ? "" : t.ui.settings,
    );
  }

  NavigationDestination _widgetsShowcaseDestination(BuildContext context) {
    return NavigationDestination(
      icon: Banner(
        textStyle: const TextStyle(
          fontWeight: FontWeight.bold,
          fontSize: 8,
          color: Colors.white,
        ),
        message: "DEBUG",
        shadow: BoxShadow(color: Colors.transparent),
        location: BannerLocation.bottomEnd,
        child: DynamicThemeImage("notifications.svg"),
      ),
      label: context.isMediumScreen() ? "" : "Widgets showcase",
    );
  }
}
