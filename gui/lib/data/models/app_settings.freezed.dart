// dart format width=80
// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'app_settings.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;
/// @nodoc
mixin _$ApplicationSettings {

 bool get notifications; ConsentLevel get analyticsConsent; bool get autoConnect; ConnectArguments? get autoConnectLocation; VpnProtocol get protocol; bool get killSwitch; bool get lanDiscovery; bool get routing; bool get postQuantum; bool get obfuscatedServers; bool get virtualServers; bool get firewall; int get firewallMark; bool get customDns; List<String> get customDnsServers; bool get threatProtection; bool get tray; bool get allowList; AllowList get allowListData;
/// Create a copy of ApplicationSettings
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ApplicationSettingsCopyWith<ApplicationSettings> get copyWith => _$ApplicationSettingsCopyWithImpl<ApplicationSettings>(this as ApplicationSettings, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ApplicationSettings&&(identical(other.notifications, notifications) || other.notifications == notifications)&&(identical(other.analyticsConsent, analyticsConsent) || other.analyticsConsent == analyticsConsent)&&(identical(other.autoConnect, autoConnect) || other.autoConnect == autoConnect)&&(identical(other.autoConnectLocation, autoConnectLocation) || other.autoConnectLocation == autoConnectLocation)&&(identical(other.protocol, protocol) || other.protocol == protocol)&&(identical(other.killSwitch, killSwitch) || other.killSwitch == killSwitch)&&(identical(other.lanDiscovery, lanDiscovery) || other.lanDiscovery == lanDiscovery)&&(identical(other.routing, routing) || other.routing == routing)&&(identical(other.postQuantum, postQuantum) || other.postQuantum == postQuantum)&&(identical(other.obfuscatedServers, obfuscatedServers) || other.obfuscatedServers == obfuscatedServers)&&(identical(other.virtualServers, virtualServers) || other.virtualServers == virtualServers)&&(identical(other.firewall, firewall) || other.firewall == firewall)&&(identical(other.firewallMark, firewallMark) || other.firewallMark == firewallMark)&&(identical(other.customDns, customDns) || other.customDns == customDns)&&const DeepCollectionEquality().equals(other.customDnsServers, customDnsServers)&&(identical(other.threatProtection, threatProtection) || other.threatProtection == threatProtection)&&(identical(other.tray, tray) || other.tray == tray)&&(identical(other.allowList, allowList) || other.allowList == allowList)&&(identical(other.allowListData, allowListData) || other.allowListData == allowListData));
}


@override
int get hashCode => Object.hashAll([runtimeType,notifications,analyticsConsent,autoConnect,autoConnectLocation,protocol,killSwitch,lanDiscovery,routing,postQuantum,obfuscatedServers,virtualServers,firewall,firewallMark,customDns,const DeepCollectionEquality().hash(customDnsServers),threatProtection,tray,allowList,allowListData]);

@override
String toString() {
  return 'ApplicationSettings(notifications: $notifications, analyticsConsent: $analyticsConsent, autoConnect: $autoConnect, autoConnectLocation: $autoConnectLocation, protocol: $protocol, killSwitch: $killSwitch, lanDiscovery: $lanDiscovery, routing: $routing, postQuantum: $postQuantum, obfuscatedServers: $obfuscatedServers, virtualServers: $virtualServers, firewall: $firewall, firewallMark: $firewallMark, customDns: $customDns, customDnsServers: $customDnsServers, threatProtection: $threatProtection, tray: $tray, allowList: $allowList, allowListData: $allowListData)';
}


}

