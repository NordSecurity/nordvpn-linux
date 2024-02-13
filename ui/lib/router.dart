import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_web_plugins/url_strategy.dart';
import 'package:go_router/go_router.dart';
import 'package:nordvpn/app_scaffold.dart';
import 'package:nordvpn/route_info.dart';

class AppRouter {
  AppRouter._();

  static final _rootNavigatorKey = GlobalKey<NavigatorState>();

  static GoRouter constructRouter() {
    if (kIsWeb) {
      usePathUrlStrategy();
      GoRouter.optionURLReflectsImperativeAPIs = true;
    }

    final routes = [
      AppRouteInfo(
          label: "VPN",
          path: "/vpn",
          icon: const Icon(Icons.lock_outline),
          selectedIcon: const Icon(Icons.lock),
          builder: (context, state) => const Text("vpn")),
      AppRouteInfo(
        label: "Help",
        path: "/help",
        icon: const Icon(Icons.help_outline),
        selectedIcon: const Icon(Icons.help),
        builder: (context, state) => const Text("help"),
      ),
      AppRouteInfo(
        label: "Settings",
        path: "/settings",
        icon: const Icon(Icons.settings_outlined),
        selectedIcon: const Icon(Icons.settings),
        builder: (context, state) => const Text("settings"),
      ),
    ];

    var isLoggedIn = !kIsWeb;

    return GoRouter(
      navigatorKey: _rootNavigatorKey,
      initialLocation: "/login",
      redirect: (context, state) {
        if (!isLoggedIn) {
          return "/login";
        }

        if (state.matchedLocation == '/login') {
          return routes.first.path;
        }

        return null;
      },
      routes: [
        for (AppRouteInfo r in routes)
          GoRoute(
            path: r.path,
            pageBuilder: (context, state) {
              var child = r.builder(context, state);
              return NoTransitionPage(
                  child: AppScaffold(
                body: child,
                routes: routes,
              ));
            },
          ),
        // login must not show the navigation bar
        GoRoute(
          path: "/login",
          builder: (context, state) {
            return TextButton(
                child: const Text('Looks like a FlatButton'),
                onPressed: () {
                  isLoggedIn = true;
                  GoRouter.of(context).go(routes.first.path);
                });
          },
        ),
      ],
    );
  }
}
