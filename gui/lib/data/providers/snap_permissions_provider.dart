import 'dart:io';

import 'package:nordvpn/logger.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'snap_permissions_provider.g.dart';

const requiredPermissions = [
  "network",
  "network-bind",
  "network-control",
  "firewall-control",
  "network-observe",
  "home",
  "login-session-observe",
  "system-observe",
  "hardware-observe",
];

// Observer of the snap permissions must implement this
abstract class SnapPermissionsObserver {
  void onPermissionsChanged();
}

@riverpod
final class SnapPermissions extends _$SnapPermissions {
  final Set<SnapPermissionsObserver> _snapObservers = {};

  @override
  FutureOr<List<String>> build() async {
    return await _getMissingPermissions();
  }

  Future<void> retry() async {
    state = AsyncData(await _getMissingPermissions());
    _notifySnapChanged();
  }

  void addSnapObserver(SnapPermissionsObserver observer) {
    _snapObservers.add(observer);
  }

  void removeSnapObserver(SnapPermissionsObserver observer) {
    _snapObservers.remove(observer);
  }

  void _notifySnapChanged() {
    if (_snapObservers.isEmpty) {
      return;
    }
    for (final observer in _snapObservers) {
      observer.onPermissionsChanged();
    }
  }

  static bool isSnapContext() {
    return Platform.environment.containsKey('SNAP_NAME');
  }

  Future<List<String>> _getMissingPermissions() async {
    if (!isSnapContext()) {
      return [];
    }

    try {
      final result = await Process.run("snapctl", ["is-connected", "--list"]);

      if (result.exitCode != 0) {
        logger.e(
          "failed to run \"snapctl is-connected --list\" exit code: ${result.exitCode}",
        );
        return [];
      }

      final output = (result.stdout as String).trim();
      if (output.isEmpty) {
        logger.e(
          "failed to run \"snapctl is-connected --list\" output is empty",
        );
        return [];
      }

      final currentPermissions = output.split('\n');

      return requiredPermissions
          .where((item) => !currentPermissions.contains(item))
          .toList();
    } catch (e) {
      logger.e("failed to detect missing permissions: $e");
      return [];
    }
  }
}
