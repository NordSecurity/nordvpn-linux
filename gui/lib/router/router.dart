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

  final redirect = ref.read(redirectStateProvider);
  final router = GoRouter(
    navigatorKey: goRouterKey,
    debugLogDiagnostics: kDebugMode,
    initialLocation: AppRoute.loadingScreen.toString(),
    refreshListenable: redirect,
    redirect: (_, state) => redirect.route(state.uri)?.toString(),
    routes: configureRoutes(),
  );

  ref.onDispose(router.dispose);

  return router;
});

// Listens for account and connection changes, updates redirect and notifies
// about the changes.
final redirectStateProvider = ChangeNotifierProvider((ref) {
  final notifier = RedirectState();

  void updateRedirect() {
    final connection = ref.read(grpcConnectionControllerProvider);
    final consentStatus = ref.read(consentStatusProvider);
    final loginState = ref.read(loginStatusProvider);
    final account = ref.read(accountControllerProvider);
    final snap = ref.read(snapPermissionsProvider);

    notifier.update(
      isLoading:
          connection is AsyncLoading ||
          consentStatus is AsyncLoading ||
          loginState is AsyncLoading ||
          snap is AsyncLoading,
      hasError:
          connection is AsyncError ||
          consentStatus is AsyncError ||
          loginState is AsyncError ||
          account is AsyncError,
      isLoggedIn: loginState is AsyncData && loginState.value == true,
      displayConsent:
          consentStatus is AsyncData &&
          consentStatus.value == ConsentLevel.none,
      missingSnapPermissions:
          snap is AsyncData && (snap.value?.isNotEmpty ?? false),
    );
  }

  for (final provider in [
    loginStatusProvider,
    grpcConnectionControllerProvider,
    accountControllerProvider,
    consentStatusProvider,
    snapPermissionsProvider
  ]) {
    ref.listen(provider, (_, _) => updateRedirect());
  }

  updateRedirect();

  return notifier;
});

// Keep state of connection, login and error and gives route for redirect
// based on that information.
final class RedirectState extends ChangeNotifier {
  bool isLoading = false;
  bool hasError = false;
  bool isLoggedIn = false;
  bool displayConsent = false;
  bool missingSnapPermissions = false;

  void update({
    required bool isLoading,
    required bool hasError,
    required bool isLoggedIn,
    required bool displayConsent,
    required bool missingSnapPermissions,
  }) {
    final changed =
        this.isLoading != isLoading ||
        this.hasError != hasError ||
        this.isLoggedIn != isLoggedIn ||
        this.displayConsent != displayConsent ||
        this.missingSnapPermissions != missingSnapPermissions;

    if (changed) {
      this.isLoading = isLoading;
      this.hasError = hasError;
      this.isLoggedIn = isLoggedIn;
      this.displayConsent = displayConsent;
      this.missingSnapPermissions = missingSnapPermissions;
      notifyListeners();
    }
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
