import 'dart:async';

import 'package:grpc/grpc.dart';

/// Interceptor that can be enabled/disabled at runtime
/// to force all RPCs to fail with a custom error
class MockErrorInterceptor extends ServerInterceptor {
  GrpcError? _error;

  MockErrorInterceptor({GrpcError? error}) : _error = error;

  void setError(GrpcError? error) {
    _error = error;
  }

  @override
  Stream<R> intercept<Q, R>(
    ServiceCall call,
    ServiceMethod<Q, R> method,
    Stream<Q> requests,
    Stream<R> Function(
      ServiceCall call,
      ServiceMethod<Q, R> method,
      Stream<Q> requests,
    )
    invoker,
  ) {
    if (_error != null) {
      return Stream<R>.error(_error!);
    }

    return invoker(call, method, requests);
  }
}
