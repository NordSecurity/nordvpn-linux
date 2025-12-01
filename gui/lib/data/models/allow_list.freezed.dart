// dart format width=80
// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'allow_list.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;
/// @nodoc
mixin _$AllowList {

 List<Subnet> get subnets; List<PortInterval> get ports;
/// Create a copy of AllowList
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AllowListCopyWith<AllowList> get copyWith => _$AllowListCopyWithImpl<AllowList>(this as AllowList, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AllowList&&const DeepCollectionEquality().equals(other.subnets, subnets)&&const DeepCollectionEquality().equals(other.ports, ports));
}


@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(subnets),const DeepCollectionEquality().hash(ports));

@override
String toString() {
  return 'AllowList(subnets: $subnets, ports: $ports)';
}


}

/// @nodoc
abstract mixin class $AllowListCopyWith<$Res>  {
  factory $AllowListCopyWith(AllowList value, $Res Function(AllowList) _then) = _$AllowListCopyWithImpl;
@useResult
$Res call({
 List<Subnet> subnets, List<PortInterval> ports
});




}
/// @nodoc
class _$AllowListCopyWithImpl<$Res>
    implements $AllowListCopyWith<$Res> {
  _$AllowListCopyWithImpl(this._self, this._then);

  final AllowList _self;
  final $Res Function(AllowList) _then;

/// Create a copy of AllowList
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? subnets = null,Object? ports = null,}) {
  return _then(_self.copyWith(
subnets: null == subnets ? _self.subnets : subnets // ignore: cast_nullable_to_non_nullable
as List<Subnet>,ports: null == ports ? _self.ports : ports // ignore: cast_nullable_to_non_nullable
as List<PortInterval>,
  ));
}

}


/// @nodoc


class _AllowList extends AllowList {
  const _AllowList({required final  List<Subnet> subnets, required final  List<PortInterval> ports}): _subnets = subnets,_ports = ports,super._();
  

 final  List<Subnet> _subnets;
@override List<Subnet> get subnets {
  if (_subnets is EqualUnmodifiableListView) return _subnets;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_subnets);
}

 final  List<PortInterval> _ports;
@override List<PortInterval> get ports {
  if (_ports is EqualUnmodifiableListView) return _ports;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_ports);
}


/// Create a copy of AllowList
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AllowListCopyWith<_AllowList> get copyWith => __$AllowListCopyWithImpl<_AllowList>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _AllowList&&const DeepCollectionEquality().equals(other._subnets, _subnets)&&const DeepCollectionEquality().equals(other._ports, _ports));
}


@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(_subnets),const DeepCollectionEquality().hash(_ports));

@override
String toString() {
  return 'AllowList(subnets: $subnets, ports: $ports)';
}


}

/// @nodoc
abstract mixin class _$AllowListCopyWith<$Res> implements $AllowListCopyWith<$Res> {
  factory _$AllowListCopyWith(_AllowList value, $Res Function(_AllowList) _then) = __$AllowListCopyWithImpl;
@override @useResult
$Res call({
 List<Subnet> subnets, List<PortInterval> ports
});




}
/// @nodoc
class __$AllowListCopyWithImpl<$Res>
    implements _$AllowListCopyWith<$Res> {
  __$AllowListCopyWithImpl(this._self, this._then);

  final _AllowList _self;
  final $Res Function(_AllowList) _then;

/// Create a copy of AllowList
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? subnets = null,Object? ports = null,}) {
  return _then(_AllowList(
subnets: null == subnets ? _self._subnets : subnets // ignore: cast_nullable_to_non_nullable
as List<Subnet>,ports: null == ports ? _self._ports : ports // ignore: cast_nullable_to_non_nullable
as List<PortInterval>,
  ));
}


}

/// @nodoc
mixin _$PortInterval {

 int get start; int get end; PortType get type;
/// Create a copy of PortInterval
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$PortIntervalCopyWith<PortInterval> get copyWith => _$PortIntervalCopyWithImpl<PortInterval>(this as PortInterval, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is PortInterval&&(identical(other.start, start) || other.start == start)&&(identical(other.end, end) || other.end == end)&&(identical(other.type, type) || other.type == type));
}


@override
int get hashCode => Object.hash(runtimeType,start,end,type);

@override
String toString() {
  return 'PortInterval(start: $start, end: $end, type: $type)';
}


}

/// @nodoc
abstract mixin class $PortIntervalCopyWith<$Res>  {
  factory $PortIntervalCopyWith(PortInterval value, $Res Function(PortInterval) _then) = _$PortIntervalCopyWithImpl;
@useResult
$Res call({
 int start, int end, PortType type
});




}
/// @nodoc
class _$PortIntervalCopyWithImpl<$Res>
    implements $PortIntervalCopyWith<$Res> {
  _$PortIntervalCopyWithImpl(this._self, this._then);

  final PortInterval _self;
  final $Res Function(PortInterval) _then;

/// Create a copy of PortInterval
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? start = null,Object? end = null,Object? type = null,}) {
  return _then(_self.copyWith(
start: null == start ? _self.start : start // ignore: cast_nullable_to_non_nullable
as int,end: null == end ? _self.end : end // ignore: cast_nullable_to_non_nullable
as int,type: null == type ? _self.type : type // ignore: cast_nullable_to_non_nullable
as PortType,
  ));
}

}


/// @nodoc