/// @nodoc
abstract mixin class $ApplicationSettingsCopyWith<$Res>  {
  factory $ApplicationSettingsCopyWith(ApplicationSettings value, $Res Function(ApplicationSettings) _then) = _$ApplicationSettingsCopyWithImpl;
@useResult
$Res call({
 bool notifications, ConsentLevel analyticsConsent, bool autoConnect, ConnectArguments? autoConnectLocation, VpnProtocol protocol, bool killSwitch, bool lanDiscovery, bool routing, bool postQuantum, bool obfuscatedServers, bool virtualServers, bool firewall, int firewallMark, bool customDns, List<String> customDnsServers, bool threatProtection, bool tray, bool allowList, AllowList allowListData
});


$AllowListCopyWith<$Res> get allowListData;

}
/// @nodoc
class _$ApplicationSettingsCopyWithImpl<$Res>
    implements $ApplicationSettingsCopyWith<$Res> {
  _$ApplicationSettingsCopyWithImpl(this._self, this._then);

  final ApplicationSettings _self;
  final $Res Function(ApplicationSettings) _then;

/// Create a copy of ApplicationSettings
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? notifications = null,Object? analyticsConsent = null,Object? autoConnect = null,Object? autoConnectLocation = freezed,Object? protocol = null,Object? killSwitch = null,Object? lanDiscovery = null,Object? routing = null,Object? postQuantum = null,Object? obfuscatedServers = null,Object? virtualServers = null,Object? firewall = null,Object? firewallMark = null,Object? customDns = null,Object? customDnsServers = null,Object? threatProtection = null,Object? tray = null,Object? allowList = null,Object? allowListData = null,}) {
  return _then(_self.copyWith(
notifications: null == notifications ? _self.notifications : notifications // ignore: cast_nullable_to_non_nullable
as bool,analyticsConsent: null == analyticsConsent ? _self.analyticsConsent : analyticsConsent // ignore: cast_nullable_to_non_nullable
as ConsentLevel,autoConnect: null == autoConnect ? _self.autoConnect : autoConnect // ignore: cast_nullable_to_non_nullable
as bool,autoConnectLocation: freezed == autoConnectLocation ? _self.autoConnectLocation : autoConnectLocation // ignore: cast_nullable_to_non_nullable
as ConnectArguments?,protocol: null == protocol ? _self.protocol : protocol // ignore: cast_nullable_to_non_nullable
as VpnProtocol,killSwitch: null == killSwitch ? _self.killSwitch : killSwitch // ignore: cast_nullable_to_non_nullable
as bool,lanDiscovery: null == lanDiscovery ? _self.lanDiscovery : lanDiscovery // ignore: cast_nullable_to_non_nullable
as bool,routing: null == routing ? _self.routing : routing // ignore: cast_nullable_to_non_nullable
as bool,postQuantum: null == postQuantum ? _self.postQuantum : postQuantum // ignore: cast_nullable_to_non_nullable
as bool,obfuscatedServers: null == obfuscatedServers ? _self.obfuscatedServers : obfuscatedServers // ignore: cast_nullable_to_non_nullable
as bool,virtualServers: null == virtualServers ? _self.virtualServers : virtualServers // ignore: cast_nullable_to_non_nullable
as bool,firewall: null == firewall ? _self.firewall : firewall // ignore: cast_nullable_to_non_nullable
as bool,firewallMark: null == firewallMark ? _self.firewallMark : firewallMark // ignore: cast_nullable_to_non_nullable
as int,customDns: null == customDns ? _self.customDns : customDns // ignore: cast_nullable_to_non_nullable
as bool,customDnsServers: null == customDnsServers ? _self.customDnsServers : customDnsServers // ignore: cast_nullable_to_non_nullable
as List<String>,threatProtection: null == threatProtection ? _self.threatProtection : threatProtection // ignore: cast_nullable_to_non_nullable
as bool,tray: null == tray ? _self.tray : tray // ignore: cast_nullable_to_non_nullable
as bool,allowList: null == allowList ? _self.allowList : allowList // ignore: cast_nullable_to_non_nullable
as bool,allowListData: null == allowListData ? _self.allowListData : allowListData // ignore: cast_nullable_to_non_nullable
as AllowList,
  ));
}
/// Create a copy of ApplicationSettings
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AllowListCopyWith<$Res> get allowListData {
  
  return $AllowListCopyWith<$Res>(_self.allowListData, (value) {
    return _then(_self.copyWith(allowListData: value));
  });
}
}


/// @nodoc


