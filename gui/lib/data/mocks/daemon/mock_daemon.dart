import 'dart:async';

import 'package:fixnum/fixnum.dart';
import 'package:grpc/grpc.dart';
import 'package:nordvpn/data/mocks/daemon/mock_account_info.dart';
import 'package:nordvpn/data/mocks/daemon/mock_servers_list.dart';
import 'package:nordvpn/data/mocks/daemon/mock_application_settings.dart';
import 'package:nordvpn/data/mocks/daemon/mock_vpn_status.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/pb/daemon/account.pb.dart';
import 'package:nordvpn/pb/daemon/cities.pb.dart';
import 'package:nordvpn/pb/daemon/common.pb.dart';
import 'package:nordvpn/pb/daemon/config/analytics_consent.pb.dart';
import 'package:nordvpn/pb/daemon/connect.pb.dart';
import 'package:nordvpn/pb/daemon/defaults.pb.dart';
import 'package:nordvpn/pb/daemon/login.pb.dart' as grpc;
import 'package:nordvpn/pb/daemon/login_with_token.pb.dart';
import 'package:nordvpn/pb/daemon/logout.pb.dart';
import 'package:nordvpn/pb/daemon/ping.pb.dart';
import 'package:nordvpn/pb/daemon/purchase.pb.dart';
import 'package:nordvpn/pb/daemon/rate.pb.dart';
import 'package:nordvpn/pb/daemon/servers.pb.dart';
import 'package:nordvpn/pb/daemon/service.pbgrpc.dart';
import 'package:nordvpn/pb/daemon/set.pb.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:nordvpn/pb/daemon/state.pb.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';
import 'package:nordvpn/pb/daemon/token.pb.dart';

// The mocked implementation for the daemon. It is created by the GrpcServer.
final class MockDaemon extends DaemonServiceBase {
  final appStateStream = StreamController<AppState>.broadcast();
  late final MockApplicationSettings appSettings;
  late final MockServersList serversList;
  late final MockAccountInfo account;
  late final MockVpnStatus vpnStatus;

  MockDaemon() {
    serversList = MockServersList(appStateStream);
    appSettings = MockApplicationSettings(appStateStream, serversList);
    account = MockAccountInfo(appStateStream, serversList);
    vpnStatus = MockVpnStatus(appStateStream, appSettings, serversList);
  }

  void dispose() {
    serversList.dispose();
  }

  @override
  Future<AccountResponse> accountInfo(
    ServiceCall call,
    AccountRequest request,
  ) async {
    return account.account;
  }

  @override
  Future<ServerGroupsList> cities(ServiceCall call, CitiesRequest request) {
    throw UnimplementedError();
  }

  @override
  Future<ClaimOnlinePurchaseResponse> claimOnlinePurchase(
    ServiceCall call,
    Empty request,
  ) {
    throw UnimplementedError();
  }

  @override
  Stream<Payload> connect(ServiceCall call, ConnectRequest request) async* {
    if (await account.isExpired()) {
      yield Payload(type: Int64(DaemonStatusCode.accountExpired));
    } else {
      yield* vpnStatus.findServerAndConnect(request);
    }
  }

  @override
  Future<Payload> connectCancel(ServiceCall call, Empty request) {
    return vpnStatus.cancel();
  }

  @override
  Future<ServerGroupsList> countries(ServiceCall call, Empty request) {
    throw UnimplementedError();
  }

  @override
  Stream<Payload> disconnect(ServiceCall call, Empty request) {
    return vpnStatus.disconnect();
  }

  @override
  Future<ServersResponse> getServers(ServiceCall call, Empty request) async {
    if (serversList.error != null) {
      throw serversList.error!;
    }

    return serversList.serversList;
  }

  @override
  Future<ServerGroupsList> groups(ServiceCall call, Empty request) {
    throw UnimplementedError();
  }

  @override
  Future<grpc.IsLoggedInResponse> isLoggedIn(
    ServiceCall call,
    Empty request,
  ) async {
    return grpc.IsLoggedInResponse(isLoggedIn: await account.isLoggedIn());
  }

  @override
  Future<grpc.LoginOAuth2Response> loginOAuth2(
    ServiceCall call,
    grpc.LoginOAuth2Request request,
  ) async {
    return await account.login();
  }

  @override
  Future<grpc.LoginOAuth2CallbackResponse> loginOAuth2Callback(
    ServiceCall call,
    grpc.LoginOAuth2CallbackRequest request,
  ) {
    throw UnimplementedError();
  }

  @override
  Future<grpc.LoginResponse> loginWithToken(
    ServiceCall call,
    LoginWithTokenRequest request,
  ) {
    throw UnimplementedError();
  }

  @override
  Future<Payload> logout(ServiceCall call, LogoutRequest request) async {
    return account.logout();
  }

  @override
  Future<PingResponse> ping(ServiceCall call, Empty request) async {
    return PingResponse();
  }

  @override
  Future<Payload> rateConnection(ServiceCall call, RateRequest request) {
    throw UnimplementedError();
  }

  @override
  Future<Payload> setAllowlist(ServiceCall call, SetAllowlistRequest request) {
    return appSettings.changeAllowList(request, true);
  }

