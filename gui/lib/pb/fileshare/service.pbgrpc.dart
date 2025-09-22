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

import 'fileshare.pb.dart' as $0;

export 'service.pb.dart';

@$pb.GrpcServiceName('filesharepb.Fileshare')
class FileshareClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  FileshareClient(super.channel, {super.options, super.interceptors});

  /// Ping to test connection between CLI and Fileshare daemon
  $grpc.ResponseFuture<$0.Empty> ping(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$ping, request, options: options);
  }

  /// Stop
  $grpc.ResponseFuture<$0.Empty> stop(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$stop, request, options: options);
  }

  /// Send a file to a peer
  $grpc.ResponseStream<$0.StatusResponse> send(
    $0.SendRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(_$send, $async.Stream.fromIterable([request]),
        options: options);
  }

  /// Accept a request from another peer to send you a file
  $grpc.ResponseStream<$0.StatusResponse> accept(
    $0.AcceptRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(_$accept, $async.Stream.fromIterable([request]),
        options: options);
  }

  /// Reject a request from another peer to send you a file
  $grpc.ResponseFuture<$0.Error> cancel(
    $0.CancelRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$cancel, request, options: options);
  }

  /// List all transfers
  $grpc.ResponseStream<$0.ListResponse> list(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(_$list, $async.Stream.fromIterable([request]),
        options: options);
  }

  /// Cancel file transfer to another peer
  $grpc.ResponseFuture<$0.Error> cancelFile(
    $0.CancelFileRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$cancelFile, request, options: options);
  }

  /// SetNotifications about transfer status changes
  $grpc.ResponseFuture<$0.SetNotificationsResponse> setNotifications(
    $0.SetNotificationsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setNotifications, request, options: options);
  }

  /// PurgeTransfersUntil provided time from fileshare implementation storage
  $grpc.ResponseFuture<$0.Error> purgeTransfersUntil(
    $0.PurgeTransfersUntilRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$purgeTransfersUntil, request, options: options);
  }

  // method descriptors

  static final _$ping = $grpc.ClientMethod<$0.Empty, $0.Empty>(
      '/filesharepb.Fileshare/Ping',
      ($0.Empty value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
  static final _$stop = $grpc.ClientMethod<$0.Empty, $0.Empty>(
      '/filesharepb.Fileshare/Stop',
      ($0.Empty value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
  static final _$send = $grpc.ClientMethod<$0.SendRequest, $0.StatusResponse>(
      '/filesharepb.Fileshare/Send',
      ($0.SendRequest value) => value.writeToBuffer(),
      $0.StatusResponse.fromBuffer);
  static final _$accept =
      $grpc.ClientMethod<$0.AcceptRequest, $0.StatusResponse>(
          '/filesharepb.Fileshare/Accept',
          ($0.AcceptRequest value) => value.writeToBuffer(),
          $0.StatusResponse.fromBuffer);
  static final _$cancel = $grpc.ClientMethod<$0.CancelRequest, $0.Error>(
      '/filesharepb.Fileshare/Cancel',
      ($0.CancelRequest value) => value.writeToBuffer(),
      $0.Error.fromBuffer);
  static final _$list = $grpc.ClientMethod<$0.Empty, $0.ListResponse>(
      '/filesharepb.Fileshare/List',
      ($0.Empty value) => value.writeToBuffer(),
      $0.ListResponse.fromBuffer);
  static final _$cancelFile =
      $grpc.ClientMethod<$0.CancelFileRequest, $0.Error>(
          '/filesharepb.Fileshare/CancelFile',
          ($0.CancelFileRequest value) => value.writeToBuffer(),
          $0.Error.fromBuffer);
  static final _$setNotifications = $grpc.ClientMethod<
          $0.SetNotificationsRequest, $0.SetNotificationsResponse>(
      '/filesharepb.Fileshare/SetNotifications',
      ($0.SetNotificationsRequest value) => value.writeToBuffer(),
      $0.SetNotificationsResponse.fromBuffer);
  static final _$purgeTransfersUntil =
      $grpc.ClientMethod<$0.PurgeTransfersUntilRequest, $0.Error>(
          '/filesharepb.Fileshare/PurgeTransfersUntil',
          ($0.PurgeTransfersUntilRequest value) => value.writeToBuffer(),
          $0.Error.fromBuffer);
}

@$pb.GrpcServiceName('filesharepb.Fileshare')
abstract class FileshareServiceBase extends $grpc.Service {
  $core.String get $name => 'filesharepb.Fileshare';

  FileshareServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.Empty>(
        'Ping',
        ping_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.Empty>(
        'Stop',
        stop_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.SendRequest, $0.StatusResponse>(
        'Send',
        send_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $0.SendRequest.fromBuffer(value),
        ($0.StatusResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AcceptRequest, $0.StatusResponse>(
        'Accept',
        accept_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $0.AcceptRequest.fromBuffer(value),
        ($0.StatusResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CancelRequest, $0.Error>(
        'Cancel',
        cancel_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.CancelRequest.fromBuffer(value),
        ($0.Error value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $0.ListResponse>(
        'List',
        list_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($0.ListResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CancelFileRequest, $0.Error>(
        'CancelFile',
        cancelFile_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.CancelFileRequest.fromBuffer(value),
        ($0.Error value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.SetNotificationsRequest,
            $0.SetNotificationsResponse>(
        'SetNotifications',
        setNotifications_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.SetNotificationsRequest.fromBuffer(value),
        ($0.SetNotificationsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.PurgeTransfersUntilRequest, $0.Error>(
        'PurgeTransfersUntil',
        purgeTransfersUntil_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.PurgeTransfersUntilRequest.fromBuffer(value),
        ($0.Error value) => value.writeToBuffer()));
  }

  $async.Future<$0.Empty> ping_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return ping($call, await $request);
  }

  $async.Future<$0.Empty> ping($grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.Empty> stop_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return stop($call, await $request);
  }

  $async.Future<$0.Empty> stop($grpc.ServiceCall call, $0.Empty request);

  $async.Stream<$0.StatusResponse> send_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.SendRequest> $request) async* {
    yield* send($call, await $request);
  }

  $async.Stream<$0.StatusResponse> send(
      $grpc.ServiceCall call, $0.SendRequest request);

  $async.Stream<$0.StatusResponse> accept_Pre($grpc.ServiceCall $call,
      $async.Future<$0.AcceptRequest> $request) async* {
    yield* accept($call, await $request);
  }

  $async.Stream<$0.StatusResponse> accept(
      $grpc.ServiceCall call, $0.AcceptRequest request);

  $async.Future<$0.Error> cancel_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.CancelRequest> $request) async {
    return cancel($call, await $request);
  }

  $async.Future<$0.Error> cancel(
      $grpc.ServiceCall call, $0.CancelRequest request);

  $async.Stream<$0.ListResponse> list_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async* {
    yield* list($call, await $request);
  }

  $async.Stream<$0.ListResponse> list($grpc.ServiceCall call, $0.Empty request);

  $async.Future<$0.Error> cancelFile_Pre($grpc.ServiceCall $call,
      $async.Future<$0.CancelFileRequest> $request) async {
    return cancelFile($call, await $request);
  }

  $async.Future<$0.Error> cancelFile(
      $grpc.ServiceCall call, $0.CancelFileRequest request);

  $async.Future<$0.SetNotificationsResponse> setNotifications_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.SetNotificationsRequest> $request) async {
    return setNotifications($call, await $request);
  }

  $async.Future<$0.SetNotificationsResponse> setNotifications(
      $grpc.ServiceCall call, $0.SetNotificationsRequest request);

  $async.Future<$0.Error> purgeTransfersUntil_Pre($grpc.ServiceCall $call,
      $async.Future<$0.PurgeTransfersUntilRequest> $request) async {
    return purgeTransfersUntil($call, await $request);
  }

  $async.Future<$0.Error> purgeTransfersUntil(
      $grpc.ServiceCall call, $0.PurgeTransfersUntilRequest request);
}
