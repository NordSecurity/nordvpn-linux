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
import 'cities.pb.dart' as $11;
import 'common.pb.dart' as $0;
import 'connect.pb.dart' as $7;
import 'defaults.pb.dart' as $13;
import 'features.pb.dart' as $16;
import 'login.pb.dart' as $1;
import 'login_with_token.pb.dart' as $2;
import 'logout.pb.dart' as $3;
import 'ping.pb.dart' as $17;
import 'purchase.pb.dart' as $6;
import 'rate.pb.dart' as $9;
import 'recent_connections.pb.dart' as $15;
import 'servers.pb.dart' as $10;
import 'set.pb.dart' as $14;
import 'settings.pb.dart' as $12;
import 'state.pb.dart' as $18;
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

  /// ==================== Server Discovery ====================
  $grpc.ResponseFuture<$10.ServersResponse> getServers(
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
    $11.CitiesRequest request, {
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
  $grpc.ResponseFuture<$12.SettingsResponse> settings(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$settings, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setDefaults(
    $13.SetDefaultsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setDefaults, request, options: options);
  }

  /// ==================== Connection Settings ====================
  $grpc.ResponseFuture<$0.Payload> setAutoConnect(
    $14.SetAutoconnectRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setAutoConnect, request, options: options);
  }

  $grpc.ResponseFuture<$14.SetProtocolResponse> setProtocol(
    $14.SetProtocolRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setProtocol, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setTechnology(
    $14.SetTechnologyRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setTechnology, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setObfuscate(
    $14.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setObfuscate, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setPostQuantum(
    $14.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setPostQuantum, request, options: options);
  }

  $grpc.ResponseFuture<$15.RecentConnectionsResponse> getRecentConnections(
    $15.RecentConnectionsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getRecentConnections, request, options: options);
  }

  /// ==================== Network Settings ====================
  $grpc.ResponseFuture<$14.SetDNSResponse> setDNS(
    $14.SetDNSRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setDNS, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setFirewall(
    $14.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setFirewall, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setFirewallMark(
    $14.SetUint32Request request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setFirewallMark, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setRouting(
    $14.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setRouting, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setKillSwitch(
    $14.SetKillSwitchRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setKillSwitch, request, options: options);
  }

  $grpc.ResponseFuture<$14.SetLANDiscoveryResponse> setLANDiscovery(
    $14.SetLANDiscoveryRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setLANDiscovery, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setVirtualLocation(
    $14.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setVirtualLocation, request, options: options);
  }

  /// ==================== UI Settings ====================
  $grpc.ResponseFuture<$0.Payload> setNotify(
    $14.SetNotifyRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setNotify, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> setTray(
    $14.SetTrayRequest request, {
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

  $grpc.ResponseFuture<$16.FeatureToggles> getFeatureToggles(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getFeatureToggles, request, options: options);
  }

  /// ==================== Allowlist Management ====================
  $grpc.ResponseFuture<$0.Payload> setAllowlist(
    $14.SetAllowlistRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setAllowlist, request, options: options);
  }

  $grpc.ResponseFuture<$0.Payload> unsetAllowlist(
    $14.SetAllowlistRequest request, {
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
    $14.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setAnalytics, request, options: options);
  }

  $grpc.ResponseFuture<$14.SetThreatProtectionLiteResponse>
      setThreatProtectionLite(
    $14.SetThreatProtectionLiteRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setThreatProtectionLite, request,
        options: options);
  }

  /// ==================== System & Monitoring ====================
  $grpc.ResponseFuture<$17.PingResponse> ping(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$ping, request, options: options);
  }

  $grpc.ResponseStream<$18.AppState> subscribeToStateChanges(
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
  static final _$getServers = $grpc.ClientMethod<$0.Empty, $10.ServersResponse>(
      '/pb.Daemon/GetServers',
      ($0.Empty value) => value.writeToBuffer(),
      $10.ServersResponse.fromBuffer);
  static final _$countries = $grpc.ClientMethod<$0.Empty, $0.ServerGroupsList>(
      '/pb.Daemon/Countries',
      ($0.Empty value) => value.writeToBuffer(),
      $0.ServerGroupsList.fromBuffer);
  static final _$cities =
      $grpc.ClientMethod<$11.CitiesRequest, $0.ServerGroupsList>(
          '/pb.Daemon/Cities',
          ($11.CitiesRequest value) => value.writeToBuffer(),
          $0.ServerGroupsList.fromBuffer);
  static final _$groups = $grpc.ClientMethod<$0.Empty, $0.ServerGroupsList>(
      '/pb.Daemon/Groups',
      ($0.Empty value) => value.writeToBuffer(),
      $0.ServerGroupsList.fromBuffer);
  static final _$settings = $grpc.ClientMethod<$0.Empty, $12.SettingsResponse>(
      '/pb.Daemon/Settings',
      ($0.Empty value) => value.writeToBuffer(),
      $12.SettingsResponse.fromBuffer);
  static final _$setDefaults =
      $grpc.ClientMethod<$13.SetDefaultsRequest, $0.Payload>(
          '/pb.Daemon/SetDefaults',
          ($13.SetDefaultsRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setAutoConnect =
      $grpc.ClientMethod<$14.SetAutoconnectRequest, $0.Payload>(
          '/pb.Daemon/SetAutoConnect',
          ($14.SetAutoconnectRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setProtocol =
      $grpc.ClientMethod<$14.SetProtocolRequest, $14.SetProtocolResponse>(
          '/pb.Daemon/SetProtocol',
          ($14.SetProtocolRequest value) => value.writeToBuffer(),
          $14.SetProtocolResponse.fromBuffer);
  static final _$setTechnology =
      $grpc.ClientMethod<$14.SetTechnologyRequest, $0.Payload>(
          '/pb.Daemon/SetTechnology',
          ($14.SetTechnologyRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setObfuscate =
      $grpc.ClientMethod<$14.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetObfuscate',
          ($14.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setPostQuantum =
      $grpc.ClientMethod<$14.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetPostQuantum',
          ($14.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$getRecentConnections = $grpc.ClientMethod<
          $15.RecentConnectionsRequest, $15.RecentConnectionsResponse>(
      '/pb.Daemon/GetRecentConnections',
      ($15.RecentConnectionsRequest value) => value.writeToBuffer(),
      $15.RecentConnectionsResponse.fromBuffer);
  static final _$setDNS =
      $grpc.ClientMethod<$14.SetDNSRequest, $14.SetDNSResponse>(
          '/pb.Daemon/SetDNS',
          ($14.SetDNSRequest value) => value.writeToBuffer(),
          $14.SetDNSResponse.fromBuffer);
  static final _$setFirewall =
      $grpc.ClientMethod<$14.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetFirewall',
          ($14.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setFirewallMark =
      $grpc.ClientMethod<$14.SetUint32Request, $0.Payload>(
          '/pb.Daemon/SetFirewallMark',
          ($14.SetUint32Request value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setRouting =
      $grpc.ClientMethod<$14.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetRouting',
          ($14.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setKillSwitch =
      $grpc.ClientMethod<$14.SetKillSwitchRequest, $0.Payload>(
          '/pb.Daemon/SetKillSwitch',
          ($14.SetKillSwitchRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setLANDiscovery = $grpc.ClientMethod<
          $14.SetLANDiscoveryRequest, $14.SetLANDiscoveryResponse>(
      '/pb.Daemon/SetLANDiscovery',
      ($14.SetLANDiscoveryRequest value) => value.writeToBuffer(),
      $14.SetLANDiscoveryResponse.fromBuffer);
  static final _$setVirtualLocation =
      $grpc.ClientMethod<$14.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetVirtualLocation',
          ($14.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setNotify =
      $grpc.ClientMethod<$14.SetNotifyRequest, $0.Payload>(
          '/pb.Daemon/SetNotify',
          ($14.SetNotifyRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setTray = $grpc.ClientMethod<$14.SetTrayRequest, $0.Payload>(
      '/pb.Daemon/SetTray',
      ($14.SetTrayRequest value) => value.writeToBuffer(),
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
      $grpc.ClientMethod<$0.Empty, $16.FeatureToggles>(
          '/pb.Daemon/GetFeatureToggles',
          ($0.Empty value) => value.writeToBuffer(),
          $16.FeatureToggles.fromBuffer);
  static final _$setAllowlist =
      $grpc.ClientMethod<$14.SetAllowlistRequest, $0.Payload>(
          '/pb.Daemon/SetAllowlist',
          ($14.SetAllowlistRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$unsetAllowlist =
      $grpc.ClientMethod<$14.SetAllowlistRequest, $0.Payload>(
          '/pb.Daemon/UnsetAllowlist',
          ($14.SetAllowlistRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$unsetAllAllowlist = $grpc.ClientMethod<$0.Empty, $0.Payload>(
      '/pb.Daemon/UnsetAllAllowlist',
      ($0.Empty value) => value.writeToBuffer(),
      $0.Payload.fromBuffer);
  static final _$setAnalytics =
      $grpc.ClientMethod<$14.SetGenericRequest, $0.Payload>(
          '/pb.Daemon/SetAnalytics',
          ($14.SetGenericRequest value) => value.writeToBuffer(),
          $0.Payload.fromBuffer);
  static final _$setThreatProtectionLite = $grpc.ClientMethod<
          $14.SetThreatProtectionLiteRequest,
          $14.SetThreatProtectionLiteResponse>(
      '/pb.Daemon/SetThreatProtectionLite',
      ($14.SetThreatProtectionLiteRequest value) => value.writeToBuffer(),
      $14.SetThreatProtectionLiteResponse.fromBuffer);
  static final _$ping = $grpc.ClientMethod<$0.Empty, $17.PingResponse>(
      '/pb.Daemon/Ping',
      ($0.Empty value) => value.writeToBuffer(),
      $17.PingResponse.fromBuffer);
  static final _$subscribeToStateChanges =
      $grpc.ClientMethod<$0.Empty, $18.AppState>(
          '/pb.Daemon/SubscribeToStateChanges',
          ($0.Empty value) => value.writeToBuffer(),
          $18.AppState.fromBuffer);
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
    $addMethod($grpc.ServiceMethod<$0.Empty, $10.ServersResponse>(
        'GetServers',
        getServers_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($10.ServersResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.ServerGroupsList>(
        'Countries',
        countries_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.ServerGroupsList value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$11.CitiesRequest, $0.ServerGroupsList>(
        'Cities',
        cities_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $11.CitiesRequest.fromBuffer(value),
        ($0.ServerGroupsList value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.ServerGroupsList>(
        'Groups',
        groups_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.ServerGroupsList value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $12.SettingsResponse>(
        'Settings',
        settings_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($12.SettingsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$13.SetDefaultsRequest, $0.Payload>(
        'SetDefaults',
        setDefaults_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $13.SetDefaultsRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetAutoconnectRequest, $0.Payload>(
        'SetAutoConnect',
        setAutoConnect_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetAutoconnectRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$14.SetProtocolRequest, $14.SetProtocolResponse>(
            'SetProtocol',
            setProtocol_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $14.SetProtocolRequest.fromBuffer(value),
            ($14.SetProtocolResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetTechnologyRequest, $0.Payload>(
        'SetTechnology',
        setTechnology_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetTechnologyRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetGenericRequest, $0.Payload>(
        'SetObfuscate',
        setObfuscate_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetGenericRequest, $0.Payload>(
        'SetPostQuantum',
        setPostQuantum_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$15.RecentConnectionsRequest,
            $15.RecentConnectionsResponse>(
        'GetRecentConnections',
        getRecentConnections_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $15.RecentConnectionsRequest.fromBuffer(value),
        ($15.RecentConnectionsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetDNSRequest, $14.SetDNSResponse>(
        'SetDNS',
        setDNS_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $14.SetDNSRequest.fromBuffer(value),
        ($14.SetDNSResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetGenericRequest, $0.Payload>(
        'SetFirewall',
        setFirewall_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetUint32Request, $0.Payload>(
        'SetFirewallMark',
        setFirewallMark_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $14.SetUint32Request.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetGenericRequest, $0.Payload>(
        'SetRouting',
        setRouting_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetKillSwitchRequest, $0.Payload>(
        'SetKillSwitch',
        setKillSwitch_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetKillSwitchRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetLANDiscoveryRequest,
            $14.SetLANDiscoveryResponse>(
        'SetLANDiscovery',
        setLANDiscovery_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetLANDiscoveryRequest.fromBuffer(value),
        ($14.SetLANDiscoveryResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetGenericRequest, $0.Payload>(
        'SetVirtualLocation',
        setVirtualLocation_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetNotifyRequest, $0.Payload>(
        'SetNotify',
        setNotify_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $14.SetNotifyRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetTrayRequest, $0.Payload>(
        'SetTray',
        setTray_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $14.SetTrayRequest.fromBuffer(value),
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
    $addMethod($grpc.ServiceMethod<$0.Empty, $16.FeatureToggles>(
        'GetFeatureToggles',
        getFeatureToggles_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($16.FeatureToggles value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetAllowlistRequest, $0.Payload>(
        'SetAllowlist',
        setAllowlist_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetAllowlistRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetAllowlistRequest, $0.Payload>(
        'UnsetAllowlist',
        unsetAllowlist_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetAllowlistRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.Payload>(
        'UnsetAllAllowlist',
        unsetAllAllowlist_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetGenericRequest, $0.Payload>(
        'SetAnalytics',
        setAnalytics_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetGenericRequest.fromBuffer(value),
        ($0.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$14.SetThreatProtectionLiteRequest,
            $14.SetThreatProtectionLiteResponse>(
        'SetThreatProtectionLite',
        setThreatProtectionLite_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $14.SetThreatProtectionLiteRequest.fromBuffer(value),
        ($14.SetThreatProtectionLiteResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $17.PingResponse>(
        'Ping',
        ping_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($17.PingResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $18.AppState>(
        'SubscribeToStateChanges',
        subscribeToStateChanges_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($18.AppState value) => value.writeToBuffer()));
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

  $async.Future<$10.ServersResponse> getServers_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return getServers($call, await $request);
  }

  $async.Future<$10.ServersResponse> getServers(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.ServerGroupsList> countries_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return countries($call, await $request);
  }

  $async.Future<$0.ServerGroupsList> countries(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.ServerGroupsList> cities_Pre($grpc.ServiceCall $call,
      $async.Future<$11.CitiesRequest> $request) async {
    return cities($call, await $request);
  }

  $async.Future<$0.ServerGroupsList> cities(
      $grpc.ServiceCall call, $11.CitiesRequest request);

  $async.Future<$0.ServerGroupsList> groups_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return groups($call, await $request);
  }

  $async.Future<$0.ServerGroupsList> groups(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$12.SettingsResponse> settings_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return settings($call, await $request);
  }

  $async.Future<$12.SettingsResponse> settings(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.Payload> setDefaults_Pre($grpc.ServiceCall $call,
      $async.Future<$13.SetDefaultsRequest> $request) async {
    return setDefaults($call, await $request);
  }

  $async.Future<$0.Payload> setDefaults(
      $grpc.ServiceCall call, $13.SetDefaultsRequest request);

  $async.Future<$0.Payload> setAutoConnect_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetAutoconnectRequest> $request) async {
    return setAutoConnect($call, await $request);
  }

  $async.Future<$0.Payload> setAutoConnect(
      $grpc.ServiceCall call, $14.SetAutoconnectRequest request);

  $async.Future<$14.SetProtocolResponse> setProtocol_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$14.SetProtocolRequest> $request) async {
    return setProtocol($call, await $request);
  }

  $async.Future<$14.SetProtocolResponse> setProtocol(
      $grpc.ServiceCall call, $14.SetProtocolRequest request);

  $async.Future<$0.Payload> setTechnology_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetTechnologyRequest> $request) async {
    return setTechnology($call, await $request);
  }

  $async.Future<$0.Payload> setTechnology(
      $grpc.ServiceCall call, $14.SetTechnologyRequest request);

  $async.Future<$0.Payload> setObfuscate_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetGenericRequest> $request) async {
    return setObfuscate($call, await $request);
  }

  $async.Future<$0.Payload> setObfuscate(
      $grpc.ServiceCall call, $14.SetGenericRequest request);

  $async.Future<$0.Payload> setPostQuantum_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetGenericRequest> $request) async {
    return setPostQuantum($call, await $request);
  }

  $async.Future<$0.Payload> setPostQuantum(
      $grpc.ServiceCall call, $14.SetGenericRequest request);

  $async.Future<$15.RecentConnectionsResponse> getRecentConnections_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$15.RecentConnectionsRequest> $request) async {
    return getRecentConnections($call, await $request);
  }

  $async.Future<$15.RecentConnectionsResponse> getRecentConnections(
      $grpc.ServiceCall call, $15.RecentConnectionsRequest request);

  $async.Future<$14.SetDNSResponse> setDNS_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetDNSRequest> $request) async {
    return setDNS($call, await $request);
  }

  $async.Future<$14.SetDNSResponse> setDNS(
      $grpc.ServiceCall call, $14.SetDNSRequest request);

  $async.Future<$0.Payload> setFirewall_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetGenericRequest> $request) async {
    return setFirewall($call, await $request);
  }

  $async.Future<$0.Payload> setFirewall(
      $grpc.ServiceCall call, $14.SetGenericRequest request);

  $async.Future<$0.Payload> setFirewallMark_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetUint32Request> $request) async {
    return setFirewallMark($call, await $request);
  }

  $async.Future<$0.Payload> setFirewallMark(
      $grpc.ServiceCall call, $14.SetUint32Request request);

  $async.Future<$0.Payload> setRouting_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetGenericRequest> $request) async {
    return setRouting($call, await $request);
  }

  $async.Future<$0.Payload> setRouting(
      $grpc.ServiceCall call, $14.SetGenericRequest request);

  $async.Future<$0.Payload> setKillSwitch_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetKillSwitchRequest> $request) async {
    return setKillSwitch($call, await $request);
  }

  $async.Future<$0.Payload> setKillSwitch(
      $grpc.ServiceCall call, $14.SetKillSwitchRequest request);

  $async.Future<$14.SetLANDiscoveryResponse> setLANDiscovery_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$14.SetLANDiscoveryRequest> $request) async {
    return setLANDiscovery($call, await $request);
  }

  $async.Future<$14.SetLANDiscoveryResponse> setLANDiscovery(
      $grpc.ServiceCall call, $14.SetLANDiscoveryRequest request);

  $async.Future<$0.Payload> setVirtualLocation_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetGenericRequest> $request) async {
    return setVirtualLocation($call, await $request);
  }

  $async.Future<$0.Payload> setVirtualLocation(
      $grpc.ServiceCall call, $14.SetGenericRequest request);

  $async.Future<$0.Payload> setNotify_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetNotifyRequest> $request) async {
    return setNotify($call, await $request);
  }

  $async.Future<$0.Payload> setNotify(
      $grpc.ServiceCall call, $14.SetNotifyRequest request);

  $async.Future<$0.Payload> setTray_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetTrayRequest> $request) async {
    return setTray($call, await $request);
  }

  $async.Future<$0.Payload> setTray(
      $grpc.ServiceCall call, $14.SetTrayRequest request);

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

  $async.Future<$16.FeatureToggles> getFeatureToggles_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return getFeatureToggles($call, await $request);
  }

  $async.Future<$16.FeatureToggles> getFeatureToggles(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.Payload> setAllowlist_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetAllowlistRequest> $request) async {
    return setAllowlist($call, await $request);
  }

  $async.Future<$0.Payload> setAllowlist(
      $grpc.ServiceCall call, $14.SetAllowlistRequest request);

  $async.Future<$0.Payload> unsetAllowlist_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetAllowlistRequest> $request) async {
    return unsetAllowlist($call, await $request);
  }

  $async.Future<$0.Payload> unsetAllowlist(
      $grpc.ServiceCall call, $14.SetAllowlistRequest request);

  $async.Future<$0.Payload> unsetAllAllowlist_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return unsetAllAllowlist($call, await $request);
  }

  $async.Future<$0.Payload> unsetAllAllowlist(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.Payload> setAnalytics_Pre($grpc.ServiceCall $call,
      $async.Future<$14.SetGenericRequest> $request) async {
    return setAnalytics($call, await $request);
  }

  $async.Future<$0.Payload> setAnalytics(
      $grpc.ServiceCall call, $14.SetGenericRequest request);

  $async.Future<$14.SetThreatProtectionLiteResponse>
      setThreatProtectionLite_Pre($grpc.ServiceCall $call,
          $async.Future<$14.SetThreatProtectionLiteRequest> $request) async {
    return setThreatProtectionLite($call, await $request);
  }

  $async.Future<$14.SetThreatProtectionLiteResponse> setThreatProtectionLite(
      $grpc.ServiceCall call, $14.SetThreatProtectionLiteRequest request);

  $async.Future<$17.PingResponse> ping_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return ping($call, await $request);
  }

  $async.Future<$17.PingResponse> ping(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Stream<$18.AppState> subscribeToStateChanges_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async* {
    yield* subscribeToStateChanges($call, await $request);
  }

  $async.Stream<$18.AppState> subscribeToStateChanges(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.GetDaemonApiVersionResponse> getDaemonApiVersion_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.GetDaemonApiVersionRequest> $request) async {
    return getDaemonApiVersion($call, await $request);
  }

  $async.Future<$0.GetDaemonApiVersionResponse> getDaemonApiVersion(
      $grpc.ServiceCall call, $0.GetDaemonApiVersionRequest request);
}
