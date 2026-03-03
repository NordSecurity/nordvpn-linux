import 'dart:io';

final class SnapHelpers {
  SnapHelpers._();

  static bool isSnapContext() {
    return Platform.environment.containsKey('SNAP_NAME');
  }
}
