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

import 'empty.pb.dart' as $0;
import 'fsnotify.pb.dart' as $4;
import 'invite.pb.dart' as $2;
import 'peer.pb.dart' as $3;
import 'service_response.pb.dart' as $1;

export 'service.pb.dart';

/// Meshnet defines a service which handles the meshnet
/// functionality on a single device
@$pb.GrpcServiceName('meshpb.Meshnet')
class MeshnetClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  MeshnetClient(super.channel, {super.options, super.interceptors});

  /// EnableMeshnet enables the meshnet on this device
  $grpc.ResponseFuture<$1.MeshnetResponse> enableMeshnet(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$enableMeshnet, request, options: options);
  }

  /// IsEnabled retrieves whether meshnet is enabled
  $grpc.ResponseFuture<$1.IsEnabledResponse> isEnabled(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$isEnabled, request, options: options);
  }

  /// DisableMeshnet disables the meshnet on this device
  $grpc.ResponseFuture<$1.MeshnetResponse> disableMeshnet(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$disableMeshnet, request, options: options);
  }

  $grpc.ResponseFuture<$1.MeshnetResponse> refreshMeshnet(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$refreshMeshnet, request, options: options);
  }

  /// GetInvites retrieves a list of all the invites related to
  /// this device
  $grpc.ResponseFuture<$2.GetInvitesResponse> getInvites(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getInvites, request, options: options);
  }

  /// Invite sends the invite to the specified email to join the
  /// meshnet.
  $grpc.ResponseFuture<$2.InviteResponse> invite(
    $2.InviteRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$invite, request, options: options);
  }

  /// Invite sends the invite to the specified email to join the
  /// meshnet.
  $grpc.ResponseFuture<$2.RespondToInviteResponse> revokeInvite(
    $2.DenyInviteRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$revokeInvite, request, options: options);
  }

  /// AcceptInvite accepts the invite to join someone's meshnet
  $grpc.ResponseFuture<$2.RespondToInviteResponse> acceptInvite(
    $2.InviteRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$acceptInvite, request, options: options);
  }

  /// AcceptInvite denies the invite to join someone's meshnet
  $grpc.ResponseFuture<$2.RespondToInviteResponse> denyInvite(
    $2.DenyInviteRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$denyInvite, request, options: options);
  }

  /// GetPeers retries the list of all meshnet peers related to
  /// this device
  $grpc.ResponseFuture<$3.GetPeersResponse> getPeers(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getPeers, request, options: options);
  }

  /// RemovePeer removes a peer from the meshnet
  $grpc.ResponseFuture<$3.RemovePeerResponse> removePeer(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$removePeer, request, options: options);
  }

  /// ChangePeerNickname changes(set/remove) the nickname for a meshnet peer
  $grpc.ResponseFuture<$3.ChangeNicknameResponse> changePeerNickname(
    $3.ChangePeerNicknameRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$changePeerNickname, request, options: options);
  }

  /// ChangeMachineNickname changes the current machine meshnet nickname
  $grpc.ResponseFuture<$3.ChangeNicknameResponse> changeMachineNickname(
    $3.ChangeMachineNicknameRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$changeMachineNickname, request, options: options);
  }

  /// AllowRouting allows a peer to route traffic through this
  /// device
  $grpc.ResponseFuture<$3.AllowRoutingResponse> allowRouting(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$allowRouting, request, options: options);
  }

  /// DenyRouting allows a peer to route traffic through this
  /// device
  $grpc.ResponseFuture<$3.DenyRoutingResponse> denyRouting(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$denyRouting, request, options: options);
  }

  /// AllowIncoming allows a peer to send traffic to this device
  $grpc.ResponseFuture<$3.AllowIncomingResponse> allowIncoming(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$allowIncoming, request, options: options);
  }

  /// DenyIncoming denies a peer to send traffic to this device
  $grpc.ResponseFuture<$3.DenyIncomingResponse> denyIncoming(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$denyIncoming, request, options: options);
  }

  /// AllowLocalNetwork allows a peer to access local network when
  /// routing through this device
  $grpc.ResponseFuture<$3.AllowLocalNetworkResponse> allowLocalNetwork(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$allowLocalNetwork, request, options: options);
  }

  /// DenyLocalNetwork denies a peer to access local network when
  /// routing through this device
  $grpc.ResponseFuture<$3.DenyLocalNetworkResponse> denyLocalNetwork(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$denyLocalNetwork, request, options: options);
  }

  /// AllowFileshare allows peer to send files to this device
  $grpc.ResponseFuture<$3.AllowFileshareResponse> allowFileshare(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$allowFileshare, request, options: options);
  }

  /// DenyFileshare denies a peer to send files to this device
  $grpc.ResponseFuture<$3.DenyFileshareResponse> denyFileshare(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$denyFileshare, request, options: options);
  }

  /// EnableAutomaticFileshare from peer
  $grpc.ResponseFuture<$3.EnableAutomaticFileshareResponse>
      enableAutomaticFileshare(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$enableAutomaticFileshare, request,
        options: options);
  }

  /// DisableAutomaticFileshare from peer
  $grpc.ResponseFuture<$3.DisableAutomaticFileshareResponse>
      disableAutomaticFileshare(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$disableAutomaticFileshare, request,
        options: options);
  }

  $grpc.ResponseFuture<$3.ConnectResponse> connect(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$connect, request, options: options);
  }

  $grpc.ResponseFuture<$3.ConnectResponse> connectCancel(
    $3.UpdatePeerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$connectCancel, request, options: options);
  }

  /// NotifyNewTransfer notifies meshnet service about a newly created transaction so it can
  /// notify a corresponding meshnet peer
  $grpc.ResponseFuture<$4.NotifyNewTransferResponse> notifyNewTransfer(
    $4.NewTransferNotification request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$notifyNewTransfer, request, options: options);
  }

  /// GetPrivateKey is used to send self private key over to fileshare daemon
  $grpc.ResponseFuture<$3.PrivateKeyResponse> getPrivateKey(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getPrivateKey, request, options: options);
  }

  // method descriptors

  static final _$enableMeshnet =
      $grpc.ClientMethod<$0.Empty, $1.MeshnetResponse>(
          '/meshpb.Meshnet/EnableMeshnet',
          ($0.Empty value) => value.writeToBuffer(),
          $1.MeshnetResponse.fromBuffer);
  static final _$isEnabled = $grpc.ClientMethod<$0.Empty, $1.IsEnabledResponse>(
      '/meshpb.Meshnet/IsEnabled',
      ($0.Empty value) => value.writeToBuffer(),
      $1.IsEnabledResponse.fromBuffer);
  static final _$disableMeshnet =
      $grpc.ClientMethod<$0.Empty, $1.MeshnetResponse>(
          '/meshpb.Meshnet/DisableMeshnet',
          ($0.Empty value) => value.writeToBuffer(),
          $1.MeshnetResponse.fromBuffer);
  static final _$refreshMeshnet =
      $grpc.ClientMethod<$0.Empty, $1.MeshnetResponse>(
          '/meshpb.Meshnet/RefreshMeshnet',
          ($0.Empty value) => value.writeToBuffer(),
          $1.MeshnetResponse.fromBuffer);
  static final _$getInvites =
      $grpc.ClientMethod<$0.Empty, $2.GetInvitesResponse>(
          '/meshpb.Meshnet/GetInvites',
          ($0.Empty value) => value.writeToBuffer(),
          $2.GetInvitesResponse.fromBuffer);
  static final _$invite =
      $grpc.ClientMethod<$2.InviteRequest, $2.InviteResponse>(
          '/meshpb.Meshnet/Invite',
          ($2.InviteRequest value) => value.writeToBuffer(),
          $2.InviteResponse.fromBuffer);
  static final _$revokeInvite =
      $grpc.ClientMethod<$2.DenyInviteRequest, $2.RespondToInviteResponse>(
          '/meshpb.Meshnet/RevokeInvite',
          ($2.DenyInviteRequest value) => value.writeToBuffer(),
          $2.RespondToInviteResponse.fromBuffer);
  static final _$acceptInvite =
      $grpc.ClientMethod<$2.InviteRequest, $2.RespondToInviteResponse>(
          '/meshpb.Meshnet/AcceptInvite',
          ($2.InviteRequest value) => value.writeToBuffer(),
          $2.RespondToInviteResponse.fromBuffer);
  static final _$denyInvite =
      $grpc.ClientMethod<$2.DenyInviteRequest, $2.RespondToInviteResponse>(
          '/meshpb.Meshnet/DenyInvite',
          ($2.DenyInviteRequest value) => value.writeToBuffer(),
          $2.RespondToInviteResponse.fromBuffer);
  static final _$getPeers = $grpc.ClientMethod<$0.Empty, $3.GetPeersResponse>(
      '/meshpb.Meshnet/GetPeers',
      ($0.Empty value) => value.writeToBuffer(),
      $3.GetPeersResponse.fromBuffer);
  static final _$removePeer =
      $grpc.ClientMethod<$3.UpdatePeerRequest, $3.RemovePeerResponse>(
          '/meshpb.Meshnet/RemovePeer',
          ($3.UpdatePeerRequest value) => value.writeToBuffer(),
          $3.RemovePeerResponse.fromBuffer);
  static final _$changePeerNickname = $grpc.ClientMethod<
          $3.ChangePeerNicknameRequest, $3.ChangeNicknameResponse>(
      '/meshpb.Meshnet/ChangePeerNickname',
      ($3.ChangePeerNicknameRequest value) => value.writeToBuffer(),
      $3.ChangeNicknameResponse.fromBuffer);
  static final _$changeMachineNickname = $grpc.ClientMethod<
          $3.ChangeMachineNicknameRequest, $3.ChangeNicknameResponse>(
      '/meshpb.Meshnet/ChangeMachineNickname',
      ($3.ChangeMachineNicknameRequest value) => value.writeToBuffer(),
      $3.ChangeNicknameResponse.fromBuffer);
  static final _$allowRouting =
      $grpc.ClientMethod<$3.UpdatePeerRequest, $3.AllowRoutingResponse>(
          '/meshpb.Meshnet/AllowRouting',
          ($3.UpdatePeerRequest value) => value.writeToBuffer(),
          $3.AllowRoutingResponse.fromBuffer);
  static final _$denyRouting =
      $grpc.ClientMethod<$3.UpdatePeerRequest, $3.DenyRoutingResponse>(
          '/meshpb.Meshnet/DenyRouting',
          ($3.UpdatePeerRequest value) => value.writeToBuffer(),
          $3.DenyRoutingResponse.fromBuffer);
  static final _$allowIncoming =
      $grpc.ClientMethod<$3.UpdatePeerRequest, $3.AllowIncomingResponse>(
          '/meshpb.Meshnet/AllowIncoming',
          ($3.UpdatePeerRequest value) => value.writeToBuffer(),
          $3.AllowIncomingResponse.fromBuffer);
  static final _$denyIncoming =
      $grpc.ClientMethod<$3.UpdatePeerRequest, $3.DenyIncomingResponse>(
          '/meshpb.Meshnet/DenyIncoming',
          ($3.UpdatePeerRequest value) => value.writeToBuffer(),
          $3.DenyIncomingResponse.fromBuffer);
  static final _$allowLocalNetwork =
      $grpc.ClientMethod<$3.UpdatePeerRequest, $3.AllowLocalNetworkResponse>(
          '/meshpb.Meshnet/AllowLocalNetwork',
          ($3.UpdatePeerRequest value) => value.writeToBuffer(),
          $3.AllowLocalNetworkResponse.fromBuffer);
  static final _$denyLocalNetwork =
      $grpc.ClientMethod<$3.UpdatePeerRequest, $3.DenyLocalNetworkResponse>(
          '/meshpb.Meshnet/DenyLocalNetwork',
          ($3.UpdatePeerRequest value) => value.writeToBuffer(),
          $3.DenyLocalNetworkResponse.fromBuffer);
  static final _$allowFileshare =
      $grpc.ClientMethod<$3.UpdatePeerRequest, $3.AllowFileshareResponse>(
          '/meshpb.Meshnet/AllowFileshare',
          ($3.UpdatePeerRequest value) => value.writeToBuffer(),
          $3.AllowFileshareResponse.fromBuffer);
  static final _$denyFileshare =
      $grpc.ClientMethod<$3.UpdatePeerRequest, $3.DenyFileshareResponse>(
          '/meshpb.Meshnet/DenyFileshare',
          ($3.UpdatePeerRequest value) => value.writeToBuffer(),
          $3.DenyFileshareResponse.fromBuffer);
  static final _$enableAutomaticFileshare = $grpc.ClientMethod<
          $3.UpdatePeerRequest, $3.EnableAutomaticFileshareResponse>(
      '/meshpb.Meshnet/EnableAutomaticFileshare',
      ($3.UpdatePeerRequest value) => value.writeToBuffer(),
      $3.EnableAutomaticFileshareResponse.fromBuffer);
  static final _$disableAutomaticFileshare = $grpc.ClientMethod<
          $3.UpdatePeerRequest, $3.DisableAutomaticFileshareResponse>(
      '/meshpb.Meshnet/DisableAutomaticFileshare',
      ($3.UpdatePeerRequest value) => value.writeToBuffer(),
      $3.DisableAutomaticFileshareResponse.fromBuffer);
  static final _$connect =
      $grpc.ClientMethod<$3.UpdatePeerRequest, $3.ConnectResponse>(
          '/meshpb.Meshnet/Connect',
          ($3.UpdatePeerRequest value) => value.writeToBuffer(),
          $3.ConnectResponse.fromBuffer);
  static final _$connectCancel =
      $grpc.ClientMethod<$3.UpdatePeerRequest, $3.ConnectResponse>(
          '/meshpb.Meshnet/ConnectCancel',
          ($3.UpdatePeerRequest value) => value.writeToBuffer(),
          $3.ConnectResponse.fromBuffer);
  static final _$notifyNewTransfer = $grpc.ClientMethod<
          $4.NewTransferNotification, $4.NotifyNewTransferResponse>(
      '/meshpb.Meshnet/NotifyNewTransfer',
      ($4.NewTransferNotification value) => value.writeToBuffer(),
      $4.NotifyNewTransferResponse.fromBuffer);
  static final _$getPrivateKey =
      $grpc.ClientMethod<$0.Empty, $3.PrivateKeyResponse>(
          '/meshpb.Meshnet/GetPrivateKey',
          ($0.Empty value) => value.writeToBuffer(),
          $3.PrivateKeyResponse.fromBuffer);
}

