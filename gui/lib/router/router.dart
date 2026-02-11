import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_web_plugins/url_strategy.dart';
import 'package:go_router/go_router.dart';
import 'package:nordvpn/data/providers/account_controller.dart';
import 'package:nordvpn/data/providers/consent_status_provider.dart';
import 'package:nordvpn/data/providers/grpc_connection_controller.dart';
import 'package:nordvpn/data/providers/login_status_provider.dart';
import 'package:nordvpn/data/providers/snap_permissions_provider.dart';
import 'package:nordvpn/router/routes.dart';

final goRouterKey = GlobalKey<NavigatorState>();

// Configures router.
final routerProvider = Provider<GoRouter>((ref) {
  if (kIsWeb) {
    usePathUrlStrategy();
    GoRouter.optionURLReflectsImperativeAPIs = true;
  }

  // Listens for account and connection changes, updates redirect and notifies
  // about the changes.
  final redirectStateProvider =
      NotifierProvider<RedirectNotifier, RedirectState>(RedirectNotifier.new);
  final redirect = ref.watch(redirectStateProvider);
  final router = GoRouter(
    navigatorKey: goRouterKey,
    debugLogDiagnostics: kDebugMode,
    initialLocation: AppRoute.loadingScreen.toString(),
    redirect: (_, state) => redirect.route(state.uri)?.toString(),
    routes: configureRoutes(),
  );

  ref.onDispose(router.dispose);

  return router;
});

// Keep state of connection, login and error and gives route for redirect
// based on that information.
class RedirectState {
  final bool isLoading;
  final bool hasError;
  final bool isLoggedIn;
  final bool displayConsent;
  final bool missingSnapPermissions;

  const RedirectState({
    required this.isLoading,
    required this.hasError,
    required this.isLoggedIn,
    required this.displayConsent,
    required this.missingSnapPermissions,
  });

  factory RedirectState.initial() => const RedirectState(
    isLoading: false,
    hasError: false,
    isLoggedIn: false,
    displayConsent: false,
    missingSnapPermissions: false,
  );

  RedirectState copyWith({
    bool? isLoading,
    bool? hasError,
    bool? isLoggedIn,
    bool? displayConsent,
    bool? missingSnapPermissions,
  }) {
    return RedirectState(
      isLoading: isLoading ?? this.isLoading,
      hasError: hasError ?? this.hasError,
      isLoggedIn: isLoggedIn ?? this.isLoggedIn,
      displayConsent: displayConsent ?? this.displayConsent,
      missingSnapPermissions:
          missingSnapPermissions ?? this.missingSnapPermissions,
    );
  }

  // Calculates the route based on the connection, account and error state:
  //
  // checking connection             => loading screen
  // no connection                   => error screen
  // is connected but not logged in  => login screen
  // is connected and just logged in => main screen
  //
  // Otherwise don't do redirect.
  AppRoute? route(Uri uri) {
    // Global redirects
    if (missingSnapPermissions) return AppRoute.missingSnapPermissions;
    if (isLoading) return AppRoute.loadingScreen;
    if (hasError) return AppRoute.errorScreen;
    if (displayConsent) return AppRoute.consentScreen;
    if (!isLoggedIn) return AppRoute.login;

    if (_shouldRedirectToVpnScreen(uri)) {
      return AppRoute.vpn;
    }
    return null;
  }

  bool _shouldRedirectToVpnScreen(Uri uri) {
    final redirectToVpnRoutes = [
      AppRoute.login.toString(),
      AppRoute.errorScreen.toString(),
      AppRoute.loadingScreen.toString(),
      AppRoute.consentScreen.toString(),
      AppRoute.missingSnapPermissions.toString(),
    ];
    final path = uri.toString();
    return redirectToVpnRoutes.contains(path);
  }
}

class RedirectNotifier extends Notifier<RedirectState> {
  @override
  RedirectState build() {
    final connection = ref.watch(grpcConnectionControllerProvider);
    final consent = ref.watch(consentStatusProvider);
    final login = ref.watch(loginStatusProvider);
    final account = ref.watch(accountControllerProvider);
    final snap = ref.watch(snapPermissionsProvider);

    final isLoading = [
      connection,
      consent,
      login,
      snap,
    ].any((v) => v.isLoading);

    final hasError = [
      connection,
      consent,
      login,
      account,
    ].any((v) => v.hasError);

    final isLoggedIn = login.maybeWhen(
      data: (v) => v == true,
      orElse: () => false,
    );

    final displayConsent = consent.maybeWhen(
      data: (v) => v == ConsentLevel.none,
      orElse: () => false,
    );

    final missingSnapPermissions = snap.maybeWhen(
      data: (v) => v.isNotEmpty,
      orElse: () => false,
    );

    return RedirectState(
      isLoading: isLoading,
      hasError: hasError,
      isLoggedIn: isLoggedIn,
      displayConsent: displayConsent,
      missingSnapPermissions: missingSnapPermissions,
    );
  }
}
