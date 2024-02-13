import 'package:flutter/material.dart';
import 'package:flutter_adaptive_scaffold/flutter_adaptive_scaffold.dart';
import 'package:go_router/go_router.dart';
import 'package:nordvpn/app_header.dart';
import 'package:nordvpn/route_info.dart';
import 'package:nordvpn/theme.dart';

class AppScaffold extends StatelessWidget {
  final Widget body;
  final List<AppRouteInfo> routes;

  const AppScaffold({super.key, required this.body, required this.routes});

  @override
  Widget build(BuildContext context) {
    final router = GoRouter.of(context);
    final routePath = router.routerDelegate.currentConfiguration.fullPath;
    var selectedIndex =
        routes.indexWhere((element) => element.path == routePath);
    if (selectedIndex == -1) {
      selectedIndex = 0;
    }

    return AdaptiveScaffold(
      appBar: AppHeader(),
      appBarBreakpoint: const WidthPlatformBreakpoint(begin: 0),
      smallBreakpoint: Breakpoints.small,
      mediumBreakpoint: Breakpoints.mediumAndUp,
      largeBreakpoint: const WidthPlatformBreakpoint(begin: double.infinity),
      body: (context) => body,
      useDrawer: false,
      selectedIndex: selectedIndex,
      internalAnimations: false,
      destinations: [
        for (AppRouteInfo item in routes)
          NavigationDestination(
            icon: item.icon,
            label: item.label,
            selectedIcon: item.selectedIcon,
          )
      ],
      onSelectedIndexChange: (index) {
        var route = routes[index];
        router.go(route.path);
      },
    );
  }
}
