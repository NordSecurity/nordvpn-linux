import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/theme/nav_rail_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

final class AppScaffoldKeys {
  AppScaffoldKeys._();
  static const homeNavIcon = Key("homeNavIcon");
  static const settingsNavIcon = Key("settingsNavIcon");
}

// The widget will be created for each route and will contain a navigation bar,
// app bar and display the screen specific screen
final class AppScaffold extends StatelessWidget {
  final Widget child;

  const AppScaffold({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Row(
        children: [
          NavRail(
            selectedIndex: context.currentLocationIdx(),
            onDestinationSelected: (index) {
              context.go(context.locationName(index));
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
  final int selectedIndex;

  final ValueChanged<int> onDestinationSelected;

  const NavRail({
    super.key,
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
            NavItem(
              key: AppScaffoldKeys.homeNavIcon,
              iconOff: "home_off.svg",
              iconOn: "home_on.svg",
              selected: selectedIndex == 0,
              onClick: () => onDestinationSelected(0),
            ),
            NavItem(
              key: AppScaffoldKeys.settingsNavIcon,
              iconOff: "settings_navigation_off.svg",
              iconOn: "settings_navigation_on.svg",
              selected: selectedIndex == 1,
              onClick: () => onDestinationSelected(1),
            ),
            // showcase only in debug mode
            if (kDebugMode) ClipRect(
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
                  child: NavItem(
                    iconOff: "notifications_off.svg",
                    iconOn: "notifications_on.svg",
                    selected: selectedIndex == 2,
                    onClick: () => onDestinationSelected(2),
                  ),
                ),
              ),
            ),
          ],
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
  State<NavItem> createState() => NavItemState();
}

final class NavItemState extends State<NavItem> {
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
