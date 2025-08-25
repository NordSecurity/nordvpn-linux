// dart format width=80
// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'vpn_status.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;
/// @nodoc
mixin _$VpnStatus {

 String? get ip; String? get hostname; City? get city; Country? get country; ConnectionState get status; VpnProtocol get protocol; bool get isVirtualLocation; ConnectionParameters get connectionParameters; bool get isMeshnetRouting;
/// Create a copy of VpnStatus
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$VpnStatusCopyWith<VpnStatus> get copyWith => _$VpnStatusCopyWithImpl<VpnStatus>(this as VpnStatus, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is VpnStatus&&(identical(other.ip, ip) || other.ip == ip)&&(identical(other.hostname, hostname) || other.hostname == hostname)&&(identical(other.city, city) || other.city == city)&&(identical(other.country, country) || other.country == country)&&(identical(other.status, status) || other.status == status)&&(identical(other.protocol, protocol) || other.protocol == protocol)&&(identical(other.isVirtualLocation, isVirtualLocation) || other.isVirtualLocation == isVirtualLocation)&&(identical(other.connectionParameters, connectionParameters) || other.connectionParameters == connectionParameters)&&(identical(other.isMeshnetRouting, isMeshnetRouting) || other.isMeshnetRouting == isMeshnetRouting));
}


@override
int get hashCode => Object.hash(runtimeType,ip,hostname,city,country,status,protocol,isVirtualLocation,connectionParameters,isMeshnetRouting);

@override
String toString() {
  return 'VpnStatus(ip: $ip, hostname: $hostname, city: $city, country: $country, status: $status, protocol: $protocol, isVirtualLocation: $isVirtualLocation, connectionParameters: $connectionParameters, isMeshnetRouting: $isMeshnetRouting)';
}


}

/// @nodoc
abstract mixin class $VpnStatusCopyWith<$Res>  {
  factory $VpnStatusCopyWith(VpnStatus value, $Res Function(VpnStatus) _then) = _$VpnStatusCopyWithImpl;
@useResult
$Res call({
 String? ip, String? hostname, City? city, Country? country, ConnectionState status, VpnProtocol protocol, bool isVirtualLocation, ConnectionParameters connectionParameters, bool isMeshnetRouting
});


$CityCopyWith<$Res>? get city;$CountryCopyWith<$Res>? get country;

}
/// @nodoc
class _$VpnStatusCopyWithImpl<$Res>
    implements $VpnStatusCopyWith<$Res> {
  _$VpnStatusCopyWithImpl(this._self, this._then);

  final VpnStatus _self;
  final $Res Function(VpnStatus) _then;

/// Create a copy of VpnStatus
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? ip = freezed,Object? hostname = freezed,Object? city = freezed,Object? country = freezed,Object? status = null,Object? protocol = null,Object? isVirtualLocation = null,Object? connectionParameters = null,Object? isMeshnetRouting = null,}) {
  return _then(_self.copyWith(
ip: freezed == ip ? _self.ip : ip // ignore: cast_nullable_to_non_nullable
as String?,hostname: freezed == hostname ? _self.hostname : hostname // ignore: cast_nullable_to_non_nullable
as String?,city: freezed == city ? _self.city : city // ignore: cast_nullable_to_non_nullable
as City?,country: freezed == country ? _self.country : country // ignore: cast_nullable_to_non_nullable
as Country?,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as ConnectionState,protocol: null == protocol ? _self.protocol : protocol // ignore: cast_nullable_to_non_nullable
as VpnProtocol,isVirtualLocation: null == isVirtualLocation ? _self.isVirtualLocation : isVirtualLocation // ignore: cast_nullable_to_non_nullable
as bool,connectionParameters: null == connectionParameters ? _self.connectionParameters : connectionParameters // ignore: cast_nullable_to_non_nullable
as ConnectionParameters,isMeshnetRouting: null == isMeshnetRouting ? _self.isMeshnetRouting : isMeshnetRouting // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}
/// Create a copy of VpnStatus
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$CityCopyWith<$Res>? get city {
    if (_self.city == null) {
    return null;
  }

  return $CityCopyWith<$Res>(_self.city!, (value) {
    return _then(_self.copyWith(city: value));
  });
}/// Create a copy of VpnStatus
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$CountryCopyWith<$Res>? get country {
    if (_self.country == null) {
    return null;
  }

  return $CountryCopyWith<$Res>(_self.country!, (value) {
    return _then(_self.copyWith(country: value));
  });
}
}


/// @nodoc


class _VpnStatus extends VpnStatus {
  const _VpnStatus({required this.ip, required this.hostname, required this.city, required this.country, required this.status, required this.protocol, required this.isVirtualLocation, required this.connectionParameters, required this.isMeshnetRouting}): super._();
  

@override final  String? ip;
@override final  String? hostname;
@override final  City? city;
@override final  Country? country;
@override final  ConnectionState status;
@override final  VpnProtocol protocol;
@override final  bool isVirtualLocation;
@override final  ConnectionParameters connectionParameters;
@override final  bool isMeshnetRouting;

/// Create a copy of VpnStatus
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$VpnStatusCopyWith<_VpnStatus> get copyWith => __$VpnStatusCopyWithImpl<_VpnStatus>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _VpnStatus&&(identical(other.ip, ip) || other.ip == ip)&&(identical(other.hostname, hostname) || other.hostname == hostname)&&(identical(other.city, city) || other.city == city)&&(identical(other.country, country) || other.country == country)&&(identical(other.status, status) || other.status == status)&&(identical(other.protocol, protocol) || other.protocol == protocol)&&(identical(other.isVirtualLocation, isVirtualLocation) || other.isVirtualLocation == isVirtualLocation)&&(identical(other.connectionParameters, connectionParameters) || other.connectionParameters == connectionParameters)&&(identical(other.isMeshnetRouting, isMeshnetRouting) || other.isMeshnetRouting == isMeshnetRouting));
}


@override
int get hashCode => Object.hash(runtimeType,ip,hostname,city,country,status,protocol,isVirtualLocation,connectionParameters,isMeshnetRouting);

@override
String toString() {
  return 'VpnStatus(ip: $ip, hostname: $hostname, city: $city, country: $country, status: $status, protocol: $protocol, isVirtualLocation: $isVirtualLocation, connectionParameters: $connectionParameters, isMeshnetRouting: $isMeshnetRouting)';
}


}

