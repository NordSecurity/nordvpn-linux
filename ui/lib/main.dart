import 'package:flutter/material.dart';
import 'package:nordvpn/router.dart';
import 'package:nordvpn/theme.dart';

void main() {
  runApp(const NordVPNApp());
}

class NordVPNApp extends StatelessWidget {
  const NordVPNApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
        debugShowCheckedModeBanner: false,
        routerConfig: AppRouter.constructRouter(),
        title: 'NordVPN',
        theme: ThemeData(
            useMaterial3: true,
            navigationRailTheme: NavigationRailThemeData(
              backgroundColor: ThemeManager.navBarBackgroundColor,
              indicatorColor: ThemeManager.navBarSelectedItemBgColor,
              labelType: NavigationRailLabelType.all,
              useIndicator: true,
              selectedLabelTextStyle: null,
            ),
            navigationBarTheme: NavigationBarThemeData(
              backgroundColor: ThemeManager.navBarBackgroundColor,
              indicatorColor: ThemeManager.navBarSelectedItemBgColor,
            )));
  }
}
