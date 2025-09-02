import 'dart:async';

import 'package:grpc/grpc.dart';
import 'package:nordvpn/data/models/application_error.dart';
import 'package:nordvpn/grpc/grpc_service.dart';
import 'package:nordvpn/grpc/error_handling_interceptor.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/common.pb.dart';
import 'package:nordvpn/pb/daemon/service.pbgrpc.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'grpc_connection_controller.g.dart';

// How often there should be a gRPC call made to the daemon to check its state
// It is also used to set timeout for the compatibility gRPC call
const _pingCallTimeout = Duration(seconds: 5);

// Monitors the gRPC connection status and notifies for changes
@riverpod
class GrpcConnectionController extends _$GrpcConnectionController {
  final DaemonClient _client = createDaemonClient();
  late Timer _pingTimer;
  late StreamSubscription _errorInterceptorSubscription;

  @override
  FutureOr<bool> build() async {
    sl<ClientChannel>().onConnectionStateChanged.listen(
      (event) => _onConnectionStateChanged(event),
      onError: (err) => _onConnectionError(err),
      cancelOnError: false,
    );

    // add listener for all the gRPC errors in the application
    _errorInterceptorSubscription = sl<ErrorHandlingInterceptor>()
        .grpcStreamErrors
        .listen((error) => _onGrpcInterceptorError(error));

    _pingTimer = Timer.periodic(_pingCallTimeout, (timer) async {
      await _pingDaemon();
    });

    _pingDaemon();
    ref.onDispose(_dispose);

    return true;
  }

  // Update state only when the new state is different,
  // to prevent too many widget rebuilds
  void _updateState(AsyncValue<bool> newState) {
    if (state == newState) {
      return;
    }

    switch (newState) {
      case AsyncData(:final value):
        if (state is AsyncData<bool> && state.value == value) {
          return;
        }
        break;

      case AsyncLoading():
        if (state is AsyncLoading) {
          return;
        }
        break;

      case AsyncError(:final error):
        assert(
          error is ApplicationError,
          "$error must be of type ApplicationError",
        );
        if (state is AsyncError<bool> &&
            error is ApplicationError &&
            state.error is ApplicationError &&
            error.code == (state.error as ApplicationError).code) {
          return;
        }
        break;
    }

    logger.i("state changed to $newState");
    state = newState;
  }

  void _dispose() {
    _pingTimer.cancel();
    _errorInterceptorSubscription.cancel();
  }

  void _checkApiCompatibility(int apiVersion) {
    if (apiVersion != DaemonApiVersion.CURRENT_VERSION.value) {
      logger.e(
        "API version error, GUI API: ${DaemonApiVersion.CURRENT_VERSION.value}, Daemon API: $apiVersion",
      );
      _updateState(
        AsyncError(
          ApplicationError(AppStatusCode.compatibilityIssue),
          StackTrace.current,
        ),
      );
      return;
    }

    // notify dependant providers only when there is actual change in the state
    if (state is AsyncData<bool> && state.value == true) return;

    state = const AsyncData(true);
  }

  Future<void> _pingDaemon() async {
    try {
      // Timeout is used for when the user was part of the nordvpn group,
      // but later was removed without rebooting the system. Because of this
      // there is a mismatch in the system: user is still part of the group
      //until reboot, but into the groups file it is not part of the group
      // anymore. In this case the daemon will reject the connection, after
      // socket is connected. This case is not handled by dart implementation
      // and the only way to detect this, is based on the deadline timeout
      // (but only for this call).
      final response = await _client.getDaemonApiVersion(
        GetDaemonApiVersionRequest(),
        options: CallOptions(timeout: _pingCallTimeout),
      );
      _checkApiCompatibility(response.apiVersion);
    } catch (error) {
      _onConnectionError(error);
    }
  }

  // This function is only for debugging because of the issue in the gRPC
  // implementation: https://github.com/grpc/grpc-dart/issues/618
  // When the daemon is not started or the user is not part of nordvpn users
  // group then this function is called several times per second with
  // connecting, idle, ready. Because of this it cannot be used to handle the
  // connection lifetime.
  // The state will be updated by the timer and by the interceptors errors
  void _onConnectionStateChanged(ConnectionState connectionState) {
    logger.d("grpc connection changed: $connectionState");
  }

  void _onConnectionError(Object error) {
    _updateState(
      AsyncError(
        ApplicationError(
          _toAppStatusCode(error, ignoreDeadlineError: false),
          error,
        ),
        StackTrace.current,
      ),
    );
  }

  void _onGrpcInterceptorError(ErrorGrpc error) {
    // deadline is ignored because this happens for a gRPC call and not for
    // the pinging of the daemon from here. And in that case the app must no
    // be put in error mode.
    final appStatusCode = _toAppStatusCode(error, ignoreDeadlineError: true);
    if (appStatusCode != AppStatusCode.unknown) {
      _updateState(
        AsyncError(
          ApplicationError(appStatusCode, error.error),
          StackTrace.current,
        ),
      );
      if (appStatusCode == AppStatusCode.compatibilityIssue) {
        // it means there were a gRPC breaking changes for a call, in this
        // case stop the timer, because otherwise the error would disappear
        // the user will need to reopen the application
        _pingTimer.cancel();
      }
    }
  }

  AppStatusCode _toAppStatusCode(
    Object error, {
    required bool ignoreDeadlineError,
  }) {
    if (error is GrpcError) {
      switch (error.code) {
        case StatusCode.invalidArgument:
        case StatusCode.unimplemented:
        case StatusCode.dataLoss:
          // compatibility error detected for a gRPC call
          return AppStatusCode.compatibilityIssue;

        case StatusCode.unavailable:
          return _mapErrorMessageToStatus(error.message);

        case StatusCode.deadlineExceeded:
          // this is important only for pinging the daemon. And it is used to
          // detect if user is not part of the nordvpn group.
          if (!ignoreDeadlineError) {
            return AppStatusCode.permissionsDenied;
          }
      }
    }

    return AppStatusCode.unknown;
  }

  AppStatusCode _mapErrorMessageToStatus(String? message) {
    final msg = message?.toLowerCase() ?? "";
    if (msg.isEmpty) return AppStatusCode.unknown;

    if (msg.contains("permission denied")) {
      // socket is not available because user is not added to nordvpn group
      return AppStatusCode.permissionsDenied;
    }

    if (msg.contains("no such file or directory")) {
      // socket is not available because the daemon is not running
      return AppStatusCode.socketNotFound;
    }

    logger.w("unknown error appeared: '$msg'");
    return AppStatusCode.unknown;
  }
}