@$pb.GrpcServiceName('meshpb.Meshnet')
abstract class MeshnetServiceBase extends $grpc.Service {
  $core.String get $name => 'meshpb.Meshnet';

  MeshnetServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.MeshnetResponse>(
        'EnableMeshnet',
        enableMeshnet_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.MeshnetResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.IsEnabledResponse>(
        'IsEnabled',
        isEnabled_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.IsEnabledResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.MeshnetResponse>(
        'DisableMeshnet',
        disableMeshnet_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.MeshnetResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.MeshnetResponse>(
        'RefreshMeshnet',
        refreshMeshnet_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.MeshnetResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $2.GetInvitesResponse>(
        'GetInvites',
        getInvites_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($2.GetInvitesResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.InviteRequest, $2.InviteResponse>(
        'Invite',
        invite_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.InviteRequest.fromBuffer(value),
        ($2.InviteResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$2.DenyInviteRequest, $2.RespondToInviteResponse>(
            'RevokeInvite',
            revokeInvite_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $2.DenyInviteRequest.fromBuffer(value),
            ($2.RespondToInviteResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$2.InviteRequest, $2.RespondToInviteResponse>(
            'AcceptInvite',
            acceptInvite_Pre,
            false,
            false,
            ($core.List<$core.int> value) => $2.InviteRequest.fromBuffer(value),
            ($2.RespondToInviteResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$2.DenyInviteRequest, $2.RespondToInviteResponse>(
            'DenyInvite',
            denyInvite_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $2.DenyInviteRequest.fromBuffer(value),
            ($2.RespondToInviteResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $3.GetPeersResponse>(
        'GetPeers',
        getPeers_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($3.GetPeersResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.UpdatePeerRequest, $3.RemovePeerResponse>(
        'RemovePeer',
        removePeer_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.UpdatePeerRequest.fromBuffer(value),
        ($3.RemovePeerResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.ChangePeerNicknameRequest,
            $3.ChangeNicknameResponse>(
        'ChangePeerNickname',
        changePeerNickname_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $3.ChangePeerNicknameRequest.fromBuffer(value),
        ($3.ChangeNicknameResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.ChangeMachineNicknameRequest,
            $3.ChangeNicknameResponse>(
        'ChangeMachineNickname',
        changeMachineNickname_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $3.ChangeMachineNicknameRequest.fromBuffer(value),
        ($3.ChangeNicknameResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$3.UpdatePeerRequest, $3.AllowRoutingResponse>(
            'AllowRouting',
            allowRouting_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $3.UpdatePeerRequest.fromBuffer(value),
            ($3.AllowRoutingResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$3.UpdatePeerRequest, $3.DenyRoutingResponse>(
            'DenyRouting',
            denyRouting_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $3.UpdatePeerRequest.fromBuffer(value),
            ($3.DenyRoutingResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$3.UpdatePeerRequest, $3.AllowIncomingResponse>(
            'AllowIncoming',
            allowIncoming_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $3.UpdatePeerRequest.fromBuffer(value),
            ($3.AllowIncomingResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$3.UpdatePeerRequest, $3.DenyIncomingResponse>(
            'DenyIncoming',
            denyIncoming_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $3.UpdatePeerRequest.fromBuffer(value),
            ($3.DenyIncomingResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$3.UpdatePeerRequest, $3.AllowLocalNetworkResponse>(
            'AllowLocalNetwork',
            allowLocalNetwork_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $3.UpdatePeerRequest.fromBuffer(value),
            ($3.AllowLocalNetworkResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$3.UpdatePeerRequest, $3.DenyLocalNetworkResponse>(
            'DenyLocalNetwork',
            denyLocalNetwork_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $3.UpdatePeerRequest.fromBuffer(value),
            ($3.DenyLocalNetworkResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$3.UpdatePeerRequest, $3.AllowFileshareResponse>(
            'AllowFileshare',
            allowFileshare_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $3.UpdatePeerRequest.fromBuffer(value),
            ($3.AllowFileshareResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$3.UpdatePeerRequest, $3.DenyFileshareResponse>(
            'DenyFileshare',
            denyFileshare_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $3.UpdatePeerRequest.fromBuffer(value),
            ($3.DenyFileshareResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.UpdatePeerRequest,
            $3.EnableAutomaticFileshareResponse>(
        'EnableAutomaticFileshare',
        enableAutomaticFileshare_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.UpdatePeerRequest.fromBuffer(value),
        ($3.EnableAutomaticFileshareResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.UpdatePeerRequest,
            $3.DisableAutomaticFileshareResponse>(
        'DisableAutomaticFileshare',
        disableAutomaticFileshare_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.UpdatePeerRequest.fromBuffer(value),
        ($3.DisableAutomaticFileshareResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.UpdatePeerRequest, $3.ConnectResponse>(
        'Connect',
        connect_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.UpdatePeerRequest.fromBuffer(value),
        ($3.ConnectResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.UpdatePeerRequest, $3.ConnectResponse>(
        'ConnectCancel',
        connectCancel_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.UpdatePeerRequest.fromBuffer(value),
        ($3.ConnectResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$4.NewTransferNotification,
            $4.NotifyNewTransferResponse>(
        'NotifyNewTransfer',
        notifyNewTransfer_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $4.NewTransferNotification.fromBuffer(value),
        ($4.NotifyNewTransferResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $3.PrivateKeyResponse>(
        'GetPrivateKey',
        getPrivateKey_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($3.PrivateKeyResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.MeshnetResponse> enableMeshnet_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return enableMeshnet($call, await $request);
  }

  $async.Future<$1.MeshnetResponse> enableMeshnet(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$1.IsEnabledResponse> isEnabled_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return isEnabled($call, await $request);
  }

  $async.Future<$1.IsEnabledResponse> isEnabled(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$1.MeshnetResponse> disableMeshnet_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return disableMeshnet($call, await $request);
  }

  $async.Future<$1.MeshnetResponse> disableMeshnet(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$1.MeshnetResponse> refreshMeshnet_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return refreshMeshnet($call, await $request);
  }

  $async.Future<$1.MeshnetResponse> refreshMeshnet(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$2.GetInvitesResponse> getInvites_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return getInvites($call, await $request);
  }

  $async.Future<$2.GetInvitesResponse> getInvites(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$2.InviteResponse> invite_Pre(
      $grpc.ServiceCall $call, $async.Future<$2.InviteRequest> $request) async {
    return invite($call, await $request);
  }

  $async.Future<$2.InviteResponse> invite(
      $grpc.ServiceCall call, $2.InviteRequest request);

  $async.Future<$2.RespondToInviteResponse> revokeInvite_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$2.DenyInviteRequest> $request) async {
    return revokeInvite($call, await $request);
  }

  $async.Future<$2.RespondToInviteResponse> revokeInvite(
      $grpc.ServiceCall call, $2.DenyInviteRequest request);

  $async.Future<$2.RespondToInviteResponse> acceptInvite_Pre(
      $grpc.ServiceCall $call, $async.Future<$2.InviteRequest> $request) async {
    return acceptInvite($call, await $request);
  }

  $async.Future<$2.RespondToInviteResponse> acceptInvite(
      $grpc.ServiceCall call, $2.InviteRequest request);

  $async.Future<$2.RespondToInviteResponse> denyInvite_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$2.DenyInviteRequest> $request) async {
    return denyInvite($call, await $request);
  }

  $async.Future<$2.RespondToInviteResponse> denyInvite(
      $grpc.ServiceCall call, $2.DenyInviteRequest request);

  $async.Future<$3.GetPeersResponse> getPeers_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return getPeers($call, await $request);
  }

  $async.Future<$3.GetPeersResponse> getPeers(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$3.RemovePeerResponse> removePeer_Pre($grpc.ServiceCall $call,
      $async.Future<$3.UpdatePeerRequest> $request) async {
    return removePeer($call, await $request);
  }

  $async.Future<$3.RemovePeerResponse> removePeer(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$3.ChangeNicknameResponse> changePeerNickname_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$3.ChangePeerNicknameRequest> $request) async {
    return changePeerNickname($call, await $request);
  }

  $async.Future<$3.ChangeNicknameResponse> changePeerNickname(
      $grpc.ServiceCall call, $3.ChangePeerNicknameRequest request);

  $async.Future<$3.ChangeNicknameResponse> changeMachineNickname_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$3.ChangeMachineNicknameRequest> $request) async {
    return changeMachineNickname($call, await $request);
  }

  $async.Future<$3.ChangeNicknameResponse> changeMachineNickname(
      $grpc.ServiceCall call, $3.ChangeMachineNicknameRequest request);

  $async.Future<$3.AllowRoutingResponse> allowRouting_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$3.UpdatePeerRequest> $request) async {
    return allowRouting($call, await $request);
  }

  $async.Future<$3.AllowRoutingResponse> allowRouting(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$3.DenyRoutingResponse> denyRouting_Pre($grpc.ServiceCall $call,
      $async.Future<$3.UpdatePeerRequest> $request) async {
    return denyRouting($call, await $request);
  }

  $async.Future<$3.DenyRoutingResponse> denyRouting(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$3.AllowIncomingResponse> allowIncoming_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$3.UpdatePeerRequest> $request) async {
    return allowIncoming($call, await $request);
  }

  $async.Future<$3.AllowIncomingResponse> allowIncoming(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$3.DenyIncomingResponse> denyIncoming_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$3.UpdatePeerRequest> $request) async {
    return denyIncoming($call, await $request);
  }

  $async.Future<$3.DenyIncomingResponse> denyIncoming(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$3.AllowLocalNetworkResponse> allowLocalNetwork_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$3.UpdatePeerRequest> $request) async {
    return allowLocalNetwork($call, await $request);
  }

  $async.Future<$3.AllowLocalNetworkResponse> allowLocalNetwork(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$3.DenyLocalNetworkResponse> denyLocalNetwork_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$3.UpdatePeerRequest> $request) async {
    return denyLocalNetwork($call, await $request);
  }

  $async.Future<$3.DenyLocalNetworkResponse> denyLocalNetwork(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$3.AllowFileshareResponse> allowFileshare_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$3.UpdatePeerRequest> $request) async {
    return allowFileshare($call, await $request);
  }

  $async.Future<$3.AllowFileshareResponse> allowFileshare(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$3.DenyFileshareResponse> denyFileshare_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$3.UpdatePeerRequest> $request) async {
    return denyFileshare($call, await $request);
  }

  $async.Future<$3.DenyFileshareResponse> denyFileshare(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$3.EnableAutomaticFileshareResponse>
      enableAutomaticFileshare_Pre($grpc.ServiceCall $call,
          $async.Future<$3.UpdatePeerRequest> $request) async {
    return enableAutomaticFileshare($call, await $request);
  }

  $async.Future<$3.EnableAutomaticFileshareResponse> enableAutomaticFileshare(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$3.DisableAutomaticFileshareResponse>
      disableAutomaticFileshare_Pre($grpc.ServiceCall $call,
          $async.Future<$3.UpdatePeerRequest> $request) async {
    return disableAutomaticFileshare($call, await $request);
  }

  $async.Future<$3.DisableAutomaticFileshareResponse> disableAutomaticFileshare(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$3.ConnectResponse> connect_Pre($grpc.ServiceCall $call,
      $async.Future<$3.UpdatePeerRequest> $request) async {
    return connect($call, await $request);
  }

  $async.Future<$3.ConnectResponse> connect(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$3.ConnectResponse> connectCancel_Pre($grpc.ServiceCall $call,
      $async.Future<$3.UpdatePeerRequest> $request) async {
    return connectCancel($call, await $request);
  }

  $async.Future<$3.ConnectResponse> connectCancel(
      $grpc.ServiceCall call, $3.UpdatePeerRequest request);

  $async.Future<$4.NotifyNewTransferResponse> notifyNewTransfer_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$4.NewTransferNotification> $request) async {
    return notifyNewTransfer($call, await $request);
  }

  $async.Future<$4.NotifyNewTransferResponse> notifyNewTransfer(
      $grpc.ServiceCall call, $4.NewTransferNotification request);

  $async.Future<$3.PrivateKeyResponse> getPrivateKey_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return getPrivateKey($call, await $request);
  }

  $async.Future<$3.PrivateKeyResponse> getPrivateKey(
      $grpc.ServiceCall call, $0.Empty request);
}