class _PortInterval extends PortInterval {
  const _PortInterval({required this.start, required this.end, required this.type}): super._();
  

@override final  int start;
@override final  int end;
@override final  PortType type;

/// Create a copy of PortInterval
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$PortIntervalCopyWith<_PortInterval> get copyWith => __$PortIntervalCopyWithImpl<_PortInterval>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _PortInterval&&(identical(other.start, start) || other.start == start)&&(identical(other.end, end) || other.end == end)&&(identical(other.type, type) || other.type == type));
}


@override
int get hashCode => Object.hash(runtimeType,start,end,type);

@override
String toString() {
  return 'PortInterval(start: $start, end: $end, type: $type)';
}


}

/// @nodoc
abstract mixin class _$PortIntervalCopyWith<$Res> implements $PortIntervalCopyWith<$Res> {
  factory _$PortIntervalCopyWith(_PortInterval value, $Res Function(_PortInterval) _then) = __$PortIntervalCopyWithImpl;
@override @useResult
$Res call({
 int start, int end, PortType type
});




}
/// @nodoc
class __$PortIntervalCopyWithImpl<$Res>
    implements _$PortIntervalCopyWith<$Res> {
  __$PortIntervalCopyWithImpl(this._self, this._then);

  final _PortInterval _self;
  final $Res Function(_PortInterval) _then;

/// Create a copy of PortInterval
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? start = null,Object? end = null,Object? type = null,}) {
  return _then(_PortInterval(
start: null == start ? _self.start : start // ignore: cast_nullable_to_non_nullable
as int,end: null == end ? _self.end : end // ignore: cast_nullable_to_non_nullable
as int,type: null == type ? _self.type : type // ignore: cast_nullable_to_non_nullable
as PortType,
  ));
}


}

/// @nodoc
mixin _$Subnet {

// string value for the subnet 0.0.0.0/32
 String get value;// int representation of the IP. It is null when value fails to be parsed
 int? get ip;// the number of bits used to describe the address /32
 int? get cidr;
/// Create a copy of Subnet
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$SubnetCopyWith<Subnet> get copyWith => _$SubnetCopyWithImpl<Subnet>(this as Subnet, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Subnet&&(identical(other.value, value) || other.value == value)&&(identical(other.ip, ip) || other.ip == ip)&&(identical(other.cidr, cidr) || other.cidr == cidr));
}


@override
int get hashCode => Object.hash(runtimeType,value,ip,cidr);

@override
String toString() {
  return 'Subnet(value: $value, ip: $ip, cidr: $cidr)';
}


}

/// @nodoc
abstract mixin class $SubnetCopyWith<$Res>  {
  factory $SubnetCopyWith(Subnet value, $Res Function(Subnet) _then) = _$SubnetCopyWithImpl;
@useResult
$Res call({
 String value, int? ip, int? cidr
});




}
/// @nodoc
class _$SubnetCopyWithImpl<$Res>
    implements $SubnetCopyWith<$Res> {
  _$SubnetCopyWithImpl(this._self, this._then);

  final Subnet _self;
  final $Res Function(Subnet) _then;

/// Create a copy of Subnet
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? value = null,Object? ip = freezed,Object? cidr = freezed,}) {
  return _then(_self.copyWith(
value: null == value ? _self.value : value // ignore: cast_nullable_to_non_nullable
as String,ip: freezed == ip ? _self.ip : ip // ignore: cast_nullable_to_non_nullable
as int?,cidr: freezed == cidr ? _self.cidr : cidr // ignore: cast_nullable_to_non_nullable
as int?,
  ));
}

}


/// @nodoc


class _Subnet extends Subnet {
  const _Subnet({required this.value, required this.ip, required this.cidr}): super._();
  

// string value for the subnet 0.0.0.0/32
@override final  String value;
// int representation of the IP. It is null when value fails to be parsed
@override final  int? ip;
// the number of bits used to describe the address /32
@override final  int? cidr;

/// Create a copy of Subnet
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$SubnetCopyWith<_Subnet> get copyWith => __$SubnetCopyWithImpl<_Subnet>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _Subnet&&(identical(other.value, value) || other.value == value)&&(identical(other.ip, ip) || other.ip == ip)&&(identical(other.cidr, cidr) || other.cidr == cidr));
}


@override
int get hashCode => Object.hash(runtimeType,value,ip,cidr);

@override
String toString() {
  return 'Subnet(value: $value, ip: $ip, cidr: $cidr)';
}


}

/// @nodoc
abstract mixin class _$SubnetCopyWith<$Res> implements $SubnetCopyWith<$Res> {
  factory _$SubnetCopyWith(_Subnet value, $Res Function(_Subnet) _then) = __$SubnetCopyWithImpl;
@override @useResult
$Res call({
 String value, int? ip, int? cidr
});




}
/// @nodoc
class __$SubnetCopyWithImpl<$Res>
    implements _$SubnetCopyWith<$Res> {
  __$SubnetCopyWithImpl(this._self, this._then);

  final _Subnet _self;
  final $Res Function(_Subnet) _then;

/// Create a copy of Subnet
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? value = null,Object? ip = freezed,Object? cidr = freezed,}) {
  return _then(_Subnet(
value: null == value ? _self.value : value // ignore: cast_nullable_to_non_nullable
as String,ip: freezed == ip ? _self.ip : ip // ignore: cast_nullable_to_non_nullable
as int?,cidr: freezed == cidr ? _self.cidr : cidr // ignore: cast_nullable_to_non_nullable
as int?,
  ));
}


}

// dart format on
