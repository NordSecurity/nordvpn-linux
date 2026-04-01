// This is a generated file - do not edit.
//
// Generated from service.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'package:protobuf/protobuf.dart' as $pb;

import 'account.pb.dart' as $4;
import 'cities.pb.dart' as $12;
import 'common.pb.dart' as $0;
import 'connect.pb.dart' as $7;
import 'defaults.pb.dart' as $14;
import 'features.pb.dart' as $17;
import 'login.pb.dart' as $1;
import 'login_with_token.pb.dart' as $2;
import 'logout.pb.dart' as $3;
import 'pause.pb.dart' as $10;
import 'ping.pb.dart' as $18;
import 'purchase.pb.dart' as $6;
import 'rate.pb.dart' as $9;
import 'recent_connections.pb.dart' as $16;
import 'servers.pb.dart' as $11;
import 'set.pb.dart' as $15;
import 'settings.pb.dart' as $13;
import 'state.pb.dart' as $19;
import 'status.pb.dart' as $8;
import 'token.pb.dart' as $5;

export 'service.pb.dart';

@$pb.GrpcServiceName('pb.Daemon')
class DaemonClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  DaemonClient(super.channel, {super.options, super.interceptors});

  /// ==================== Authentication ====================
  $grpc.ResponseFuture<$1.IsLoggedInResponse> isLoggedIn(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$isLoggedIn, request, options: options);
  }

  $grpc.ResponseFuture<$1.LoginResponse> loginWithToken(
    $2.LoginWithTokenRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$loginWithToken, request, options: options);
  }

  $grpc.ResponseFuture<$1.LoginOAuth2Response> loginOAuth2(
    $1.LoginOAuth2Request request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$loginOAuth2, request, options: options);
  }

  $grpc.ResponseFuture<$1.LoginOAuth2CallbackResponse> loginOAuth2Callback(
    $1.LoginOAuth2CallbackRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$loginOAuth2Callback, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> logout(
    $3.LogoutRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$logout, request, options: options);
  }

  /// ==================== Account Management ====================
  $grpc.ResponseFuture<$4.AccountResponse> accountInfo(
    $4.AccountRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$accountInfo, request, options: options);
  }

  $grpc.ResponseFuture<$5.TokenInfoResponse> tokenInfo(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$tokenInfo, request, options: options);
  }

  $grpc.ResponseFuture<$6.ClaimOnlinePurchaseResponse> claimOnlinePurchase(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$claimOnlinePurchase, request, options: options);
  }

  /// ==================== Connection Operations ====================
  $grpc.ResponseStream<$0.Payload> connect(
    $7.ConnectRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(
        _$connect, $async.Stream.fromIterable([request]),
        options: options);
  }

  $grpc.ResponseFuture<$0.Payload> connectCancel(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$connectCancel, request, options: options);
  }

  $grpc.ResponseStream<$0.Payload> disconnect(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(
        _$disconnect, $async.Stream.fromIterable([request]),
        options: options);
  }

  $grpc.ResponseFuture<$8.StatusResponse> status(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$status, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> rateConnection(
    $9.RateRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$rateConnection, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> pauseConnection(
    $10.PauseRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$pauseConnection, request, options: options);
  }

  /// ==================== Server Discovery ====================
  $grpc.ResponseFuture<$11.ServersResponse> getServers(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getServers, request, options: options);
  }

  $grpc.ResponseFuture<$0.ServerGroupsList> countries(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$countries, request, options: options);
  }

  $grpc.ResponseFuture<$0.ServerGroupsList> cities(
    $12.CitiesRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$cities, request, options: options);
  }

  $grpc.ResponseFuture<$0.ServerGroupsList> groups(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$groups, request, options: options);
  }

  /// ==================== General Settings ====================
  $grpc.ResponseFuture<$13.SettingsResponse> settings(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$settings, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setDefaults(
    $14.SetDefaultsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setDefaults, request, options: options);
  }

  /// ==================== Connection Settings ====================
  $grpc.ResponseFuture<$0.Payload> setAutoConnect(
    $15.SetAutoconnectRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setAutoConnect, request, options: options);
  }

  $grpc.ResponseFuture<$15.SetProtocolResponse> setProtocol(
    $15.SetProtocolRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setProtocol, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setTechnology(
    $15.SetTechnologyRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setTechnology, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setObfuscate(
    $15.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setObfuscate, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setPostQuantum(
    $15.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setPostQuantum, request, options: options);
  }

  $grpc.ResponseFuture<$16.RecentConnectionsResponse> getRecentConnections(
    $16.RecentConnectionsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getRecentConnections, request, options: options);
  }

  /// ==================== Network Settings ====================
  $grpc.ResponseFuture<$15.SetDNSResponse> setDNS(
    $15.SetDNSRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setDNS, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setFirewall(
    $15.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setFirewall, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setFirewallMark(
    $15.SetUint32Request request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setFirewallMark, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setRouting(
    $15.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setRouting, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setKillSwitch(
    $15.SetKillSwitchRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setKillSwitch, request, options: options);
  }

  $grpc.ResponseFuture<$15.SetLANDiscoveryResponse> setLANDiscovery(
    $15.SetLANDiscoveryRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setLANDiscovery, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setVirtualLocation(
    $15.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setVirtualLocation, request, options: options);
  }

  /// ==================== UI Settings ====================
  $grpc.ResponseFuture<$0.Payload> setNotify(
    $15.SetNotifyRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setNotify, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setTray(
    $15.SetTrayRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setTray, request, options: options);
  }

  /// ==================== Configuration Info ====================
  $grpc.ResponseFuture<$0.Payload> settingsProtocols(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$settingsProtocols, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> settingsTechnologies(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$settingsTechnologies, request, options: options);
  }

  $grpc.ResponseFuture<$17.FeatureToggles> getFeatureToggles(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getFeatureToggles, request, options: options);
  }

  /// ==================== Allowlist Management ====================
  $grpc.ResponseFuture<$0.Payload> setAllowlist(
    $15.SetAllowlistRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setAllowlist, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setARPIgnore(
    $15.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setARPIgnore, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> unsetAllowlist(
    $15.SetAllowlistRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$unsetAllowlist, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> unsetAllAllowlist(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$unsetAllAllowlist, request, options: options);
  }

  /// ==================== Privacy & Security ====================
  $grpc.ResponseFuture<$0.Payload> setAnalytics(
    $15.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setAnalytics, request, options: options);
  }

  $grpc.ResponseFuture<$15.SetThreatProtectionLiteResponse>
      setThreatProtectionLite(
    $15.SetThreatProtectionLiteRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setThreatProtectionLite, request,
        options: options);
  }

  /// ==================== System & Monitoring ====================
  $grpc.ResponseFuture<$18.PingResponse> ping(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$ping, request, options: options);
  }

  $grpc.ResponseStream<$19.AppState> subscribeToStateChanges(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(
        _$subscribeToStateChanges, $async.Stream.fromIterable([request]),
        options: options);
  }

  $grpc.ResponseFuture<$0.GetDaemonApiVersionResponse> getDaemonApiVersion(
    $0.GetDaemonApiVersionRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getDaemonApiVersion, request, options: options);
  }

  // method descriptors

  static final _$isLoggedIn =
      $grpc.ClientMethod<$0.Empty, $1.IsLoggedInResponse>(
          '/pb.Daemon/IsLoggedIn',
          ($0.Empty value) => value.writeToBuffer(),
          $1.IsLoggedInResponse.fromBuffer);
  static final _$loginWithToken =
      $grpc.ClientMethod<$2.LoginWithTokenRequest, $1.LoginResponse>(
          '/pb.Daemon/LoginWithToken',
          ($2.LoginWithTokenRequest value) => value.writeToBuffer(),
          $1.LoginResponse.fromBuffer);
  static final _$loginOAuth2 =
      $grpc.ClientMethod<$1.LoginOAuth2Request, $1.LoginOAuth2Response>(
          '/pb.Daemon/LoginOAuth2',
          ($1.LoginOAuth2Request value) => value.writeToBuffer(),
          $1.LoginOAuth2Response.fromBuffer);
  static final _$loginOAuth2Callback = $grpc.ClientMethod<
          $1.LoginOAuth2CallbackRequest, $1.LoginOAuth2CallbackResponse>(
      '/pb.Daemon/LoginOAuth2Callback',
      ($1.LoginOAuth2CallbackRequest value) => value.writeToBuffer(),
      $1.LoginOAuth2CallbackResponse.fromBuffer);
  static final _$logout = $grpc.ClientMethod<$3.LogoutRequest, $0.Payload>(
      '/pb.Daemon/Logout',
      ($3.LogoutRequest value) => value.writeToBuffer(),
      $0.Payload.fromBuffer);
  static final _$accountInfo =
      $grpc.ClientMethod<$4.AccountRequest, $4.AccountResponse>(
          '/pb.Daemon/AccountInfo',
          ($4.AccountRequest value) => value.writeToBuffer(),
          $4.AccountResponse.fromBuffer);
  static final _$tokenInfo = $grpc.ClientMethod<$0.Empty, $5.TokenInfoResponse>(
      '/pb.Daemon/TokenInfo',
      ($0.Empty value) => value.writeToBuffer(),
      $5.TokenInfoResponse.fromBuffer);
  static final _$claimOnlinePurchase =
      $grpc.ClientMethod<$0.Empty, $6.ClaimOnlinePurchaseResponse>(
          '/pb.Daemon/ClaimOnlinePurchase',
          ($0.Empty value) => value.writeToBuffer(),
          $6.ClaimOnlinePurchaseResponse.fromBuffer);
  static final _$connect = $grpc.ClientMethod<$7.ConnectRequest, $0.Payload>(
      '/pb.Daemon/Connect',
      ($7.ConnectRequest value) => value.writeToBuffer(),
      $0.Payload.fromBuffer);
  static final _$connectCancel = $grpc.ClientMethod<$0.Empty, $0.Payload>(
      '/pb.Daemon/ConnectCancel',
      ($0.Empty value) => value.writeToBuffer(),
      $0.Payload.fromBuffer);
  static final _$disconnect = $grpc.ClientMethod<$0.Empty, $0.Payload>(
      '/pb.Daemon/Disconnect',
      ($0.Empty value) => value.writeToBuffer(),
      $0.Payload.fromBuffer);
  static final _$status = $grpc.ClientMethod<$0.Empty, $8.StatusResponse>(
      '/pb.Daemon/Status',
      ($0.Empty value) => value.writeToBuffer(),
      $8.StatusResponse.fromBuffer);
  static final _$rateConnection =
      $grpc.ClientMethod<$9.RateRequest, $0.Payload>(
          '/pb.Daemon/RateConnection',
          ($9.RateRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$pauseConnection =
      $grpc.ClientMethod<$10.PauseRequest, $0.Payload>(
          '/pb.Daemon/PauseConnection',
          ($10.PauseRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$getServers = $grpc.ClientMethod<$0.Empty, $11.ServersResponse>(
      '/pb.Daemon/GetServers',
      ($0.Empty value) => value.writeToBuffer(),
      $11.ServersResponse.fromBuffer);
  static final _$countries = $grpc.ClientMethod<$0.Empty, $0.ServerGroupsList>(
      '/pb.Daemon/Countries',
      ($0.Empty value) => value.writeToBuffer(),
      $0.ServerGroupsList.fromBuffer);
  static final _$cities =
      $grpc.ClientMethod<$12.CitiesRequest, $0.ServerGroupsList>(
          '/pb.Daemon/Cities',
          ($12.CitiesRequest value) => value.writeToBuffer(),
          $0.ServerGroupsList.fromBuffer);
  static final _$groups = $grpc.ClientMethod<$0.Empty, $0.ServerGroupsList>(
      '/pb.Daemon/Groups',
      ($0.Empty value) => value.writeToBuffer(),
      $0.ServerGroupsList.fromBuffer);
  static final _$settings = $grpc.ClientMethod<$0.Empty, $13.SettingsResponse>(
      '/pb.Daemon/Settings',
      ($0.Empty value) => value.writeToBuffer(),
      $13.SettingsResponse.fromBuffer);
  static final _$setDefaults =
      $grpc.ClientMethod<$14.SetDefaultsRequest, $0.Payload>(
          '/pb.Daemon/SetDefaults',
          ($14.SetDefaultsRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setAutoConnect =
      $grpc.ClientMethod<$15.SetAutoconnectRequest, $0.Payload>(
          '/pb.Daemon/SetAutoConnect',
          ($15.SetAutoconnectRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setProtocol =
      $grpc.ClientMethod<$15.SetProtocolRequest, $15.SetProtocolResponse>(
          '/pb.Daemon/SetProtocol',
          ($15.SetProtocolRequest value) => value.writeToBuffer(),
          $15.SetProtocolResponse.fromBuffer);
  static final _$setTechnology =
      $grpc.ClientMethod<$15.SetTechnologyRequest, $0.Payload>(
          '/pb.Daemon/SetTechnology',
          ($15.SetTechnologyRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setObfuscate =
      $grpc.ClientMethod<$15.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetObfuscate',
          ($15.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setPostQuantum =
      $grpc.ClientMethod<$15.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetPostQuantum',
          ($15.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$getRecentConnections = $grpc.ClientMethod<
          $16.RecentConnectionsRequest, $16.RecentConnectionsResponse>(
      '/pb.Daemon/GetRecentConnections',
      ($16.RecentConnectionsRequest value) => value.writeToBuffer(),
      $16.RecentConnectionsResponse.fromBuffer);
  static final _$setDNS =
      $grpc.ClientMethod<$15.SetDNSRequest, $15.SetDNSResponse>(
          '/pb.Daemon/SetDNS',
          ($15.SetDNSRequest value) => value.writeToBuffer(),
          $15.SetDNSResponse.fromBuffer);
  static final _$setFirewall =
      $grpc.ClientMethod<$15.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetFirewall',
          ($15.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setFirewallMark =
      $grpc.ClientMethod<$15.SetUint32Request, $0.Payload>(
          '/pb.Daemon/SetFirewallMark',
          ($15.SetUint32Request value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setRouting =
      $grpc.ClientMethod<$15.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetRouting',
          ($15.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setKillSwitch =
      $grpc.ClientMethod<$15.SetKillSwitchRequest, $0.Payload>(
          '/pb.Daemon/SetKillSwitch',
          ($15.SetKillSwitchRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setLANDiscovery = $grpc.ClientMethod<
          $15.SetLANDiscoveryRequest, $15.SetLANDiscoveryResponse>(
      '/pb.Daemon/SetLANDiscovery',
      ($15.SetLANDiscoveryRequest value) => value.writeToBuffer(),
      $15.SetLANDiscoveryResponse.fromBuffer);
  static final _$setVirtualLocation =
      $grpc.ClientMethod<$15.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetVirtualLocation',
          ($15.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setNotify =
      $grpc.ClientMethod<$15.SetNotifyRequest, $0.Payload>(
          '/pb.Daemon/SetNotify',
          ($15.SetNotifyRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setTray = $grpc.ClientMethod<$15.SetTrayRequest, $0.Payload>(
      '/pb.Daemon/SetTray',
      ($15.SetTrayRequest value) => value.writeToBuffer(),
      $0.Payload.fromBuffer);
  static final _$settingsProtocols = $grpc.ClientMethod<$0.Empty, $0.Payload>(
      '/pb.Daemon/SettingsProtocols',
      ($0.Empty value) => value.writeToBuffer(),
      $0.Payload.fromBuffer);
  static final _$settingsTechnologies =
      $grpc.ClientMethod<$0.Empty, $0.Payload>(
          '/pb.Daemon/SettingsTechnologies',
          ($0.Empty value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$getFeatureToggles =
      $grpc.ClientMethod<$0.Empty, $17.FeatureToggles>(
          '/pb.Daemon/GetFeatureToggles',
          ($0.Empty value) => value.writeToBuffer(),
          $17.FeatureToggles.fromBuffer);
  static final _$setAllowlist =
      $grpc.ClientMethod<$15.SetAllowlistRequest, $0.Payload>(
          '/pb.Daemon/SetAllowlist',
          ($15.SetAllowlistRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setARPIgnore =
      $grpc.ClientMethod<$15.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetARPIgnore',
          ($15.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$unsetAllowlist =
      $grpc.ClientMethod<$15.SetAllowlistRequest, $0.Payload>(
          '/pb.Daemon/UnsetAllowlist',
          ($15.SetAllowlistRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$unsetAllAllowlist = $grpc.ClientMethod<$0.Empty, $0.Payload>(
      '/pb.Daemon/UnsetAllAllowlist',
      ($0.Empty value) => value.writeToBuffer(),
      $0.Payload.fromBuffer);
  static final _$setAnalytics =
      $grpc.ClientMethod<$15.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetAnalytics',
          ($15.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setThreatProtectionLite = $grpc.ClientMethod<
          $15.SetThreatProtectionLiteRequest,
          $15.SetThreatProtectionLiteResponse>(
      '/pb.Daemon/SetThreatProtectionLite',
      ($15.SetThreatProtectionLiteRequest value) => value.writeToBuffer(),
      $15.SetThreatProtectionLiteResponse.fromBuffer);
  static final _$ping = $grpc.ClientMethod<$0.Empty, $18.PingResponse>(
      '/pb.Daemon/Ping',
      ($0.Empty value) => value.writeToBuffer(),
      $18.PingResponse.fromBuffer);
  static final _$subscribeToStateChanges =
      $grpc.ClientMethod<$0.Empty, $19.AppState>(
          '/pb.Daemon/SubscribeToStateChanges',
          ($0.Empty value) => value.writeToBuffer(),
          $19.AppState.fromBuffer);
  static final _$getDaemonApiVersion = $grpc.ClientMethod<
          $0.GetDaemonApiVersionRequest, $0.GetDaemonApiVersionResponse>(
      '/pb.Daemon/GetDaemonApiVersion',
      ($0.GetDaemonApiVersionRequest value) => value.writeToBuffer(),
      $0.GetDaemonApiVersionResponse.fromBuffer);
}

@$pb.GrpcServiceName('pb.Daemon')
abstract class DaemonServiceBase extends $grpc.Service {
  $core.String get $name => 'pb.Daemon';

  DaemonServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.IsLoggedInResponse>(
        'IsLoggedIn',
        isLoggedIn_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.IsLoggedInResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.LoginWithTokenRequest, $1.LoginResponse>(
        'LoginWithToken',
        loginWithToken_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.LoginWithTokenRequest.fromBuffer(value),
        ($1.LoginResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$1.LoginOAuth2Request, $1.LoginOAuth2Response>(
            'LoginOAuth2',
            loginOAuth2_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.LoginOAuth2Request.fromBuffer(value),
            ($1.LoginOAuth2Response value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.LoginOAuth2CallbackRequest,
            $1.LoginOAuth2CallbackResponse>(
        'LoginOAuth2Callback',
        loginOAuth2Callback_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.LoginOAuth2CallbackRequest.fromBuffer(value),
        ($1.LoginOAuth2CallbackResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.LogoutRequest, $0.Payload>(
        'Logout',
        logout_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.LogoutRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$4.AccountRequest, $4.AccountResponse>(
        'AccountInfo',
        accountInfo_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $4.AccountRequest.fromBuffer(value),
        ($4.AccountResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $5.TokenInfoResponse>(
        'TokenInfo',
        tokenInfo_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($5.TokenInfoResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $6.ClaimOnlinePurchaseResponse>(
        'ClaimOnlinePurchase',
        claimOnlinePurchase_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($6.ClaimOnlinePurchaseResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$7.ConnectRequest, $0.Payload>(
        'Connect',
        connect_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $7.ConnectRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.Payload>(
        'ConnectCancel',
        connectCancel_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.Payload>(
        'Disconnect',
        disconnect_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $8.StatusResponse>(
        'Status',
        status_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($8.StatusResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$9.RateRequest, $0.Payload>(
        'RateConnection',
        rateConnection_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $9.RateRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.PauseRequest, $0.Payload>(
        'PauseConnection',
        pauseConnection_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $10.PauseRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $11.ServersResponse>(
        'GetServers',
        getServers_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($11.ServersResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.ServerGroupsList>(
        'Countries',
        countries_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.ServerGroupsList value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$12.CitiesRequest, $0.ServerGroupsList>(
        'Cities',
        cities_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $12.CitiesRequest.fromBuffer(value),
        ($0.ServerGroupsList value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.ServerGroupsList>(
        'Groups',
        groups_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.ServerGroupsList value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $13.SettingsResponse>(
        'Settings',
        settings_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($13.SettingsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetDefaultsRequest, $0.Payload>(
        'SetDefaults',
        setDefaults_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetDefaultsRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetAutoconnectRequest, $0.Payload>(
        'SetAutoConnect',
        setAutoConnect_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetAutoconnectRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$15.SetProtocolRequest, $15.SetProtocolResponse>(
            'SetProtocol',
            setProtocol_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $15.SetProtocolRequest.fromBuffer(value),
            ($15.SetProtocolResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetTechnologyRequest, $0.Payload>(
        'SetTechnology',
        setTechnology_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetTechnologyRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetGenericRequest, $0.Payload>(
        'SetObfuscate',
        setObfuscate_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetGenericRequest, $0.Payload>(
        'SetPostQuantum',
        setPostQuantum_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$16.RecentConnectionsRequest,
            $16.RecentConnectionsResponse>(
        'GetRecentConnections',
        getRecentConnections_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $16.RecentConnectionsRequest.fromBuffer(value),
        ($16.RecentConnectionsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetDNSRequest, $15.SetDNSResponse>(
        'SetDNS',
        setDNS_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $15.SetDNSRequest.fromBuffer(value),
        ($15.SetDNSResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetGenericRequest, $0.Payload>(
        'SetFirewall',
        setFirewall_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetUint32Request, $0.Payload>(
        'SetFirewallMark',
        setFirewallMark_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $15.SetUint32Request.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetGenericRequest, $0.Payload>(
        'SetRouting',
        setRouting_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetKillSwitchRequest, $0.Payload>(
        'SetKillSwitch',
        setKillSwitch_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetKillSwitchRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetLANDiscoveryRequest,
            $15.SetLANDiscoveryResponse>(
        'SetLANDiscovery',
        setLANDiscovery_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetLANDiscoveryRequest.fromBuffer(value),
        ($15.SetLANDiscoveryResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetGenericRequest, $0.Payload>(
        'SetVirtualLocation',
        setVirtualLocation_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetNotifyRequest, $0.Payload>(
        'SetNotify',
        setNotify_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $15.SetNotifyRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetTrayRequest, $0.Payload>(
        'SetTray',
        setTray_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $15.SetTrayRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.Payload>(
        'SettingsProtocols',
        settingsProtocols_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.Payload>(
        'SettingsTechnologies',
        settingsTechnologies_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $17.FeatureToggles>(
        'GetFeatureToggles',
        getFeatureToggles_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($17.FeatureToggles value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetAllowlistRequest, $0.Payload>(
        'SetAllowlist',
        setAllowlist_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetAllowlistRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetGenericRequest, $0.Payload>(
        'SetARPIgnore',
        setARPIgnore_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetAllowlistRequest, $0.Payload>(
        'UnsetAllowlist',
        unsetAllowlist_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetAllowlistRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.Payload>(
        'UnsetAllAllowlist',
        unsetAllAllowlist_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetGenericRequest, $0.Payload>(
        'SetAnalytics',
        setAnalytics_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.SetThreatProtectionLiteRequest,
            $15.SetThreatProtectionLiteResponse>(
        'SetThreatProtectionLite',
        setThreatProtectionLite_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.SetThreatProtectionLiteRequest.fromBuffer(value),
        ($15.SetThreatProtectionLiteResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $18.PingResponse>(
        'Ping',
        ping_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($18.PingResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $19.AppState>(
        'SubscribeToStateChanges',
        subscribeToStateChanges_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($19.AppState value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.GetDaemonApiVersionRequest,
            $0.GetDaemonApiVersionResponse>(
        'GetDaemonApiVersion',
        getDaemonApiVersion_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.GetDaemonApiVersionRequest.fromBuffer(value),
        ($0.GetDaemonApiVersionResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.IsLoggedInResponse> isLoggedIn_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return isLoggedIn($call, await $request);
  }

  $async.Future<$1.IsLoggedInResponse> isLoggedIn(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$1.LoginResponse> loginWithToken_Pre($grpc.ServiceCall $call,
      $async.Future<$2.LoginWithTokenRequest> $request) async {
    return loginWithToken($call, await $request);
  }

  $async.Future<$1.LoginResponse> loginWithToken(
      $grpc.ServiceCall call, $2.LoginWithTokenRequest request);

  $async.Future<$1.LoginOAuth2Response> loginOAuth2_Pre($grpc.ServiceCall $call,
      $async.Future<$1.LoginOAuth2Request> $request) async {
    return loginOAuth2($call, await $request);
  }

  $async.Future<$1.LoginOAuth2Response> loginOAuth2(
      $grpc.ServiceCall call, $1.LoginOAuth2Request request);

  $async.Future<$1.LoginOAuth2CallbackResponse> loginOAuth2Callback_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$1.LoginOAuth2CallbackRequest> $request) async {
    return loginOAuth2Callback($call, await $request);
  }

  $async.Future<$1.LoginOAuth2CallbackResponse> loginOAuth2Callback(
      $grpc.ServiceCall call, $1.LoginOAuth2CallbackRequest request);

  $async.Future<$0.Payload> logout_Pre(
      $grpc.ServiceCall $call, $async.Future<$3.LogoutRequest> $request) async {
    return logout($call, await $request);
  }

  $async.Future<$0.Payload> logout(
      $grpc.ServiceCall call, $3.LogoutRequest request);

  $async.Future<$4.AccountResponse> accountInfo_Pre($grpc.ServiceCall $call,
      $async.Future<$4.AccountRequest> $request) async {
    return accountInfo($call, await $request);
  }

  $async.Future<$4.AccountResponse> accountInfo(
      $grpc.ServiceCall call, $4.AccountRequest request);

  $async.Future<$5.TokenInfoResponse> tokenInfo_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return tokenInfo($call, await $request);
  }

  $async.Future<$5.TokenInfoResponse> tokenInfo(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$6.ClaimOnlinePurchaseResponse> claimOnlinePurchase_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return claimOnlinePurchase($call, await $request);
  }

  $async.Future<$6.ClaimOnlinePurchaseResponse> claimOnlinePurchase(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Stream<$0.Payload> connect_Pre($grpc.ServiceCall $call,
      $async.Future<$7.ConnectRequest> $request) async* {
    yield* connect($call, await $request);
  }

  $async.Stream<$0.Payload> connect(
      $grpc.ServiceCall call, $7.ConnectRequest request);

  $async.Future<$0.Payload> connectCancel_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return connectCancel($call, await $request);
  }

  $async.Future<$0.Payload> connectCancel(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Stream<$0.Payload> disconnect_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async* {
    yield* disconnect($call, await $request);
  }

  $async.Stream<$0.Payload> disconnect(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$8.StatusResponse> status_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return status($call, await $request);
  }

  $async.Future<$8.StatusResponse> status(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.Payload> rateConnection_Pre(
      $grpc.ServiceCall $call, $async.Future<$9.RateRequest> $request) async {
    return rateConnection($call, await $request);
  }

  $async.Future<$0.Payload> rateConnection(
      $grpc.ServiceCall call, $9.RateRequest request);

  $async.Future<$0.Payload> pauseConnection_Pre(
      $grpc.ServiceCall $call, $async.Future<$10.PauseRequest> $request) async {
    return pauseConnection($call, await $request);
  }

  $async.Future<$0.Payload> pauseConnection(
      $grpc.ServiceCall call, $10.PauseRequest request);

  $async.Future<$11.ServersResponse> getServers_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return getServers($call, await $request);
  }

  $async.Future<$11.ServersResponse> getServers(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.ServerGroupsList> countries_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return countries($call, await $request);
  }

  $async.Future<$0.ServerGroupsList> countries(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.ServerGroupsList> cities_Pre($grpc.ServiceCall $call,
      $async.Future<$12.CitiesRequest> $request) async {
    return cities($call, await $request);
  }

  $async.Future<$0.ServerGroupsList> cities(
      $grpc.ServiceCall call, $12.CitiesRequest request);

  $async.Future<$0.ServerGroupsList> groups_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return groups($call, await $request);
  }

  $async.Future<$0.ServerGroupsList> groups(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$13.SettingsResponse> settings_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return settings($call, await $request);
  }

  $async.Future<$13.SettingsResponse> settings(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.Payload> setDefaults_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetDefaultsRequest> $request) async {
    return setDefaults($call, await $request);
  }

  $async.Future<$0.Payload> setDefaults(
      $grpc.ServiceCall call, $14.SetDefaultsRequest request);

  $async.Future<$0.Payload> setAutoConnect_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetAutoconnectRequest> $request) async {
    return setAutoConnect($call, await $request);
  }

  $async.Future<$0.Payload> setAutoConnect(
      $grpc.ServiceCall call, $15.SetAutoconnectRequest request);

  $async.Future<$15.SetProtocolResponse> setProtocol_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$15.SetProtocolRequest> $request) async {
    return setProtocol($call, await $request);
  }

  $async.Future<$15.SetProtocolResponse> setProtocol(
      $grpc.ServiceCall call, $15.SetProtocolRequest request);

  $async.Future<$0.Payload> setTechnology_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetTechnologyRequest> $request) async {
    return setTechnology($call, await $request);
  }

  $async.Future<$0.Payload> setTechnology(
      $grpc.ServiceCall call, $15.SetTechnologyRequest request);

  $async.Future<$0.Payload> setObfuscate_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetGenericRequest> $request) async {
    return setObfuscate($call, await $request);
  }

  $async.Future<$0.Payload> setObfuscate(
      $grpc.ServiceCall call, $15.SetGenericRequest request);

  $async.Future<$0.Payload> setPostQuantum_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetGenericRequest> $request) async {
    return setPostQuantum($call, await $request);
  }

  $async.Future<$0.Payload> setPostQuantum(
      $grpc.ServiceCall call, $15.SetGenericRequest request);

  $async.Future<$16.RecentConnectionsResponse> getRecentConnections_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$16.RecentConnectionsRequest> $request) async {
    return getRecentConnections($call, await $request);
  }

  $async.Future<$16.RecentConnectionsResponse> getRecentConnections(
      $grpc.ServiceCall call, $16.RecentConnectionsRequest request);

  $async.Future<$15.SetDNSResponse> setDNS_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetDNSRequest> $request) async {
    return setDNS($call, await $request);
  }

  $async.Future<$15.SetDNSResponse> setDNS(
      $grpc.ServiceCall call, $15.SetDNSRequest request);

  $async.Future<$0.Payload> setFirewall_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetGenericRequest> $request) async {
    return setFirewall($call, await $request);
  }

  $async.Future<$0.Payload> setFirewall(
      $grpc.ServiceCall call, $15.SetGenericRequest request);

  $async.Future<$0.Payload> setFirewallMark_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetUint32Request> $request) async {
    return setFirewallMark($call, await $request);
  }

  $async.Future<$0.Payload> setFirewallMark(
      $grpc.ServiceCall call, $15.SetUint32Request request);

  $async.Future<$0.Payload> setRouting_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetGenericRequest> $request) async {
    return setRouting($call, await $request);
  }

  $async.Future<$0.Payload> setRouting(
      $grpc.ServiceCall call, $15.SetGenericRequest request);

  $async.Future<$0.Payload> setKillSwitch_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetKillSwitchRequest> $request) async {
    return setKillSwitch($call, await $request);
  }

  $async.Future<$0.Payload> setKillSwitch(
      $grpc.ServiceCall call, $15.SetKillSwitchRequest request);

  $async.Future<$15.SetLANDiscoveryResponse> setLANDiscovery_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$15.SetLANDiscoveryRequest> $request) async {
    return setLANDiscovery($call, await $request);
  }

  $async.Future<$15.SetLANDiscoveryResponse> setLANDiscovery(
      $grpc.ServiceCall call, $15.SetLANDiscoveryRequest request);

  $async.Future<$0.Payload> setVirtualLocation_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetGenericRequest> $request) async {
    return setVirtualLocation($call, await $request);
  }

  $async.Future<$0.Payload> setVirtualLocation(
      $grpc.ServiceCall call, $15.SetGenericRequest request);

  $async.Future<$0.Payload> setNotify_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetNotifyRequest> $request) async {
    return setNotify($call, await $request);
  }

  $async.Future<$0.Payload> setNotify(
      $grpc.ServiceCall call, $15.SetNotifyRequest request);

  $async.Future<$0.Payload> setTray_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetTrayRequest> $request) async {
    return setTray($call, await $request);
  }

  $async.Future<$0.Payload> setTray(
      $grpc.ServiceCall call, $15.SetTrayRequest request);

  $async.Future<$0.Payload> settingsProtocols_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return settingsProtocols($call, await $request);
  }

  $async.Future<$0.Payload> settingsProtocols(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.Payload> settingsTechnologies_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return settingsTechnologies($call, await $request);
  }

  $async.Future<$0.Payload> settingsTechnologies(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$17.FeatureToggles> getFeatureToggles_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return getFeatureToggles($call, await $request);
  }

  $async.Future<$17.FeatureToggles> getFeatureToggles(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.Payload> setAllowlist_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetAllowlistRequest> $request) async {
    return setAllowlist($call, await $request);
  }

  $async.Future<$0.Payload> setAllowlist(
      $grpc.ServiceCall call, $15.SetAllowlistRequest request);

  $async.Future<$0.Payload> setARPIgnore_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetGenericRequest> $request) async {
    return setARPIgnore($call, await $request);
  }

  $async.Future<$0.Payload> setARPIgnore(
      $grpc.ServiceCall call, $15.SetGenericRequest request);

  $async.Future<$0.Payload> unsetAllowlist_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetAllowlistRequest> $request) async {
    return unsetAllowlist($call, await $request);
  }

  $async.Future<$0.Payload> unsetAllowlist(
      $grpc.ServiceCall call, $15.SetAllowlistRequest request);

  $async.Future<$0.Payload> unsetAllAllowlist_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return unsetAllAllowlist($call, await $request);
  }

  $async.Future<$0.Payload> unsetAllAllowlist(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.Payload> setAnalytics_Pre($grpc.ServiceCall $call,
      $async.Future<$15.SetGenericRequest> $request) async {
    return setAnalytics($call, await $request);
  }

  $async.Future<$0.Payload> setAnalytics(
      $grpc.ServiceCall call, $15.SetGenericRequest request);

  $async.Future<$15.SetThreatProtectionLiteResponse>
      setThreatProtectionLite_Pre($grpc.ServiceCall $call,
          $async.Future<$15.SetThreatProtectionLiteRequest> $request) async {
    return setThreatProtectionLite($call, await $request);
  }

  $async.Future<$15.SetThreatProtectionLiteResponse> setThreatProtectionLite(
      $grpc.ServiceCall call, $15.SetThreatProtectionLiteRequest request);

  $async.Future<$18.PingResponse> ping_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return ping($call, await $request);
  }

  $async.Future<$18.PingResponse> ping(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Stream<$19.AppState> subscribeToStateChanges_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async* {
    yield* subscribeToStateChanges($call, await $request);
  }

  $async.Stream<$19.AppState> subscribeToStateChanges(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.GetDaemonApiVersionResponse> getDaemonApiVersion_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.GetDaemonApiVersionRequest> $request) async {
    return getDaemonApiVersion($call, await $request);
  }

  $async.Future<$0.GetDaemonApiVersionResponse> getDaemonApiVersion(
      $grpc.ServiceCall call, $0.GetDaemonApiVersionRequest request);
}