  @override
  Future<Payload> setAnalytics(ServiceCall call, SetGenericRequest request) {
    final consentValue = request.enabled
        ? ConsentMode.GRANTED
        : ConsentMode.DENIED;
    return appSettings.setSettings(analyticsConsent: consentValue);
  }

  @override
  Future<Payload> setAutoConnect(
    ServiceCall call,
    SetAutoconnectRequest request,
  ) {
    return appSettings.setAutoConnect(request);
  }

  @override
  Future<SetDNSResponse> setDNS(ServiceCall call, SetDNSRequest request) {
    return appSettings.setDNS(request);
  }

  @override
  Future<Payload> setDefaults(
    ServiceCall call,
    SetDefaultsRequest request,
  ) async {
    final payload = await appSettings.setDefaults();
    if (payload.type.toInt() != DaemonStatusCode.success) {
      return payload;
    }

    await for (final _ in vpnStatus.disconnect()) {}

    return payload;
  }

  @override
  Future<Payload> setFirewall(ServiceCall call, SetGenericRequest request) {
    return appSettings.setSettings(firewall: request.enabled);
  }

  @override
  Future<Payload> setFirewallMark(ServiceCall call, SetUint32Request request) {
    return appSettings.setSettings(fwmark: request.value);
  }

  @override
  Future<Payload> setIpv6(ServiceCall call, SetGenericRequest request) {
    throw UnimplementedError();
  }

  @override
  Future<Payload> setKillSwitch(
    ServiceCall call,
    SetKillSwitchRequest request,
  ) {
    return appSettings.setSettings(killSwitch: request.killSwitch);
  }

  @override
  Future<SetLANDiscoveryResponse> setLANDiscovery(
    ServiceCall call,
    SetLANDiscoveryRequest request,
  ) async {
    return appSettings.setLanDiscovery(request.enabled);
  }

  @override
  Future<Payload> setNotify(ServiceCall call, SetNotifyRequest request) {
    return appSettings.setSettings(notify: request.notify);
  }

  @override
  Future<Payload> setObfuscate(ServiceCall call, SetGenericRequest request) {
    return appSettings.setSettings(obfuscate: request.enabled);
  }

  @override
  Future<Payload> setPostQuantum(ServiceCall call, SetGenericRequest request) {
    return appSettings.setSettings(postquantumVpn: request.enabled);
  }

  @override
  Future<SetProtocolResponse> setProtocol(
    ServiceCall call,
    SetProtocolRequest request,
  ) {
    return appSettings.setProtocol(request);
  }

  @override
  Future<Payload> setRouting(ServiceCall call, SetGenericRequest request) {
    return appSettings.setSettings(routing: request.enabled);
  }

  @override
  Future<Payload> setTechnology(
    ServiceCall call,
    SetTechnologyRequest request,
  ) {
    return appSettings.setSettings(technology: request.technology);
  }

  @override
  Future<SetThreatProtectionLiteResponse> setThreatProtectionLite(
    ServiceCall call,
    SetThreatProtectionLiteRequest request,
  ) {
    return appSettings.setThreatProtectionLite(request);
  }

  @override
  Future<Payload> setTray(ServiceCall call, SetTrayRequest request) {
    return appSettings.setSettings(tray: request.tray);
  }

  @override
  Future<Payload> setVirtualLocation(
    ServiceCall call,
    SetGenericRequest request,
  ) {
    return appSettings.setSettings(virtualLocation: request.enabled);
  }

  @override
  Future<SettingsResponse> settings(ServiceCall call, Empty request) async {
    if (appSettings.error != null) {
      throw appSettings.error!;
    }

    return appSettings.settings;
  }

  @override
  Future<Payload> settingsProtocols(ServiceCall call, Empty request) {
    throw UnimplementedError();
  }

  @override
  Future<Payload> settingsTechnologies(ServiceCall call, Empty request) {
    throw UnimplementedError();
  }

  @override
  Future<StatusResponse> status(ServiceCall call, Empty request) async {
    if (vpnStatus.error != null) {
      throw vpnStatus.error!;
    }

    return vpnStatus.status;
  }

  @override
  Stream<AppState> subscribeToStateChanges(ServiceCall call, Empty request) {
    return appStateStream.stream;
  }

  @override
  Future<TokenInfoResponse> tokenInfo(ServiceCall call, Empty request) {
    throw UnimplementedError();
  }

  @override
  Future<Payload> unsetAllAllowlist(ServiceCall call, Empty request) {
    return appSettings.setSettings(allowList: Allowlist());
  }

  @override
  Future<Payload> unsetAllowlist(
    ServiceCall call,
    SetAllowlistRequest request,
  ) {
    return appSettings.changeAllowList(request, false);
  }

  @override
  Future<GetDaemonApiVersionResponse> getDaemonApiVersion(
    ServiceCall call,
    GetDaemonApiVersionRequest request,
  ) async {
    return GetDaemonApiVersionResponse(
      apiVersion: DaemonApiVersion.CURRENT_VERSION.value,
    );
  }
}
