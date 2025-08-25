import 'package:grpc/grpc.dart';

final class DaemonStatusCode {
  static const success = 1000;
  static const connecting = 1001;
  static const connected = 1002;
  static const disconnected = 1003;

  static const nothingToDo = 2000;
  static const vpnIsRunning = 2002;
  static const vpnNotRunning = 2003;
  static const tokenInvalidated = 2005;

  static const failure = 3000;
  static const unauthorized = 3001;
  static const configError = 3004;
  static const offline = 3007;
  static const accountExpired = 3008;
  static const noService = 3020;
  static const tokenRenewError = 3022;
  static const tokenLoginFailure = 3035;
  static const serverNotObfuscated = 3037;
  static const serverObfuscated = 3038;
  static const privateSubnetLANDiscovery = 3040;
  static const allowlistSubnetNoop = 3045;
  static const allowlistPortOutOfRange = 3046;
  static const featureHidden = 3050;
  static const technologyDisabled = 3051;

  // custom GUI defined error codes
  static const invalidTechnology = 5000;
  static const allowListModified = 5001;
  static const dnsListModified = 5002;
  static const tooManyValues = 5003;
  static const invalidDnsAddress = 5004;
  static const tpLiteDisabled = 5005;
  static const alreadyExists = 5006;
  static const restartDaemonRequiredForFwMark = 5007;
  static const grpcTimeout = 5008;
  static const grpcError = 5009; // Generic error. TODO (dfe): Handle better
  static const missingExchangeToken = 5010;
  static const alreadyLoggedIn = 5011;
  static const notLoggedIn = 5012;
  static const failedToOpenBrowserToLogin = 5013;
  static const failedToOpenBrowserToCreateAccount = 5014;
  // generic error when application is not able to connect to VPN, but it returns failure=3000
  static const failedToConnectToVpn = 5015;

  static final Map<String, int> _errorsMap = {
    "exchange token not provided": missingExchangeToken,
    "you are already logged in": alreadyLoggedIn,
    "Token parameter value is missing": tokenLoginFailure,
    "you are not logged in": notLoggedIn,
    "Please check your internet connection and try again.": offline,
  };

  static String errorMessageForCode(int code) {
    switch (code) {
      case success:
        return "";
    }

    return "Error code: $code";
  }

  static int fromGrpcError(GrpcError e) {
    switch (e.code) {
      case StatusCode.deadlineExceeded:
        return grpcTimeout;
      case StatusCode.unknown:
        return _fromErrorMessageString(e.message ?? "Unknown error");
      default:
        // This is a gRPC error
        return grpcError;
    }
  }

  static int _fromErrorMessageString(String message) {
    if (_errorsMap.containsKey(message)) {
      return _errorsMap[message] ?? failure;
    }

    for (var key in _errorsMap.keys) {
      if (message.contains(key)) {
        return _errorsMap[key] ?? failure;
      }
    }

    return failure;
  }

  DaemonStatusCode._();
}
