import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/internal/assets_manager.dart';

void main() {
  test('All assets from rootBundle are found', () async {
    WidgetsFlutterBinding.ensureInitialized();
    final assetManager = await AssetManager.create();

    final assets = await AssetManifest.loadFromAssetBundle(rootBundle);
    for (var asset in assets.listAssets()) {
      final actual = assetManager.exists(asset);
      expect(actual, true);
    }

    final actual = assetManager.exists("abc/tst");
    expect(actual, false);
  });
}