class _ApplicationSettings extends ApplicationSettings {
  const _ApplicationSettings({required this.notifications, required this.analyticsConsent, required this.autoConnect, this.autoConnectLocation, required this.protocol, required this.killSwitch, required this.lanDiscovery, required this.routing, required this.postQuantum, required this.obfuscatedServers, required this.virtualServers, required this.firewall, required this.firewallMark, required this.customDns, required final  List<String> customDnsServers, required this.threatProtection, required this.tray, required this.allowList, required this.allowListData}): _customDnsServers = customDnsServers,super._();
  

@override final  bool notifications;
@override final  ConsentLevel analyticsConsent;
@override final  bool autoConnect;
@override final  ConnectArguments? autoConnectLocation;
@override final  VpnProtocol protocol;
@override final  bool killSwitch;
@override final  bool lanDiscovery;
@override final  bool routing;
@override final  bool postQuantum;
@override final  bool obfuscatedServers;
@override final  bool virtualServers;
@override final  bool firewall;
@override final  int firewallMark;
@override final  bool customDns;
 final  List<String> _customDnsServers;
@override List<String> get customDnsServers {
  if (_customDnsServers is EqualUnmodifiableListView) return _customDnsServers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_customDnsServers);
}

@override final  bool threatProtection;
@override final  bool tray;
@override final  bool allowList;
@override final  AllowList allowListData;

/// Create a copy of ApplicationSettings
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ApplicationSettingsCopyWith<_ApplicationSettings> get copyWith => __$ApplicationSettingsCopyWithImpl<_ApplicationSettings>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _ApplicationSettings&&(identical(other.notifications, notifications) || other.notifications == notifications)&&(identical(other.analyticsConsent, analyticsConsent) || other.analyticsConsent == analyticsConsent)&&(identical(other.autoConnect, autoConnect) || other.autoConnect == autoConnect)&&(identical(other.autoConnectLocation, autoConnectLocation) || other.autoConnectLocation == autoConnectLocation)&&(identical(other.protocol, protocol) || other.protocol == protocol)&&(identical(other.killSwitch, killSwitch) || other.killSwitch == killSwitch)&&(identical(other.lanDiscovery, lanDiscovery) || other.lanDiscovery == lanDiscovery)&&(identical(other.routing, routing) || other.routing == routing)&&(identical(other.postQuantum, postQuantum) || other.postQuantum == postQuantum)&&(identical(other.obfuscatedServers, obfuscatedServers) || other.obfuscatedServers == obfuscatedServers)&&(identical(other.virtualServers, virtualServers) || other.virtualServers == virtualServers)&&(identical(other.firewall, firewall) || other.firewall == firewall)&&(identical(other.firewallMark, firewallMark) || other.firewallMark == firewallMark)&&(identical(other.customDns, customDns) || other.customDns == customDns)&&const DeepCollectionEquality().equals(other._customDnsServers, _customDnsServers)&&(identical(other.threatProtection, threatProtection) || other.threatProtection == threatProtection)&&(identical(other.tray, tray) || other.tray == tray)&&(identical(other.allowList, allowList) || other.allowList == allowList)&&(identical(other.allowListData, allowListData) || other.allowListData == allowListData));
}


@override
int get hashCode => Object.hashAll([runtimeType,notifications,analyticsConsent,autoConnect,autoConnectLocation,protocol,killSwitch,lanDiscovery,routing,postQuantum,obfuscatedServers,virtualServers,firewall,firewallMark,customDns,const DeepCollectionEquality().hash(_customDnsServers),threatProtection,tray,allowList,allowListData]);

@override
String toString() {
  return 'ApplicationSettings(notifications: $notifications, analyticsConsent: $analyticsConsent, autoConnect: $autoConnect, autoConnectLocation: $autoConnectLocation, protocol: $protocol, killSwitch: $killSwitch, lanDiscovery: $lanDiscovery, routing: $routing, postQuantum: $postQuantum, obfuscatedServers: $obfuscatedServers, virtualServers: $virtualServers, firewall: $firewall, firewallMark: $firewallMark, customDns: $customDns, customDnsServers: $customDnsServers, threatProtection: $threatProtection, tray: $tray, allowList: $allowList, allowListData: $allowListData)';
}


}

