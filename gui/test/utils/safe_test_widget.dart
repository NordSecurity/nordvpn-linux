import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grpc/grpc.dart';

/// Wraps a testWidgets function with runZonedGuarded safety.
// This is needed because when a gRPC call throws an exception then the test will
// fail with ArgumentError for Future result, if the gRPC call doesn't return an nullable Future type
void safeTestWidgets(
  String description,
  Future<void> Function(WidgetTester tester) testBody,
) {
  testWidgets(description, (tester) async {
    await runZonedGuarded(
      () async {
        await testBody(tester);
      },
      (error, stackTrace) {
        if (_shouldIgnoreError(error)) {
          debugPrint("Exception: $error \n $stackTrace");
        } else {
          fail('❗ runZonedGuarded caught an exception: $error\n$stackTrace');
        }
      },
    );
  });
}

// Check if the error should be ignored or not
bool _shouldIgnoreError(Object error) {
  // ignore all gRPC errors
  if (error is GrpcError) {
    return true;
  }

  // ignore ArgumentError errors for Future.catchError. This are a consequence of gRPC exceptions
  if (error is ArgumentError) {
    return (error.message as String).contains(
      "The error handler of Future.catchError must return a value of the future's type",
    );
  }

  return false;
}
