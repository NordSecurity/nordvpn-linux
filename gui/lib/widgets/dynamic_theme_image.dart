import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/providers/preferences_controller.dart';
import 'package:nordvpn/internal/assets_manager.dart';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_svg/svg.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/service_locator.dart';

const _imagesPath = "assets/images/";
const _notThemedPath = "assets/images/not_themed/";

typedef _Paths = ({String regularPath, String notThemedPath});

// The widget will decide when it is created what image to display, light or dark,
// based on the current theme status
final class DynamicThemeImage extends ConsumerWidget {
  final String fileName;
  final AssetManager assetManager;

  DynamicThemeImage(this.fileName, {super.key, AssetManager? assetManager})
    : assetManager = assetManager ?? sl();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final paths = (
      regularPath: _imagesPath + fileName,
      notThemedPath: _notThemedPath + fileName,
    );
    final userPreferences = ref.watch(preferencesControllerProvider);
    return userPreferences.maybeWhen(
      data: (preferences) => _loadSvg(context, paths, preferences.appearance),
      orElse: () => _loadSvg(context, paths, defaultTheme),
    );
  }

  Widget _loadSvg(BuildContext context, _Paths paths, ThemeMode mode) {
    if (!assetManager.exists(paths.regularPath) &&
        !assetManager.exists(paths.notThemedPath)) {
      logger.f("no SVGs: '${paths.regularPath}', '${paths.notThemedPath}'");
      return _noImage(paths);
    }

    // try not-themed assets
    if (assetManager.exists(paths.notThemedPath)) {
      return SvgPicture.asset(paths.notThemedPath);
    }

    // dark/light images depending on theme
    return SvgPicture.asset(
      paths.regularPath,
      colorFilter: _isDarkTheme(context, mode)
          ? ColorFilter.mode(Colors.white, BlendMode.srcIn)
          : null,
    );
  }

  Widget _noImage(_Paths paths) {
    return kDebugMode
        ? const Icon(Icons.warning, color: Colors.red)
        : SizedBox.shrink();
  }

  bool _isDarkTheme(BuildContext context, ThemeMode mode) {
    // when theme is set to `system`
    final brightness = MediaQuery.platformBrightnessOf(context);
    return mode == ThemeMode.dark ||
        (mode == ThemeMode.system && brightness == Brightness.dark);
  }
}
