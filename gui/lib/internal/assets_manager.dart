/// Contains a list of functions used to interact with the assets.
/// The functions that works on web and desktop environment without using dart:io.

library;

import 'package:flutter/services.dart';

final class AssetManager {
  Set<String>? _assetFilesSet;

  AssetManager._();

  static Future<AssetManager> create() async {
    final manager = AssetManager._();
    await manager._loadAssets();
    return manager;
  }

  Future<void> _loadAssets() async {
    if (_assetFilesSet == null) {
      final asset = await AssetManifest.loadFromAssetBundle(rootBundle);
      _assetFilesSet = asset.listAssets().toSet();
    }
  }

  // Check if an asset exists into the assets files list.
  // Path is relative starting with "assets", e.g. assets/images/image.svg
  bool exists(String path) {
    if (_assetFilesSet == null) {
      throw Exception("Assets list is not loaded. Call AssetManager.init()");
    }
    return _assetFilesSet!.contains(path);
  }
}