/// @nodoc
abstract mixin class _$ApplicationSettingsCopyWith<$Res> implements $ApplicationSettingsCopyWith<$Res> {
  factory _$ApplicationSettingsCopyWith(_ApplicationSettings value, $Res Function(_ApplicationSettings) _then) = __$ApplicationSettingsCopyWithImpl;
@override @useResult
$Res call({
 bool notifications, ConsentLevel analyticsConsent, bool autoConnect, ConnectArguments? autoConnectLocation, VpnProtocol protocol, bool killSwitch, bool lanDiscovery, bool routing, bool postQuantum, bool obfuscatedServers, bool virtualServers, bool firewall, int firewallMark, bool customDns, List<String> customDnsServers, bool threatProtection, bool tray, bool allowList, AllowList allowListData
});


@override $AllowListCopyWith<$Res> get allowListData;

}
/// @nodoc
class __$ApplicationSettingsCopyWithImpl<$Res>
    implements _$ApplicationSettingsCopyWith<$Res> {
  __$ApplicationSettingsCopyWithImpl(this._self, this._then);

  final _ApplicationSettings _self;
  final $Res Function(_ApplicationSettings) _then;

/// Create a copy of ApplicationSettings
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? notifications = null,Object? analyticsConsent = null,Object? autoConnect = null,Object? autoConnectLocation = freezed,Object? protocol = null,Object? killSwitch = null,Object? lanDiscovery = null,Object? routing = null,Object? postQuantum = null,Object? obfuscatedServers = null,Object? virtualServers = null,Object? firewall = null,Object? firewallMark = null,Object? customDns = null,Object? customDnsServers = null,Object? threatProtection = null,Object? tray = null,Object? allowList = null,Object? allowListData = null,}) {
  return _then(_ApplicationSettings(
notifications: null == notifications ? _self.notifications : notifications // ignore: cast_nullable_to_non_nullable
as bool,analyticsConsent: null == analyticsConsent ? _self.analyticsConsent : analyticsConsent // ignore: cast_nullable_to_non_nullable
as ConsentLevel,autoConnect: null == autoConnect ? _self.autoConnect : autoConnect // ignore: cast_nullable_to_non_nullable
as bool,autoConnectLocation: freezed == autoConnectLocation ? _self.autoConnectLocation : autoConnectLocation // ignore: cast_nullable_to_non_nullable
as ConnectArguments?,protocol: null == protocol ? _self.protocol : protocol // ignore: cast_nullable_to_non_nullable
as VpnProtocol,killSwitch: null == killSwitch ? _self.killSwitch : killSwitch // ignore: cast_nullable_to_non_nullable
as bool,lanDiscovery: null == lanDiscovery ? _self.lanDiscovery : lanDiscovery // ignore: cast_nullable_to_non_nullable
as bool,routing: null == routing ? _self.routing : routing // ignore: cast_nullable_to_non_nullable
as bool,postQuantum: null == postQuantum ? _self.postQuantum : postQuantum // ignore: cast_nullable_to_non_nullable
as bool,obfuscatedServers: null == obfuscatedServers ? _self.obfuscatedServers : obfuscatedServers // ignore: cast_nullable_to_non_nullable
as bool,virtualServers: null == virtualServers ? _self.virtualServers : virtualServers // ignore: cast_nullable_to_non_nullable
as bool,firewall: null == firewall ? _self.firewall : firewall // ignore: cast_nullable_to_non_nullable
as bool,firewallMark: null == firewallMark ? _self.firewallMark : firewallMark // ignore: cast_nullable_to_non_nullable
as int,customDns: null == customDns ? _self.customDns : customDns // ignore: cast_nullable_to_non_nullable
as bool,customDnsServers: null == customDnsServers ? _self._customDnsServers : customDnsServers // ignore: cast_nullable_to_non_nullable
as List<String>,threatProtection: null == threatProtection ? _self.threatProtection : threatProtection // ignore: cast_nullable_to_non_nullable
as bool,tray: null == tray ? _self.tray : tray // ignore: cast_nullable_to_non_nullable
as bool,allowList: null == allowList ? _self.allowList : allowList // ignore: cast_nullable_to_non_nullable
as bool,allowListData: null == allowListData ? _self.allowListData : allowListData // ignore: cast_nullable_to_non_nullable
as AllowList,
  ));
}

/// Create a copy of ApplicationSettings
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AllowListCopyWith<$Res> get allowListData {
  
  return $AllowListCopyWith<$Res>(_self.allowListData, (value) {
    return _then(_self.copyWith(allowListData: value));
  });
}
}

// dart format on
