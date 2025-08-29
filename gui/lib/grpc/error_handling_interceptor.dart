import 'dart:async';

import 'package:grpc/grpc.dart';
import 'package:nordvpn/logger.dart';

// Class used to intercept the gRPC errors.
// This will be used to detect the incompatibility errors between GUI and daemon
final class ErrorHandlingInterceptor implements ClientInterceptor {
  final _grpcStreamErrors = StreamController<ErrorGrpc>();
  Stream<ErrorGrpc> get grpcStreamErrors => _grpcStreamErrors.stream;

  ErrorHandlingInterceptor();

  @override
  ResponseFuture<R> interceptUnary<Q, R>(
    ClientMethod<Q, R> method,
    Q request,
    CallOptions options,
    ClientUnaryInvoker<Q, R> invoker,
  ) {
    final response = invoker(method, request, options);
    // add ignore because otherwise the application will throw an exception
    // because the return type is incorrect. The response any way is
    // not used, this is just for propagating the error
    response.catchError((error) => _handleError(method.path, error)).ignore();
    return response;
  }

  @override
  ResponseStream<R> interceptStreaming<Q, R>(
    ClientMethod<Q, R> method,
    Stream<Q> requests,
    CallOptions options,
    ClientStreamingInvoker<Q, R> invoker,
  ) {
    final response = invoker(method, requests, options);
    response.handleError((error) => _handleError(method.path, error));
    return response;
  }

  void _handleError(String method, dynamic error) {
    if (!_grpcStreamErrors.hasListener) {
      logger.i("no listener for gRPC errors $error");
    } else {
      _grpcStreamErrors.add(ErrorGrpc(error, method));
    }
  }
}

// Error wrapper for gRPC to include the method name for which the error happen
final class ErrorGrpc {
  final dynamic error;
  final String call;
  ErrorGrpc(this.error, this.call);

  @override
  String toString() {
    return "$call: $error";
  }
}
