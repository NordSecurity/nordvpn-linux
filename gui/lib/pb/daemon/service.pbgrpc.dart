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

import 'account.pb.dart' as $0;
import 'cities.pb.dart' as $3;
import 'common.pb.dart' as $1;
import 'connect.pb.dart' as $4;
import 'defaults.pb.dart' as $11;
import 'features.pb.dart' as $17;
import 'login.pb.dart' as $5;
import 'login_with_token.pb.dart' as $6;
import 'logout.pb.dart' as $7;
import 'ping.pb.dart' as $8;
import 'purchase.pb.dart' as $14;
import 'rate.pb.dart' as $9;
import 'servers.pb.dart' as $16;
import 'set.pb.dart' as $10;
import 'settings.pb.dart' as $12;
import 'state.pb.dart' as $15;
import 'status.pb.dart' as $13;
import 'token.pb.dart' as $2;

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

  $grpc.ResponseFuture<$0.AccountResponse> accountInfo(
    $0.AccountRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$accountInfo, request, options: options);
  }

  $grpc.ResponseFuture<$2.TokenInfoResponse> tokenInfo(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$tokenInfo, request, options: options);
  }

  $grpc.ResponseFuture<$1.ServerGroupsList> cities(
    $3.CitiesRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$cities, request, options: options);
  }

  $grpc.ResponseStream<$1.Payload> connect(
    $4.ConnectRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(
        _$connect, $async.Stream.fromIterable([request]),
        options: options);
  }

  $grpc.ResponseFuture<$1.Payload> connectCancel(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$connectCancel, request, options: options);
  }

  $grpc.ResponseFuture<$1.ServerGroupsList> countries(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$countries, request, options: options);
  }

  $grpc.ResponseStream<$1.Payload> disconnect(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(
        _$disconnect, $async.Stream.fromIterable([request]),
        options: options);
  }

  $grpc.ResponseFuture<$1.ServerGroupsList> groups(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$groups, request, options: options);
  }

  $grpc.ResponseFuture<$5.IsLoggedInResponse> isLoggedIn(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$isLoggedIn, request, options: options);
  }

  $grpc.ResponseFuture<$5.LoginResponse> loginWithToken(
    $6.LoginWithTokenRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$loginWithToken, request, options: options);
  }

  $grpc.ResponseFuture<$5.LoginOAuth2Response> loginOAuth2(
    $5.LoginOAuth2Request request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$loginOAuth2, request, options: options);
  }

  $grpc.ResponseFuture<$5.LoginOAuth2CallbackResponse> loginOAuth2Callback(
    $5.LoginOAuth2CallbackRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$loginOAuth2Callback, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> logout(
    $7.LogoutRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$logout, request, options: options);
  }

  $grpc.ResponseFuture<$8.PingResponse> ping(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$ping, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> rateConnection(
    $9.RateRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$rateConnection, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setAutoConnect(
    $10.SetAutoconnectRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setAutoConnect, request, options: options);
  }

  $grpc.ResponseFuture<$10.SetThreatProtectionLiteResponse>
      setThreatProtectionLite(
    $10.SetThreatProtectionLiteRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setThreatProtectionLite, request,
        options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setDefaults(
    $11.SetDefaultsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setDefaults, request, options: options);
  }

  $grpc.ResponseFuture<$10.SetDNSResponse> setDNS(
    $10.SetDNSRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setDNS, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setFirewall(
    $10.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setFirewall, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setFirewallMark(
    $10.SetUint32Request request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setFirewallMark, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setRouting(
    $10.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setRouting, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setAnalytics(
    $10.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setAnalytics, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setKillSwitch(
    $10.SetKillSwitchRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setKillSwitch, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setNotify(
    $10.SetNotifyRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setNotify, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setTray(
    $10.SetTrayRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setTray, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setObfuscate(
    $10.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setObfuscate, request, options: options);
  }

  $grpc.ResponseFuture<$10.SetProtocolResponse> setProtocol(
    $10.SetProtocolRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setProtocol, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setTechnology(
    $10.SetTechnologyRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setTechnology, request, options: options);
  }

  $grpc.ResponseFuture<$10.SetLANDiscoveryResponse> setLANDiscovery(
    $10.SetLANDiscoveryRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setLANDiscovery, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setAllowlist(
    $10.SetAllowlistRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setAllowlist, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> unsetAllowlist(
    $10.SetAllowlistRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$unsetAllowlist, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> unsetAllAllowlist(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$unsetAllAllowlist, request, options: options);
  }

  $grpc.ResponseFuture<$12.SettingsResponse> settings(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$settings, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> settingsProtocols(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$settingsProtocols, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> settingsTechnologies(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$settingsTechnologies, request, options: options);
  }

  $grpc.ResponseFuture<$13.StatusResponse> status(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$status, request, options: options);
  }

  $grpc.ResponseFuture<$14.ClaimOnlinePurchaseResponse> claimOnlinePurchase(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$claimOnlinePurchase, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setVirtualLocation(
    $10.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setVirtualLocation, request, options: options);
  }

  $grpc.ResponseStream<$15.AppState> subscribeToStateChanges(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(
        _$subscribeToStateChanges, $async.Stream.fromIterable([request]),
        options: options);
  }

  $grpc.ResponseFuture<$16.ServersResponse> getServers(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getServers, request, options: options);
  }

  $grpc.ResponseFuture<$1.Payload> setPostQuantum(
    $10.SetGenericRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setPostQuantum, request, options: options);
  }

  $grpc.ResponseFuture<$1.GetDaemonApiVersionResponse> getDaemonApiVersion(
    $1.GetDaemonApiVersionRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getDaemonApiVersion, request, options: options);
  }

  $grpc.ResponseFuture<$17.FeatureToggles> getFeatureToggles(
    $1.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getFeatureToggles, request, options: options);
  }

  // method descriptors

  static final _$accountInfo =
      $grpc.ClientMethod<$0.AccountRequest, $0.AccountResponse>(
          '/pb.Daemon/AccountInfo',
          ($0.AccountRequest value) => value.writeToBuffer(),
          $0.AccountResponse.fromBuffer);
  static final _$tokenInfo = $grpc.ClientMethod<$1.Empty, $2.TokenInfoResponse>(
      '/pb.Daemon/TokenInfo',
      ($1.Empty value) => value.writeToBuffer(),
      $2.TokenInfoResponse.fromBuffer);
  static final _$cities =
      $grpc.ClientMethod<$3.CitiesRequest, $1.ServerGroupsList>(
          '/pb.Daemon/Cities',
          ($3.CitiesRequest value) => value.writeToBuffer(),
          $1.ServerGroupsList.fromBuffer);
  static final _$connect = $grpc.ClientMethod<$4.ConnectRequest, $1.Payload>(
      '/pb.Daemon/Connect',
      ($4.ConnectRequest value) => value.writeToBuffer(),
      $1.Payload.fromBuffer);
  static final _$connectCancel = $grpc.ClientMethod<$1.Empty, $1.Payload>(
      '/pb.Daemon/ConnectCancel',
      ($1.Empty value) => value.writeToBuffer(),
      $1.Payload.fromBuffer);
  static final _$countries = $grpc.ClientMethod<$1.Empty, $1.ServerGroupsList>(
      '/pb.Daemon/Countries',
      ($1.Empty value) => value.writeToBuffer(),
      $1.ServerGroupsList.fromBuffer);
  static final _$disconnect = $grpc.ClientMethod<$1.Empty, $1.Payload>(
      '/pb.Daemon/Disconnect',
      ($1.Empty value) => value.writeToBuffer(),
      $1.Payload.fromBuffer);
  static final _$groups = $grpc.ClientMethod<$1.Empty, $1.ServerGroupsList>(
      '/pb.Daemon/Groups',
      ($1.Empty value) => value.writeToBuffer(),
      $1.ServerGroupsList.fromBuffer);
  static final _$isLoggedIn =
      $grpc.ClientMethod<$1.Empty, $5.IsLoggedInResponse>(
          '/pb.Daemon/IsLoggedIn',
          ($1.Empty value) => value.writeToBuffer(),
          $5.IsLoggedInResponse.fromBuffer);
  static final _$loginWithToken =
      $grpc.ClientMethod<$6.LoginWithTokenRequest, $5.LoginResponse>(
          '/pb.Daemon/LoginWithToken',
          ($6.LoginWithTokenRequest value) => value.writeToBuffer(),
          $5.LoginResponse.fromBuffer);
  static final _$loginOAuth2 =
      $grpc.ClientMethod<$5.LoginOAuth2Request, $5.LoginOAuth2Response>(
          '/pb.Daemon/LoginOAuth2',
          ($5.LoginOAuth2Request value) => value.writeToBuffer(),
          $5.LoginOAuth2Response.fromBuffer);
  static final _$loginOAuth2Callback = $grpc.ClientMethod<
          $5.LoginOAuth2CallbackRequest, $5.LoginOAuth2CallbackResponse>(
      '/pb.Daemon/LoginOAuth2Callback',
      ($5.LoginOAuth2CallbackRequest value) => value.writeToBuffer(),
      $5.LoginOAuth2CallbackResponse.fromBuffer);
  static final _$logout = $grpc.ClientMethod<$7.LogoutRequest, $1.Payload>(
      '/pb.Daemon/Logout',
      ($7.LogoutRequest value) => value.writeToBuffer(),
      $1.Payload.fromBuffer);
  static final _$ping = $grpc.ClientMethod<$1.Empty, $8.PingResponse>(
      '/pb.Daemon/Ping',
      ($1.Empty value) => value.writeToBuffer(),
      $8.PingResponse.fromBuffer);
  static final _$rateConnection =
      $grpc.ClientMethod<$9.RateRequest, $1.Payload>(
          '/pb.Daemon/RateConnection',
          ($9.RateRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$setAutoConnect =
      $grpc.ClientMethod<$10.SetAutoconnectRequest, $1.Payload>(
          '/pb.Daemon/SetAutoConnect',
          ($10.SetAutoconnectRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$setThreatProtectionLite = $grpc.ClientMethod<
          $10.SetThreatProtectionLiteRequest,
          $10.SetThreatProtectionLiteResponse>(
      '/pb.Daemon/SetThreatProtectionLite',
      ($10.SetThreatProtectionLiteRequest value) => value.writeToBuffer(),
      $10.SetThreatProtectionLiteResponse.fromBuffer);
  static final _$setDefaults =
      $grpc.ClientMethod<$11.SetDefaultsRequest, $1.Payload>(
          '/pb.Daemon/SetDefaults',
          ($11.SetDefaultsRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$setDNS =
      $grpc.ClientMethod<$10.SetDNSRequest, $10.SetDNSResponse>(
          '/pb.Daemon/SetDNS',
          ($10.SetDNSRequest value) => value.writeToBuffer(),
          $10.SetDNSResponse.fromBuffer);
  static final _$setFirewall =
      $grpc.ClientMethod<$10.SetGenericRequest, $1.Payload>(
          '/pb.Daemon/SetFirewall',
          ($10.SetGenericRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$setFirewallMark =
      $grpc.ClientMethod<$10.SetUint32Request, $1.Payload>(
          '/pb.Daemon/SetFirewallMark',
          ($10.SetUint32Request value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$setRouting =
      $grpc.ClientMethod<$10.SetGenericRequest, $1.Payload>(
          '/pb.Daemon/SetRouting',
          ($10.SetGenericRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$setAnalytics =
      $grpc.ClientMethod<$10.SetGenericRequest, $1.Payload>(
          '/pb.Daemon/SetAnalytics',
          ($10.SetGenericRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$setKillSwitch =
      $grpc.ClientMethod<$10.SetKillSwitchRequest, $1.Payload>(
          '/pb.Daemon/SetKillSwitch',
          ($10.SetKillSwitchRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$setNotify =
      $grpc.ClientMethod<$10.SetNotifyRequest, $1.Payload>(
          '/pb.Daemon/SetNotify',
          ($10.SetNotifyRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$setTray = $grpc.ClientMethod<$10.SetTrayRequest, $1.Payload>(
      '/pb.Daemon/SetTray',
      ($10.SetTrayRequest value) => value.writeToBuffer(),
      $1.Payload.fromBuffer);
  static final _$setObfuscate =
      $grpc.ClientMethod<$10.SetGenericRequest, $1.Payload>(
          '/pb.Daemon/SetObfuscate',
          ($10.SetGenericRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$setProtocol =
      $grpc.ClientMethod<$10.SetProtocolRequest, $10.SetProtocolResponse>(
          '/pb.Daemon/SetProtocol',
          ($10.SetProtocolRequest value) => value.writeToBuffer(),
          $10.SetProtocolResponse.fromBuffer);
  static final _$setTechnology =
      $grpc.ClientMethod<$10.SetTechnologyRequest, $1.Payload>(
          '/pb.Daemon/SetTechnology',
          ($10.SetTechnologyRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$setLANDiscovery = $grpc.ClientMethod<
          $10.SetLANDiscoveryRequest, $10.SetLANDiscoveryResponse>(
      '/pb.Daemon/SetLANDiscovery',
      ($10.SetLANDiscoveryRequest value) => value.writeToBuffer(),
      $10.SetLANDiscoveryResponse.fromBuffer);
  static final _$setAllowlist =
      $grpc.ClientMethod<$10.SetAllowlistRequest, $1.Payload>(
          '/pb.Daemon/SetAllowlist',
          ($10.SetAllowlistRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$unsetAllowlist =
      $grpc.ClientMethod<$10.SetAllowlistRequest, $1.Payload>(
          '/pb.Daemon/UnsetAllowlist',
          ($10.SetAllowlistRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$unsetAllAllowlist = $grpc.ClientMethod<$1.Empty, $1.Payload>(
      '/pb.Daemon/UnsetAllAllowlist',
      ($1.Empty value) => value.writeToBuffer(),
      $1.Payload.fromBuffer);
  static final _$settings = $grpc.ClientMethod<$1.Empty, $12.SettingsResponse>(
      '/pb.Daemon/Settings',
      ($1.Empty value) => value.writeToBuffer(),
      $12.SettingsResponse.fromBuffer);
  static final _$settingsProtocols = $grpc.ClientMethod<$1.Empty, $1.Payload>(
      '/pb.Daemon/SettingsProtocols',
      ($1.Empty value) => value.writeToBuffer(),
      $1.Payload.fromBuffer);
  static final _$settingsTechnologies =
      $grpc.ClientMethod<$1.Empty, $1.Payload>(
          '/pb.Daemon/SettingsTechnologies',
          ($1.Empty value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$status = $grpc.ClientMethod<$1.Empty, $13.StatusResponse>(
      '/pb.Daemon/Status',
      ($1.Empty value) => value.writeToBuffer(),
      $13.StatusResponse.fromBuffer);
  static final _$claimOnlinePurchase =
      $grpc.ClientMethod<$1.Empty, $14.ClaimOnlinePurchaseResponse>(
          '/pb.Daemon/ClaimOnlinePurchase',
          ($1.Empty value) => value.writeToBuffer(),
          $14.ClaimOnlinePurchaseResponse.fromBuffer);
  static final _$setVirtualLocation =
      $grpc.ClientMethod<$10.SetGenericRequest, $1.Payload>(
          '/pb.Daemon/SetVirtualLocation',
          ($10.SetGenericRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$subscribeToStateChanges =
      $grpc.ClientMethod<$1.Empty, $15.AppState>(
          '/pb.Daemon/SubscribeToStateChanges',
          ($1.Empty value) => value.writeToBuffer(),
          $15.AppState.fromBuffer);
  static final _$getServers = $grpc.ClientMethod<$1.Empty, $16.ServersResponse>(
      '/pb.Daemon/GetServers',
      ($1.Empty value) => value.writeToBuffer(),
      $16.ServersResponse.fromBuffer);
  static final _$setPostQuantum =
      $grpc.ClientMethod<$10.SetGenericRequest, $1.Payload>(
          '/pb.Daemon/SetPostQuantum',
          ($10.SetGenericRequest value) => value.writeToBuffer(),
          $1.Payload.fromBuffer);
  static final _$getDaemonApiVersion = $grpc.ClientMethod<
          $1.GetDaemonApiVersionRequest, $1.GetDaemonApiVersionResponse>(
      '/pb.Daemon/GetDaemonApiVersion',
      ($1.GetDaemonApiVersionRequest value) => value.writeToBuffer(),
      $1.GetDaemonApiVersionResponse.fromBuffer);
  static final _$getFeatureToggles =
      $grpc.ClientMethod<$1.Empty, $17.FeatureToggles>(
          '/pb.Daemon/GetFeatureToggles',
          ($1.Empty value) => value.writeToBuffer(),
          $17.FeatureToggles.fromBuffer);
}

@$pb.GrpcServiceName('pb.Daemon')
abstract class DaemonServiceBase extends $grpc.Service {
  $core.String get $name => 'pb.Daemon';

  DaemonServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.AccountRequest, $0.AccountResponse>(
        'AccountInfo',
        accountInfo_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.AccountRequest.fromBuffer(value),
        ($0.AccountResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $2.TokenInfoResponse>(
        'TokenInfo',
        tokenInfo_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($2.TokenInfoResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.CitiesRequest, $1.ServerGroupsList>(
        'Cities',
        cities_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.CitiesRequest.fromBuffer(value),
        ($1.ServerGroupsList value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$4.ConnectRequest, $1.Payload>(
        'Connect',
        connect_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $4.ConnectRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $1.Payload>(
        'ConnectCancel',
        connectCancel_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $1.ServerGroupsList>(
        'Countries',
        countries_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($1.ServerGroupsList value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $1.Payload>(
        'Disconnect',
        disconnect_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $1.ServerGroupsList>(
        'Groups',
        groups_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($1.ServerGroupsList value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $5.IsLoggedInResponse>(
        'IsLoggedIn',
        isLoggedIn_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($5.IsLoggedInResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$6.LoginWithTokenRequest, $5.LoginResponse>(
        'LoginWithToken',
        loginWithToken_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $6.LoginWithTokenRequest.fromBuffer(value),
        ($5.LoginResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$5.LoginOAuth2Request, $5.LoginOAuth2Response>(
            'LoginOAuth2',
            loginOAuth2_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $5.LoginOAuth2Request.fromBuffer(value),
            ($5.LoginOAuth2Response value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$5.LoginOAuth2CallbackRequest,
            $5.LoginOAuth2CallbackResponse>(
        'LoginOAuth2Callback',
        loginOAuth2Callback_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $5.LoginOAuth2CallbackRequest.fromBuffer(value),
        ($5.LoginOAuth2CallbackResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$7.LogoutRequest, $1.Payload>(
        'Logout',
        logout_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $7.LogoutRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $8.PingResponse>(
        'Ping',
        ping_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($8.PingResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$9.RateRequest, $1.Payload>(
        'RateConnection',
        rateConnection_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $9.RateRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetAutoconnectRequest, $1.Payload>(
        'SetAutoConnect',
        setAutoConnect_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetAutoconnectRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetThreatProtectionLiteRequest,
            $10.SetThreatProtectionLiteResponse>(
        'SetThreatProtectionLite',
        setThreatProtectionLite_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetThreatProtectionLiteRequest.fromBuffer(value),
        ($10.SetThreatProtectionLiteResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$11.SetDefaultsRequest, $1.Payload>(
        'SetDefaults',
        setDefaults_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $11.SetDefaultsRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetDNSRequest, $10.SetDNSResponse>(
        'SetDNS',
        setDNS_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $10.SetDNSRequest.fromBuffer(value),
        ($10.SetDNSResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetGenericRequest, $1.Payload>(
        'SetFirewall',
        setFirewall_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetGenericRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetUint32Request, $1.Payload>(
        'SetFirewallMark',
        setFirewallMark_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $10.SetUint32Request.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetGenericRequest, $1.Payload>(
        'SetRouting',
        setRouting_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetGenericRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetGenericRequest, $1.Payload>(
        'SetAnalytics',
        setAnalytics_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetGenericRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetKillSwitchRequest, $1.Payload>(
        'SetKillSwitch',
        setKillSwitch_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetKillSwitchRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetNotifyRequest, $1.Payload>(
        'SetNotify',
        setNotify_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $10.SetNotifyRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetTrayRequest, $1.Payload>(
        'SetTray',
        setTray_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $10.SetTrayRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetGenericRequest, $1.Payload>(
        'SetObfuscate',
        setObfuscate_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetGenericRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$10.SetProtocolRequest, $10.SetProtocolResponse>(
            'SetProtocol',
            setProtocol_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $10.SetProtocolRequest.fromBuffer(value),
            ($10.SetProtocolResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetTechnologyRequest, $1.Payload>(
        'SetTechnology',
        setTechnology_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetTechnologyRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetLANDiscoveryRequest,
            $10.SetLANDiscoveryResponse>(
        'SetLANDiscovery',
        setLANDiscovery_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetLANDiscoveryRequest.fromBuffer(value),
        ($10.SetLANDiscoveryResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetAllowlistRequest, $1.Payload>(
        'SetAllowlist',
        setAllowlist_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetAllowlistRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetAllowlistRequest, $1.Payload>(
        'UnsetAllowlist',
        unsetAllowlist_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetAllowlistRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $1.Payload>(
        'UnsetAllAllowlist',
        unsetAllAllowlist_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $12.SettingsResponse>(
        'Settings',
        settings_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($12.SettingsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $1.Payload>(
        'SettingsProtocols',
        settingsProtocols_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $1.Payload>(
        'SettingsTechnologies',
        settingsTechnologies_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $13.StatusResponse>(
        'Status',
        status_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($13.StatusResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $14.ClaimOnlinePurchaseResponse>(
        'ClaimOnlinePurchase',
        claimOnlinePurchase_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($14.ClaimOnlinePurchaseResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetGenericRequest, $1.Payload>(
        'SetVirtualLocation',
        setVirtualLocation_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetGenericRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $15.AppState>(
        'SubscribeToStateChanges',
        subscribeToStateChanges_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($15.AppState value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $16.ServersResponse>(
        'GetServers',
        getServers_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($16.ServersResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$10.SetGenericRequest, $1.Payload>(
        'SetPostQuantum',
        setPostQuantum_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $10.SetGenericRequest.fromBuffer(value),
        ($1.Payload value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.GetDaemonApiVersionRequest,
            $1.GetDaemonApiVersionResponse>(
        'GetDaemonApiVersion',
        getDaemonApiVersion_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.GetDaemonApiVersionRequest.fromBuffer(value),
        ($1.GetDaemonApiVersionResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $17.FeatureToggles>(
        'GetFeatureToggles',
        getFeatureToggles_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($17.FeatureToggles value) => value.writeToBuffer()));
  }

  $async.Future<$0.AccountResponse> accountInfo_Pre($grpc.ServiceCall $call,
      $async.Future<$0.AccountRequest> $request) async {
    return accountInfo($call, await $request);
  }

  $async.Future<$0.AccountResponse> accountInfo(
      $grpc.ServiceCall call, $0.AccountRequest request);

  $async.Future<$2.TokenInfoResponse> tokenInfo_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return tokenInfo($call, await $request);
  }

  $async.Future<$2.TokenInfoResponse> tokenInfo(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$1.ServerGroupsList> cities_Pre(
      $grpc.ServiceCall $call, $async.Future<$3.CitiesRequest> $request) async {
    return cities($call, await $request);
  }

  $async.Future<$1.ServerGroupsList> cities(
      $grpc.ServiceCall call, $3.CitiesRequest request);

  $async.Stream<$1.Payload> connect_Pre($grpc.ServiceCall $call,
      $async.Future<$4.ConnectRequest> $request) async* {
    yield* connect($call, await $request);
  }

  $async.Stream<$1.Payload> connect(
      $grpc.ServiceCall call, $4.ConnectRequest request);

  $async.Future<$1.Payload> connectCancel_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return connectCancel($call, await $request);
  }

  $async.Future<$1.Payload> connectCancel(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$1.ServerGroupsList> countries_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return countries($call, await $request);
  }

  $async.Future<$1.ServerGroupsList> countries(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Stream<$1.Payload> disconnect_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async* {
    yield* disconnect($call, await $request);
  }

  $async.Stream<$1.Payload> disconnect(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$1.ServerGroupsList> groups_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return groups($call, await $request);
  }

  $async.Future<$1.ServerGroupsList> groups(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$5.IsLoggedInResponse> isLoggedIn_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return isLoggedIn($call, await $request);
  }

  $async.Future<$5.IsLoggedInResponse> isLoggedIn(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$5.LoginResponse> loginWithToken_Pre($grpc.ServiceCall $call,
      $async.Future<$6.LoginWithTokenRequest> $request) async {
    return loginWithToken($call, await $request);
  }

  $async.Future<$5.LoginResponse> loginWithToken(
      $grpc.ServiceCall call, $6.LoginWithTokenRequest request);

  $async.Future<$5.LoginOAuth2Response> loginOAuth2_Pre($grpc.ServiceCall $call,
      $async.Future<$5.LoginOAuth2Request> $request) async {
    return loginOAuth2($call, await $request);
  }

  $async.Future<$5.LoginOAuth2Response> loginOAuth2(
      $grpc.ServiceCall call, $5.LoginOAuth2Request request);

  $async.Future<$5.LoginOAuth2CallbackResponse> loginOAuth2Callback_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$5.LoginOAuth2CallbackRequest> $request) async {
    return loginOAuth2Callback($call, await $request);
  }

  $async.Future<$5.LoginOAuth2CallbackResponse> loginOAuth2Callback(
      $grpc.ServiceCall call, $5.LoginOAuth2CallbackRequest request);

  $async.Future<$1.Payload> logout_Pre(
      $grpc.ServiceCall $call, $async.Future<$7.LogoutRequest> $request) async {
    return logout($call, await $request);
  }

  $async.Future<$1.Payload> logout(
      $grpc.ServiceCall call, $7.LogoutRequest request);

  $async.Future<$8.PingResponse> ping_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return ping($call, await $request);
  }

  $async.Future<$8.PingResponse> ping($grpc.ServiceCall call, $1.Empty request);

  $async.Future<$1.Payload> rateConnection_Pre(
      $grpc.ServiceCall $call, $async.Future<$9.RateRequest> $request) async {
    return rateConnection($call, await $request);
  }

  $async.Future<$1.Payload> rateConnection(
      $grpc.ServiceCall call, $9.RateRequest request);

  $async.Future<$1.Payload> setAutoConnect_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetAutoconnectRequest> $request) async {
    return setAutoConnect($call, await $request);
  }

  $async.Future<$1.Payload> setAutoConnect(
      $grpc.ServiceCall call, $10.SetAutoconnectRequest request);

  $async.Future<$10.SetThreatProtectionLiteResponse>
      setThreatProtectionLite_Pre($grpc.ServiceCall $call,
          $async.Future<$10.SetThreatProtectionLiteRequest> $request) async {
    return setThreatProtectionLite($call, await $request);
  }

  $async.Future<$10.SetThreatProtectionLiteResponse> setThreatProtectionLite(
      $grpc.ServiceCall call, $10.SetThreatProtectionLiteRequest request);

  $async.Future<$1.Payload> setDefaults_Pre($grpc.ServiceCall $call,
      $async.Future<$11.SetDefaultsRequest> $request) async {
    return setDefaults($call, await $request);
  }

  $async.Future<$1.Payload> setDefaults(
      $grpc.ServiceCall call, $11.SetDefaultsRequest request);

  $async.Future<$10.SetDNSResponse> setDNS_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetDNSRequest> $request) async {
    return setDNS($call, await $request);
  }

  $async.Future<$10.SetDNSResponse> setDNS(
      $grpc.ServiceCall call, $10.SetDNSRequest request);

  $async.Future<$1.Payload> setFirewall_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetGenericRequest> $request) async {
    return setFirewall($call, await $request);
  }

  $async.Future<$1.Payload> setFirewall(
      $grpc.ServiceCall call, $10.SetGenericRequest request);

  $async.Future<$1.Payload> setFirewallMark_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetUint32Request> $request) async {
    return setFirewallMark($call, await $request);
  }

  $async.Future<$1.Payload> setFirewallMark(
      $grpc.ServiceCall call, $10.SetUint32Request request);

  $async.Future<$1.Payload> setRouting_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetGenericRequest> $request) async {
    return setRouting($call, await $request);
  }

  $async.Future<$1.Payload> setRouting(
      $grpc.ServiceCall call, $10.SetGenericRequest request);

  $async.Future<$1.Payload> setAnalytics_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetGenericRequest> $request) async {
    return setAnalytics($call, await $request);
  }

  $async.Future<$1.Payload> setAnalytics(
      $grpc.ServiceCall call, $10.SetGenericRequest request);

  $async.Future<$1.Payload> setKillSwitch_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetKillSwitchRequest> $request) async {
    return setKillSwitch($call, await $request);
  }

  $async.Future<$1.Payload> setKillSwitch(
      $grpc.ServiceCall call, $10.SetKillSwitchRequest request);

  $async.Future<$1.Payload> setNotify_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetNotifyRequest> $request) async {
    return setNotify($call, await $request);
  }

  $async.Future<$1.Payload> setNotify(
      $grpc.ServiceCall call, $10.SetNotifyRequest request);

  $async.Future<$1.Payload> setTray_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetTrayRequest> $request) async {
    return setTray($call, await $request);
  }

  $async.Future<$1.Payload> setTray(
      $grpc.ServiceCall call, $10.SetTrayRequest request);

  $async.Future<$1.Payload> setObfuscate_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetGenericRequest> $request) async {
    return setObfuscate($call, await $request);
  }

  $async.Future<$1.Payload> setObfuscate(
      $grpc.ServiceCall call, $10.SetGenericRequest request);

  $async.Future<$10.SetProtocolResponse> setProtocol_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$10.SetProtocolRequest> $request) async {
    return setProtocol($call, await $request);
  }

  $async.Future<$10.SetProtocolResponse> setProtocol(
      $grpc.ServiceCall call, $10.SetProtocolRequest request);

  $async.Future<$1.Payload> setTechnology_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetTechnologyRequest> $request) async {
    return setTechnology($call, await $request);
  }

  $async.Future<$1.Payload> setTechnology(
      $grpc.ServiceCall call, $10.SetTechnologyRequest request);

  $async.Future<$10.SetLANDiscoveryResponse> setLANDiscovery_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$10.SetLANDiscoveryRequest> $request) async {
    return setLANDiscovery($call, await $request);
  }

  $async.Future<$10.SetLANDiscoveryResponse> setLANDiscovery(
      $grpc.ServiceCall call, $10.SetLANDiscoveryRequest request);

  $async.Future<$1.Payload> setAllowlist_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetAllowlistRequest> $request) async {
    return setAllowlist($call, await $request);
  }

  $async.Future<$1.Payload> setAllowlist(
      $grpc.ServiceCall call, $10.SetAllowlistRequest request);

  $async.Future<$1.Payload> unsetAllowlist_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetAllowlistRequest> $request) async {
    return unsetAllowlist($call, await $request);
  }

  $async.Future<$1.Payload> unsetAllowlist(
      $grpc.ServiceCall call, $10.SetAllowlistRequest request);

  $async.Future<$1.Payload> unsetAllAllowlist_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return unsetAllAllowlist($call, await $request);
  }

  $async.Future<$1.Payload> unsetAllAllowlist(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$12.SettingsResponse> settings_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return settings($call, await $request);
  }

  $async.Future<$12.SettingsResponse> settings(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$1.Payload> settingsProtocols_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return settingsProtocols($call, await $request);
  }

  $async.Future<$1.Payload> settingsProtocols(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$1.Payload> settingsTechnologies_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return settingsTechnologies($call, await $request);
  }

  $async.Future<$1.Payload> settingsTechnologies(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$13.StatusResponse> status_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return status($call, await $request);
  }

  $async.Future<$13.StatusResponse> status(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$14.ClaimOnlinePurchaseResponse> claimOnlinePurchase_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return claimOnlinePurchase($call, await $request);
  }

  $async.Future<$14.ClaimOnlinePurchaseResponse> claimOnlinePurchase(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$1.Payload> setVirtualLocation_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetGenericRequest> $request) async {
    return setVirtualLocation($call, await $request);
  }

  $async.Future<$1.Payload> setVirtualLocation(
      $grpc.ServiceCall call, $10.SetGenericRequest request);

  $async.Stream<$15.AppState> subscribeToStateChanges_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async* {
    yield* subscribeToStateChanges($call, await $request);
  }

  $async.Stream<$15.AppState> subscribeToStateChanges(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$16.ServersResponse> getServers_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return getServers($call, await $request);
  }

  $async.Future<$16.ServersResponse> getServers(
      $grpc.ServiceCall call, $1.Empty request);

  $async.Future<$1.Payload> setPostQuantum_Pre($grpc.ServiceCall $call,
      $async.Future<$10.SetGenericRequest> $request) async {
    return setPostQuantum($call, await $request);
  }

  $async.Future<$1.Payload> setPostQuantum(
      $grpc.ServiceCall call, $10.SetGenericRequest request);

  $async.Future<$1.GetDaemonApiVersionResponse> getDaemonApiVersion_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$1.GetDaemonApiVersionRequest> $request) async {
    return getDaemonApiVersion($call, await $request);
  }

  $async.Future<$1.GetDaemonApiVersionResponse> getDaemonApiVersion(
      $grpc.ServiceCall call, $1.GetDaemonApiVersionRequest request);

  $async.Future<$17.FeatureToggles> getFeatureToggles_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.Empty> $request) async {
    return getFeatureToggles($call, await $request);
  }

  $async.Future<$17.FeatureToggles> getFeatureToggles(
      $grpc.ServiceCall call, $1.Empty request);
}
