import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/mocks/daemon/grpc_server.dart';
import 'package:nordvpn/data/providers/preferences_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/router/router.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/theme.dart';
import 'package:nordvpn/widgets/popups_listener.dart';
import 'package:package_info_plus/package_info_plus.dart';
import 'package:window_manager/window_manager.dart';

void main() async {
  await setupLogger();
  final packageInfo = await PackageInfo.fromPlatform();
  logger.i(
    "starting nordvpn gui ${packageInfo.version}+${packageInfo.buildNumber}",
  );

  PlatformDispatcher.instance.onError = (error, stack) {
    logger.e("$error: $stack");
    return true;
  };

  // some change
  WidgetsFlutterBinding.ensureInitialized();
  await initServiceLocator();

  await resizeMainWindow();

  runApp(const ProviderScope(child: NordVpnApp()));
}

@visibleForTesting
Future<void> resizeMainWindow() async {
  if (!kIsWeb) {
    await windowManager.ensureInitialized();

    // TODO: handle text factor scale changes
    final textScaleFactor =
        WidgetsBinding.instance.platformDispatcher.textScaleFactor;

    WindowOptions windowOptions = WindowOptions(
      size: windowDefaultSize * textScaleFactor,
      minimumSize: windowMinSize * textScaleFactor,
    );
    windowManager.waitUntilReadyToShow(windowOptions, () async {
      await windowManager.show();
      await windowManager.focus();
      // Hackish fix to set minimum window size.
      // First setMinimumSize is not working, but works calling with delay.
      // Sometimes works also without delayed future. to be sure call with delay.
      Future.delayed(Duration(microseconds: 0), () async {
        await windowManager.setMinimumSize(windowMinSize * textScaleFactor);
      });
    });
  }
}

// This is used by the SnackbarService to identify the main window.
// This must be assigned to only one widget
final scaffoldKey = GlobalKey<ScaffoldMessengerState>();

class NordVpnApp extends ConsumerStatefulWidget {
  const NordVpnApp({super.key});

  @override
  ConsumerState<NordVpnApp> createState() => _NordVpnAppState();
}

final class _NordVpnAppState extends ConsumerState<NordVpnApp> {
  @override
  void initState() {
    super.initState();
    if (useMockDaemon) {
      GrpcServer().start();
    }
  }

  @override
  Widget build(BuildContext context) {
    final userPreferences = ref.watch(preferencesControllerProvider);
    return userPreferences.maybeWhen(
      data: (preferences) => _buildApp(preferences.appearance),
      orElse: () => _buildApp(defaultTheme),
    );
  }

  MaterialApp _buildApp(ThemeMode appearance) {
    return MaterialApp.router(
      scaffoldMessengerKey: scaffoldKey,
      debugShowCheckedModeBanner: useMockDaemon,
      routerConfig: ref.watch(routerProvider),
      // wrap into a scaffold without maximum width to allow some screen to use
      // the entire windows size
      builder: (context, child) =>
          Scaffold(body: PopupsListener(child: child!)),
      title: t.ui.nordVpn,
      theme: lightTheme(),
      darkTheme: darkTheme(),
      themeMode: appearance,
    );
  }

  @override
  void dispose() {
    super.dispose();
    if (useMockDaemon) {
      GrpcServer().stop();
    }
  }
}
