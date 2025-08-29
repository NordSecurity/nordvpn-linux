enum AppStatusCode {
  unknown,
  compatibilityIssue, // gRPC changed and the GUI is not able to communicate with the daemon
  socketNotFound, // the daemon socket is not found
  permissionsDenied, // the application is not able to connect to the daemon, not part of the nordvpn group
}

// Class to define application errors and to store optionally the original error
final class ApplicationError {
  final AppStatusCode code;
  final Object? originalError;
  ApplicationError(this.code, [this.originalError]);

  @override
  String toString() {
    return "ApplicationError, code: $code, error: $originalError)";
  }
}