/// @nodoc
abstract mixin class _$VpnStatusCopyWith<$Res> implements $VpnStatusCopyWith<$Res> {
  factory _$VpnStatusCopyWith(_VpnStatus value, $Res Function(_VpnStatus) _then) = __$VpnStatusCopyWithImpl;
@override @useResult
$Res call({
 String? ip, String? hostname, City? city, Country? country, ConnectionState status, VpnProtocol protocol, bool isVirtualLocation, ConnectionParameters connectionParameters, bool isMeshnetRouting
});


@override $CityCopyWith<$Res>? get city;@override $CountryCopyWith<$Res>? get country;

}
/// @nodoc
class __$VpnStatusCopyWithImpl<$Res>
    implements _$VpnStatusCopyWith<$Res> {
  __$VpnStatusCopyWithImpl(this._self, this._then);

  final _VpnStatus _self;
  final $Res Function(_VpnStatus) _then;

/// Create a copy of VpnStatus
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? ip = freezed,Object? hostname = freezed,Object? city = freezed,Object? country = freezed,Object? status = null,Object? protocol = null,Object? isVirtualLocation = null,Object? connectionParameters = null,Object? isMeshnetRouting = null,}) {
  return _then(_VpnStatus(
ip: freezed == ip ? _self.ip : ip // ignore: cast_nullable_to_non_nullable
as String?,hostname: freezed == hostname ? _self.hostname : hostname // ignore: cast_nullable_to_non_nullable
as String?,city: freezed == city ? _self.city : city // ignore: cast_nullable_to_non_nullable
as City?,country: freezed == country ? _self.country : country // ignore: cast_nullable_to_non_nullable
as Country?,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as ConnectionState,protocol: null == protocol ? _self.protocol : protocol // ignore: cast_nullable_to_non_nullable
as VpnProtocol,isVirtualLocation: null == isVirtualLocation ? _self.isVirtualLocation : isVirtualLocation // ignore: cast_nullable_to_non_nullable
as bool,connectionParameters: null == connectionParameters ? _self.connectionParameters : connectionParameters // ignore: cast_nullable_to_non_nullable
as ConnectionParameters,isMeshnetRouting: null == isMeshnetRouting ? _self.isMeshnetRouting : isMeshnetRouting // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}

/// Create a copy of VpnStatus
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$CityCopyWith<$Res>? get city {
    if (_self.city == null) {
    return null;
  }

  return $CityCopyWith<$Res>(_self.city!, (value) {
    return _then(_self.copyWith(city: value));
  });
}/// Create a copy of VpnStatus
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$CountryCopyWith<$Res>? get country {
    if (_self.country == null) {
    return null;
  }

  return $CountryCopyWith<$Res>(_self.country!, (value) {
    return _then(_self.copyWith(country: value));
  });
}
}

// dart format on
