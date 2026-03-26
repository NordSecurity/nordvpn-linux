import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/theme/nav_rail_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

enum NavDestination {
  home(
    route: AppRoute.vpn,
    iconOff: "home_off.svg",
    iconOn: "home_on.svg",
    key: Key("homeNavIcon"),
  ),
  settings(
    route: AppRoute.settings,
    iconOff: "settings_navigation_off.svg",
    iconOn: "settings_navigation_on.svg",
    key: Key("settingsNavIcon"),
  ),
  showcase(
    route: AppRoute.showcase,
    iconOff: "notifications_off.svg",
    iconOn: "notifications_on.svg",
    debugOnly: true,
  );

  final AppRoute route;
  final String iconOff;
  final String iconOn;
  final Key? key;
  final bool debugOnly;

  const NavDestination({
    required this.route,
    required this.iconOff,
    required this.iconOn,
    this.key,
    this.debugOnly = false,
  });

  static List<NavDestination> get visible =>
      values.where((d) => !d.debugOnly || kDebugMode).toList(growable: false);

  static NavDestination? fromRoute(String location) {
    final topLevel = location.indexOf("/", 1);
    final prefix = topLevel != -1 ? location.substring(0, topLevel) : location;
    for (final dest in visible) {
      if (dest.route.toString() == prefix) return dest;
    }
    return null;
  }
}

// The widget will be created for each route and will contain a navigation bar,
// app bar and display the screen specific screen
final class AppScaffold extends StatelessWidget {
  final Widget child;

  const AppScaffold({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    final destinations = NavDestination.visible;
    final location = GoRouterState.of(context).uri.toString();
    final current = NavDestination.fromRoute(location);
    final selectedIndex = current != null ? destinations.indexOf(current) : -1;

    return Scaffold(
      body: Row(
        children: [
          NavRail(
            destinations: destinations,
            selectedIndex: selectedIndex,
            onDestinationSelected: (dest) {
              context.go(dest.route.toString());
            },
          ),
          Expanded(
            child: ScalerResponsiveBox(
              maxWidth: windowMaxSize.width,
              child: child,
            ),
          ),
        ],
      ),
    );
  }
}

final class NavRail extends StatelessWidget {
  final List<NavDestination> destinations;
  final int selectedIndex;
  final ValueChanged<NavDestination> onDestinationSelected;

  const NavRail({
    super.key,
    required this.destinations,
    required this.selectedIndex,
    required this.onDestinationSelected,
  });

  @override
  Widget build(BuildContext context) {
    final navTheme = context.navRailTheme;
    final textScaler = MediaQuery.textScalerOf(context);

    return Padding(
      padding: EdgeInsets.only(top: navTheme.iconsPaddingTop),
      child: Container(
        width: textScaler.scale(navTheme.railWidth),
        color: navTheme.railBg,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.center,
          spacing: textScaler.scale(navTheme.betweenIconsGap),
          children: [
            for (final (i, dest) in destinations.indexed)
              _buildNavItem(dest, i, navTheme, textScaler),
          ],
        ),
      ),
    );
  }

  Widget _buildNavItem(
    NavDestination dest,
    int index,
    NavRailTheme navTheme,
    TextScaler textScaler,
  ) {
    final item = NavItem(
      key: dest.key,
      iconOff: dest.iconOff,
      iconOn: dest.iconOn,
      selected: selectedIndex == index,
      onClick: () => onDestinationSelected(dest),
    );

    if (!dest.debugOnly) return item;

    return ClipRect(
      child: SizedBox(
        width: textScaler.scale(navTheme.containerWidth),
        height: textScaler.scale(navTheme.containerHeight),
        child: Banner(
          textStyle: const TextStyle(
            fontWeight: FontWeight.bold,
            fontSize: 8,
            color: Colors.white,
          ),
          message: "DEBUG",
          shadow: const BoxShadow(color: Colors.transparent),
          location: BannerLocation.bottomEnd,
          child: item,
        ),
      ),
    );
  }
}

final class NavItem extends StatefulWidget {
  final String iconOff;

  final String iconOn;

  final bool selected;

  final VoidCallback onClick;

  const NavItem({
    super.key,
    required this.iconOff,
    required this.iconOn,
    required this.selected,
    required this.onClick,
  });

  @override
  State<NavItem> createState() => _NavItemState();
}

final class _NavItemState extends State<NavItem> {
  bool _hovered = false;

  @override
  Widget build(BuildContext context) {
    final navTheme = context.navRailTheme;
    final textScaler = MediaQuery.textScalerOf(context);
    final active = widget.selected || _hovered;

    return GestureDetector(
      onTap: widget.onClick,
      child: MouseRegion(
        onEnter: (_) => setState(() => _hovered = true),
        onExit: (_) => setState(() => _hovered = false),
        child: Container(
          width: textScaler.scale(navTheme.containerWidth),
          height: textScaler.scale(navTheme.containerHeight),
          decoration: active
              ? BoxDecoration(
                  color: navTheme.selectedItemBg,
                  borderRadius: navTheme.radius,
                )
              : null,
          child: Center(
            child: Padding(
              padding: EdgeInsets.all(navTheme.iconsMargin),
              child: DynamicThemeImage(
                widget.selected ? widget.iconOn : widget.iconOff,
              ),
            ),
          ),
        ),
      ),
    );
  }
}
