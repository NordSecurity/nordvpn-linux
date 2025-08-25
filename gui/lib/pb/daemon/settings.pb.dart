// This is a generated file - do not edit.
//
// Generated from settings.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'common.pb.dart' as $0;
import 'config/analytics_consent.pbenum.dart' as $3;
import 'config/group.pbenum.dart' as $1;
import 'config/protocol.pbenum.dart' as $4;
import 'config/technology.pbenum.dart' as $2;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

class SettingsResponse extends $pb.GeneratedMessage {
  factory SettingsResponse({
    $fixnum.Int64? type,
    Settings? data,
  }) {
    final result = create();
    if (type != null) result.type = type;
    if (data != null) result.data = data;
    return result;
  }

  SettingsResponse._();

  factory SettingsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SettingsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SettingsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(1, _omitFieldNames ? '' : 'type')
    ..aOM<Settings>(2, _omitFieldNames ? '' : 'data',
        subBuilder: Settings.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SettingsResponse clone() => SettingsResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SettingsResponse copyWith(void Function(SettingsResponse) updates) =>
      super.copyWith((message) => updates(message as SettingsResponse))
          as SettingsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SettingsResponse create() => SettingsResponse._();
  @$core.override
  SettingsResponse createEmptyInstance() => create();
  static $pb.PbList<SettingsResponse> createRepeated() =>
      $pb.PbList<SettingsResponse>();
  @$core.pragma('dart2js:noInline')
  static SettingsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SettingsResponse>(create);
  static SettingsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get type => $_getI64(0);
  @$pb.TagNumber(1)
  set type($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  @$pb.TagNumber(2)
  Settings get data => $_getN(1);
  @$pb.TagNumber(2)
  set data(Settings value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasData() => $_has(1);
  @$pb.TagNumber(2)
  void clearData() => $_clearField(2);
  @$pb.TagNumber(2)
  Settings ensureData() => $_ensure(1);
}

class AutoconnectData extends $pb.GeneratedMessage {
  factory AutoconnectData({
    $core.bool? enabled,
    $core.String? country,
    $core.String? city,
    $1.ServerGroup? serverGroup,
  }) {
    final result = create();
    if (enabled != null) result.enabled = enabled;
    if (country != null) result.country = country;
    if (city != null) result.city = city;
    if (serverGroup != null) result.serverGroup = serverGroup;
    return result;
  }

  AutoconnectData._();

  factory AutoconnectData.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AutoconnectData.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutoconnectData',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'enabled')
    ..aOS(2, _omitFieldNames ? '' : 'country')
    ..aOS(3, _omitFieldNames ? '' : 'city')
    ..e<$1.ServerGroup>(
        4, _omitFieldNames ? '' : 'serverGroup', $pb.PbFieldType.OE,
        defaultOrMaker: $1.ServerGroup.UNDEFINED,
        valueOf: $1.ServerGroup.valueOf,
        enumValues: $1.ServerGroup.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutoconnectData clone() => AutoconnectData()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutoconnectData copyWith(void Function(AutoconnectData) updates) =>
      super.copyWith((message) => updates(message as AutoconnectData))
          as AutoconnectData;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AutoconnectData create() => AutoconnectData._();
  @$core.override
  AutoconnectData createEmptyInstance() => create();
  static $pb.PbList<AutoconnectData> createRepeated() =>
      $pb.PbList<AutoconnectData>();
  @$core.pragma('dart2js:noInline')
  static AutoconnectData getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AutoconnectData>(create);
  static AutoconnectData? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get enabled => $_getBF(0);
  @$pb.TagNumber(1)
  set enabled($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEnabled() => $_has(0);
  @$pb.TagNumber(1)
  void clearEnabled() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get country => $_getSZ(1);
  @$pb.TagNumber(2)
  set country($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCountry() => $_has(1);
  @$pb.TagNumber(2)
  void clearCountry() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get city => $_getSZ(2);
  @$pb.TagNumber(3)
  set city($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCity() => $_has(2);
  @$pb.TagNumber(3)
  void clearCity() => $_clearField(3);

  @$pb.TagNumber(4)
  $1.ServerGroup get serverGroup => $_getN(3);
  @$pb.TagNumber(4)
  set serverGroup($1.ServerGroup value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasServerGroup() => $_has(3);
  @$pb.TagNumber(4)
  void clearServerGroup() => $_clearField(4);
}

class Settings extends $pb.GeneratedMessage {
  factory Settings({
    $2.Technology? technology,
    $core.bool? firewall,
    $core.bool? killSwitch,
    AutoconnectData? autoConnectData,
    $core.bool? ipv6,
    $core.bool? meshnet,
    $core.bool? routing,
    $core.int? fwmark,
    $3.ConsentMode? analyticsConsent,
    $core.Iterable<$core.String>? dns,
    $core.bool? threatProtectionLite,
    $4.Protocol? protocol,
    $core.bool? lanDiscovery,
    $0.Allowlist? allowlist,
    $core.bool? obfuscate,
    $core.bool? virtualLocation,
    $core.bool? postquantumVpn,
    UserSpecificSettings? userSettings,
  }) {
    final result = create();
    if (technology != null) result.technology = technology;
    if (firewall != null) result.firewall = firewall;
    if (killSwitch != null) result.killSwitch = killSwitch;
    if (autoConnectData != null) result.autoConnectData = autoConnectData;
    if (ipv6 != null) result.ipv6 = ipv6;
    if (meshnet != null) result.meshnet = meshnet;
    if (routing != null) result.routing = routing;
    if (fwmark != null) result.fwmark = fwmark;
    if (analyticsConsent != null) result.analyticsConsent = analyticsConsent;
    if (dns != null) result.dns.addAll(dns);
    if (threatProtectionLite != null)
      result.threatProtectionLite = threatProtectionLite;
    if (protocol != null) result.protocol = protocol;
    if (lanDiscovery != null) result.lanDiscovery = lanDiscovery;
    if (allowlist != null) result.allowlist = allowlist;
    if (obfuscate != null) result.obfuscate = obfuscate;
    if (virtualLocation != null) result.virtualLocation = virtualLocation;
    if (postquantumVpn != null) result.postquantumVpn = postquantumVpn;
    if (userSettings != null) result.userSettings = userSettings;
    return result;
  }

  Settings._();

  factory Settings.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Settings.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Settings',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..e<$2.Technology>(
        1, _omitFieldNames ? '' : 'technology', $pb.PbFieldType.OE,
        defaultOrMaker: $2.Technology.UNKNOWN_TECHNOLOGY,
        valueOf: $2.Technology.valueOf,
        enumValues: $2.Technology.values)
    ..aOB(2, _omitFieldNames ? '' : 'firewall')
    ..aOB(3, _omitFieldNames ? '' : 'killSwitch')
    ..aOM<AutoconnectData>(4, _omitFieldNames ? '' : 'autoConnectData',
        subBuilder: AutoconnectData.create)
    ..aOB(5, _omitFieldNames ? '' : 'ipv6')
    ..aOB(6, _omitFieldNames ? '' : 'meshnet')
    ..aOB(7, _omitFieldNames ? '' : 'routing')
    ..a<$core.int>(8, _omitFieldNames ? '' : 'fwmark', $pb.PbFieldType.OU3)
    ..e<$3.ConsentMode>(
        9, _omitFieldNames ? '' : 'analyticsConsent', $pb.PbFieldType.OE,
        defaultOrMaker: $3.ConsentMode.UNDEFINED,
        valueOf: $3.ConsentMode.valueOf,
        enumValues: $3.ConsentMode.values)
    ..pPS(10, _omitFieldNames ? '' : 'dns')
    ..aOB(11, _omitFieldNames ? '' : 'threatProtectionLite')
    ..e<$4.Protocol>(12, _omitFieldNames ? '' : 'protocol', $pb.PbFieldType.OE,
        defaultOrMaker: $4.Protocol.UNKNOWN_PROTOCOL,
        valueOf: $4.Protocol.valueOf,
        enumValues: $4.Protocol.values)
    ..aOB(13, _omitFieldNames ? '' : 'lanDiscovery')
    ..aOM<$0.Allowlist>(14, _omitFieldNames ? '' : 'allowlist',
        subBuilder: $0.Allowlist.create)
    ..aOB(15, _omitFieldNames ? '' : 'obfuscate')
    ..aOB(16, _omitFieldNames ? '' : 'virtualLocation',
        protoName: 'virtualLocation')
    ..aOB(17, _omitFieldNames ? '' : 'postquantumVpn')
    ..aOM<UserSpecificSettings>(18, _omitFieldNames ? '' : 'userSettings',
        subBuilder: UserSpecificSettings.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Settings clone() => Settings()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Settings copyWith(void Function(Settings) updates) =>
      super.copyWith((message) => updates(message as Settings)) as Settings;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Settings create() => Settings._();
  @$core.override
  Settings createEmptyInstance() => create();
  static $pb.PbList<Settings> createRepeated() => $pb.PbList<Settings>();
  @$core.pragma('dart2js:noInline')
  static Settings getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Settings>(create);
  static Settings? _defaultInstance;

  @$pb.TagNumber(1)
  $2.Technology get technology => $_getN(0);
  @$pb.TagNumber(1)
  set technology($2.Technology value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasTechnology() => $_has(0);
  @$pb.TagNumber(1)
  void clearTechnology() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get firewall => $_getBF(1);
  @$pb.TagNumber(2)
  set firewall($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFirewall() => $_has(1);
  @$pb.TagNumber(2)
  void clearFirewall() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get killSwitch => $_getBF(2);
  @$pb.TagNumber(3)
  set killSwitch($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasKillSwitch() => $_has(2);
  @$pb.TagNumber(3)
  void clearKillSwitch() => $_clearField(3);

  @$pb.TagNumber(4)
  AutoconnectData get autoConnectData => $_getN(3);
  @$pb.TagNumber(4)
  set autoConnectData(AutoconnectData value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasAutoConnectData() => $_has(3);
  @$pb.TagNumber(4)
  void clearAutoConnectData() => $_clearField(4);
  @$pb.TagNumber(4)
  AutoconnectData ensureAutoConnectData() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.bool get ipv6 => $_getBF(4);
  @$pb.TagNumber(5)
  set ipv6($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasIpv6() => $_has(4);
  @$pb.TagNumber(5)
  void clearIpv6() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.bool get meshnet => $_getBF(5);
  @$pb.TagNumber(6)
  set meshnet($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasMeshnet() => $_has(5);
  @$pb.TagNumber(6)
  void clearMeshnet() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get routing => $_getBF(6);
  @$pb.TagNumber(7)
  set routing($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasRouting() => $_has(6);
  @$pb.TagNumber(7)
  void clearRouting() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.int get fwmark => $_getIZ(7);
  @$pb.TagNumber(8)
  set fwmark($core.int value) => $_setUnsignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasFwmark() => $_has(7);
  @$pb.TagNumber(8)
  void clearFwmark() => $_clearField(8);

  @$pb.TagNumber(9)
  $3.ConsentMode get analyticsConsent => $_getN(8);
  @$pb.TagNumber(9)
  set analyticsConsent($3.ConsentMode value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasAnalyticsConsent() => $_has(8);
  @$pb.TagNumber(9)
  void clearAnalyticsConsent() => $_clearField(9);

  @$pb.TagNumber(10)
  $pb.PbList<$core.String> get dns => $_getList(9);

  @$pb.TagNumber(11)
  $core.bool get threatProtectionLite => $_getBF(10);
  @$pb.TagNumber(11)
  set threatProtectionLite($core.bool value) => $_setBool(10, value);
  @$pb.TagNumber(11)
  $core.bool hasThreatProtectionLite() => $_has(10);
  @$pb.TagNumber(11)
  void clearThreatProtectionLite() => $_clearField(11);

  @$pb.TagNumber(12)
  $4.Protocol get protocol => $_getN(11);
  @$pb.TagNumber(12)
  set protocol($4.Protocol value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasProtocol() => $_has(11);
  @$pb.TagNumber(12)
  void clearProtocol() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.bool get lanDiscovery => $_getBF(12);
  @$pb.TagNumber(13)
  set lanDiscovery($core.bool value) => $_setBool(12, value);
  @$pb.TagNumber(13)
  $core.bool hasLanDiscovery() => $_has(12);
  @$pb.TagNumber(13)
  void clearLanDiscovery() => $_clearField(13);

  @$pb.TagNumber(14)
  $0.Allowlist get allowlist => $_getN(13);
  @$pb.TagNumber(14)
  set allowlist($0.Allowlist value) => $_setField(14, value);
  @$pb.TagNumber(14)
  $core.bool hasAllowlist() => $_has(13);
  @$pb.TagNumber(14)
  void clearAllowlist() => $_clearField(14);
  @$pb.TagNumber(14)
  $0.Allowlist ensureAllowlist() => $_ensure(13);

  @$pb.TagNumber(15)
  $core.bool get obfuscate => $_getBF(14);
  @$pb.TagNumber(15)
  set obfuscate($core.bool value) => $_setBool(14, value);
  @$pb.TagNumber(15)
  $core.bool hasObfuscate() => $_has(14);
  @$pb.TagNumber(15)
  void clearObfuscate() => $_clearField(15);

  @$pb.TagNumber(16)
  $core.bool get virtualLocation => $_getBF(15);
  @$pb.TagNumber(16)
  set virtualLocation($core.bool value) => $_setBool(15, value);
  @$pb.TagNumber(16)
  $core.bool hasVirtualLocation() => $_has(15);
  @$pb.TagNumber(16)
  void clearVirtualLocation() => $_clearField(16);

  @$pb.TagNumber(17)
  $core.bool get postquantumVpn => $_getBF(16);
  @$pb.TagNumber(17)
  set postquantumVpn($core.bool value) => $_setBool(16, value);
  @$pb.TagNumber(17)
  $core.bool hasPostquantumVpn() => $_has(16);
  @$pb.TagNumber(17)
  void clearPostquantumVpn() => $_clearField(17);

  @$pb.TagNumber(18)
  UserSpecificSettings get userSettings => $_getN(17);
  @$pb.TagNumber(18)
  set userSettings(UserSpecificSettings value) => $_setField(18, value);
  @$pb.TagNumber(18)
  $core.bool hasUserSettings() => $_has(17);
  @$pb.TagNumber(18)
  void clearUserSettings() => $_clearField(18);
  @$pb.TagNumber(18)
  UserSpecificSettings ensureUserSettings() => $_ensure(17);
}

class UserSpecificSettings extends $pb.GeneratedMessage {
  factory UserSpecificSettings({
    $fixnum.Int64? uid,
    $core.bool? notify,
    $core.bool? tray,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    if (notify != null) result.notify = notify;
    if (tray != null) result.tray = tray;
    return result;
  }

  UserSpecificSettings._();

  factory UserSpecificSettings.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UserSpecificSettings.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UserSpecificSettings',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(1, _omitFieldNames ? '' : 'uid')
    ..aOB(2, _omitFieldNames ? '' : 'notify')
    ..aOB(3, _omitFieldNames ? '' : 'tray')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UserSpecificSettings clone() =>
      UserSpecificSettings()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UserSpecificSettings copyWith(void Function(UserSpecificSettings) updates) =>
      super.copyWith((message) => updates(message as UserSpecificSettings))
          as UserSpecificSettings;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UserSpecificSettings create() => UserSpecificSettings._();
  @$core.override
  UserSpecificSettings createEmptyInstance() => create();
  static $pb.PbList<UserSpecificSettings> createRepeated() =>
      $pb.PbList<UserSpecificSettings>();
  @$core.pragma('dart2js:noInline')
  static UserSpecificSettings getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UserSpecificSettings>(create);
  static UserSpecificSettings? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get uid => $_getI64(0);
  @$pb.TagNumber(1)
  set uid($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get notify => $_getBF(1);
  @$pb.TagNumber(2)
  set notify($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNotify() => $_has(1);
  @$pb.TagNumber(2)
  void clearNotify() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get tray => $_getBF(2);
  @$pb.TagNumber(3)
  set tray($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTray() => $_has(2);
  @$pb.TagNumber(3)
  void clearTray() => $_clearField(3);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
