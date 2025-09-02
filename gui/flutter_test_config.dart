// Do not rename this file. It's being looked for by flutter testing framework
// based on a convention that it is named `flutter_test_config.dart`.

import 'dart:async';

import 'package:leak_tracker_flutter_testing/leak_tracker_flutter_testing.dart';
import 'package:logger/logger.dart';
import 'package:nordvpn/logger.dart';

// Enables leaks testing. It will fail the test when the memory leak is detected.
Future<void> testExecutable(FutureOr<void> Function() testMain) async {
  logger = Logger();
  LeakTesting.enable();
  LeakTesting.settings = LeakTesting.settings
      .withIgnored(
        classes: [
          // NOTE: Not created directly by the app code.
          "CurvedAnimation",
          "ValueNotifier<Key?>",
          "_NativePicture",
          "PictureLayer",
        ],
      )
      .withCreationStackTrace();
  await testMain();
}
