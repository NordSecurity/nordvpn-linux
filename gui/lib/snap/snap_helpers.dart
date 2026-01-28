
import 'dart:io';

final class SnapHelpers {
    static bool isSnapContext() {
    return Platform.environment.containsKey('SNAP_NAME');
  }
}
