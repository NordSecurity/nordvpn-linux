// This is a generated file - do not edit.
//
// Generated from state.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'settings.pb.dart' as $1;
import 'state.pbenum.dart';
import 'status.pb.dart' as $0;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'state.pbenum.dart';

class LoginEvent extends $pb.GeneratedMessage {
  factory LoginEvent({
    LoginEventType? type,
  }) {
    final result = create();
    if (type != null) result.type = type;
    return result;
  }

  LoginEvent._();

  factory LoginEvent.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LoginEvent.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginEvent',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..e<LoginEventType>(1, _omitFieldNames ? '' : 'type', $pb.PbFieldType.OE,
        defaultOrMaker: LoginEventType.LOGIN,
        valueOf: LoginEventType.valueOf,
        enumValues: LoginEventType.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginEvent clone() => LoginEvent()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginEvent copyWith(void Function(LoginEvent) updates) =>
      super.copyWith((message) => updates(message as LoginEvent)) as LoginEvent;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LoginEvent create() => LoginEvent._();
  @$core.override
  LoginEvent createEmptyInstance() => create();
  static $pb.PbList<LoginEvent> createRepeated() => $pb.PbList<LoginEvent>();
  @$core.pragma('dart2js:noInline')
  static LoginEvent getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LoginEvent>(create);
  static LoginEvent? _defaultInstance;

  @$pb.TagNumber(1)
  LoginEventType get type => $_getN(0);
  @$pb.TagNumber(1)
  set type(LoginEventType value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);
}

class AccountModification extends $pb.GeneratedMessage {
  factory AccountModification({
    $core.String? expiresAt,
  }) {
    final result = create();
    if (expiresAt != null) result.expiresAt = expiresAt;
    return result;
  }

  AccountModification._();

  factory AccountModification.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AccountModification.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AccountModification',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'expiresAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccountModification clone() => AccountModification()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccountModification copyWith(void Function(AccountModification) updates) =>
      super.copyWith((message) => updates(message as AccountModification))
          as AccountModification;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AccountModification create() => AccountModification._();
  @$core.override
  AccountModification createEmptyInstance() => create();
  static $pb.PbList<AccountModification> createRepeated() =>
      $pb.PbList<AccountModification>();
  @$core.pragma('dart2js:noInline')
  static AccountModification getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AccountModification>(create);
  static AccountModification? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get expiresAt => $_getSZ(0);
  @$pb.TagNumber(1)
  set expiresAt($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasExpiresAt() => $_has(0);
  @$pb.TagNumber(1)
  void clearExpiresAt() => $_clearField(1);
}

enum AppState_State {
  error,
  connectionStatus,
  loginEvent,
  settingsChange,
  updateEvent,
  accountModification,
  notSet
}

class AppState extends $pb.GeneratedMessage {
  factory AppState({
    AppStateError? error,
    $0.StatusResponse? connectionStatus,
    LoginEvent? loginEvent,
    $1.Settings? settingsChange,
    UpdateEvent? updateEvent,
    AccountModification? accountModification,
  }) {
    final result = create();
    if (error != null) result.error = error;
    if (connectionStatus != null) result.connectionStatus = connectionStatus;
    if (loginEvent != null) result.loginEvent = loginEvent;
    if (settingsChange != null) result.settingsChange = settingsChange;
    if (updateEvent != null) result.updateEvent = updateEvent;
    if (accountModification != null)
      result.accountModification = accountModification;
    return result;
  }

  AppState._();

  factory AppState.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AppState.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, AppState_State> _AppState_StateByTag = {
    1: AppState_State.error,
    2: AppState_State.connectionStatus,
    3: AppState_State.loginEvent,
    4: AppState_State.settingsChange,
    5: AppState_State.updateEvent,
    6: AppState_State.accountModification,
    0: AppState_State.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AppState',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2, 3, 4, 5, 6])
    ..e<AppStateError>(1, _omitFieldNames ? '' : 'error', $pb.PbFieldType.OE,
        defaultOrMaker: AppStateError.FAILED_TO_GET_UID,
        valueOf: AppStateError.valueOf,
        enumValues: AppStateError.values)
    ..aOM<$0.StatusResponse>(2, _omitFieldNames ? '' : 'connectionStatus',
        subBuilder: $0.StatusResponse.create)
    ..aOM<LoginEvent>(3, _omitFieldNames ? '' : 'loginEvent',
        subBuilder: LoginEvent.create)
    ..aOM<$1.Settings>(4, _omitFieldNames ? '' : 'settingsChange',
        subBuilder: $1.Settings.create)
    ..e<UpdateEvent>(
        5, _omitFieldNames ? '' : 'updateEvent', $pb.PbFieldType.OE,
        defaultOrMaker: UpdateEvent.SERVERS_LIST_UPDATE,
        valueOf: UpdateEvent.valueOf,
        enumValues: UpdateEvent.values)
    ..aOM<AccountModification>(6, _omitFieldNames ? '' : 'accountModification',
        subBuilder: AccountModification.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AppState clone() => AppState()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AppState copyWith(void Function(AppState) updates) =>
      super.copyWith((message) => updates(message as AppState)) as AppState;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AppState create() => AppState._();
  @$core.override
  AppState createEmptyInstance() => create();
  static $pb.PbList<AppState> createRepeated() => $pb.PbList<AppState>();
  @$core.pragma('dart2js:noInline')
  static AppState getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AppState>(create);
  static AppState? _defaultInstance;

  AppState_State whichState() => _AppState_StateByTag[$_whichOneof(0)]!;
  void clearState() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  AppStateError get error => $_getN(0);
  @$pb.TagNumber(1)
  set error(AppStateError value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasError() => $_has(0);
  @$pb.TagNumber(1)
  void clearError() => $_clearField(1);

  @$pb.TagNumber(2)
  $0.StatusResponse get connectionStatus => $_getN(1);
  @$pb.TagNumber(2)
  set connectionStatus($0.StatusResponse value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasConnectionStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearConnectionStatus() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.StatusResponse ensureConnectionStatus() => $_ensure(1);

  @$pb.TagNumber(3)
  LoginEvent get loginEvent => $_getN(2);
  @$pb.TagNumber(3)
  set loginEvent(LoginEvent value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasLoginEvent() => $_has(2);
  @$pb.TagNumber(3)
  void clearLoginEvent() => $_clearField(3);
  @$pb.TagNumber(3)
  LoginEvent ensureLoginEvent() => $_ensure(2);

  @$pb.TagNumber(4)
  $1.Settings get settingsChange => $_getN(3);
  @$pb.TagNumber(4)
  set settingsChange($1.Settings value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasSettingsChange() => $_has(3);
  @$pb.TagNumber(4)
  void clearSettingsChange() => $_clearField(4);
  @$pb.TagNumber(4)
  $1.Settings ensureSettingsChange() => $_ensure(3);

  @$pb.TagNumber(5)
  UpdateEvent get updateEvent => $_getN(4);
  @$pb.TagNumber(5)
  set updateEvent(UpdateEvent value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasUpdateEvent() => $_has(4);
  @$pb.TagNumber(5)
  void clearUpdateEvent() => $_clearField(5);

  @$pb.TagNumber(6)
  AccountModification get accountModification => $_getN(5);
  @$pb.TagNumber(6)
  set accountModification(AccountModification value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasAccountModification() => $_has(5);
  @$pb.TagNumber(6)
  void clearAccountModification() => $_clearField(6);
  @$pb.TagNumber(6)
  AccountModification ensureAccountModification() => $_ensure(5);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
