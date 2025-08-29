import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_svg/svg.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/config.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/mocks/daemon/grpc_server.dart';
import 'package:nordvpn/main.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/theme.dart';

import 'app_ctl.dart';

extension Helper on WidgetTester {
  // duration is how much time to give for the app to run
  // timeout is how many times to run pump with duration
  // It stops when there are no more animations after <duration>
  Future<int> pumpAndSettleWithTimeout({
    Duration duration = const Duration(milliseconds: 100),
    Duration timeout = const Duration(seconds: 5),
  }) async {
    try {
      return await pumpAndSettle(
        duration,
        EnginePhase.sendSemanticsUpdate,
        timeout,
      );
    } catch (error) {
      return 1;
    }
  }

  Future<void> pumpUntilFound(
    Finder finder, {
    Duration timeout = const Duration(seconds: 5),
    Duration interval = const Duration(milliseconds: 100),
    Matcher matcher = findsOneWidget,
  }) async {
    final endTime = DateTime.now().add(timeout);
    while (DateTime.now().isBefore(endTime)) {
      await pump(interval);
      if (finder.evaluate().isNotEmpty) {
        expect(finder, matcher);
        await pumpAndSettleWithTimeout(timeout: Duration(milliseconds: 100));
        return;
      }
    }
    throw Exception('$finder within the timeout period.');
  }

  Future<void> pumpUntilTrue(
    bool Function() condition, {
    Duration timeout = const Duration(seconds: 5),
    Duration interval = const Duration(milliseconds: 100),
  }) async {
    final endTime = DateTime.now().add(timeout);
    while (DateTime.now().isBefore(endTime)) {
      await pump(interval);
      if (condition()) {
        return;
      }
    }
    throw Exception('timeout waiting for condition');
  }

  // Used in integration tests.
  // It will start the mocked GrpcServer, configure sl and create a NordVpnApp widget
  Future<AppCtl> setupIntegrationTests({
    Widget child = const NordVpnApp(),
    Config? config,
    bool skipLoadingScreen = true,
    Size? windowSize,
    Settings? appSettings,
  }) async {
    if (config != null) {
      if (sl.isRegistered<Config>()) {
        sl.unregister<Config>();
      }
      sl.registerSingleton<Config>(config);
    }

    await resizeMainWindow();

    // ensure mock gRPC server is used
    expect(useMockDaemon, true);
    await GrpcServer.instance.start();
    await binding.setSurfaceSize(windowSize);
    GrpcServer.instance.account.delayDuration = skipLoadingScreen
        ? Duration.zero
        : Duration(seconds: 3);

    if (appSettings != null) {
      GrpcServer.instance.appSettings.replaceSettings(appSettings);
    }

    await pumpWidget(
      ProviderScope(
        child: Builder(
          builder: (context) {
            final MediaQueryData data = MediaQuery.of(context);
            return MediaQuery(
              data: data.copyWith(textScaler: TextScaler.linear(1)),
              child: child,
            );
          },
        ),
      ),
    );

    await pumpAndSettleWithTimeout();
    addTearDown(() async {
      // unmount ProviderScope and
      // wait for async tasks to finish
      await pumpWidget(Container());
      await pumpAndSettle();

      await GrpcServer().stop();
      await binding.setSurfaceSize(null);
    });

    return AppCtl(tester: this);
  }

  // Used in widget testing to create the MaterialApp and the theme
  // In this case the gRPC server is not started.
  Future<void> setupWidgetTest(Widget child) async {
    await pumpWidget(
      ProviderScope(
        overrides: [],
        child: Builder(
          builder: (context) {
            final MediaQueryData data = MediaQuery.of(context);
            return MediaQuery(
              data: data.copyWith(textScaler: TextScaler.linear(1)),
              child: MaterialApp(
                debugShowCheckedModeBanner: true,
                title: "NordVPN",
                theme: lightTheme(),
                darkTheme: darkTheme(),
                home: Scaffold(body: child),
              ),
            );
          },
        ),
      ),
    );

    await pumpAndSettleWithTimeout();
  }

  Finder findSvgWithPath(String path) {
    return find.byWidgetPredicate(
      (widget) =>
          widget is SvgPicture &&
          widget.bytesLoader is SvgAssetLoader &&
          (widget.bytesLoader as SvgAssetLoader).assetName ==
              "assets/images/$path",
    );
  }
}
