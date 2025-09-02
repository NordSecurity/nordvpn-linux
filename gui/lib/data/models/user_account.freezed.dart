// dart format width=80
// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'user_account.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;
/// @nodoc
mixin _$UserAccount {

 bool get hasDipSubscription; String get name; String get email; DateTime? get vpnExpirationDate; List<CountryServersGroup>? get dedicatedIpServers;
/// Create a copy of UserAccount
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$UserAccountCopyWith<UserAccount> get copyWith => _$UserAccountCopyWithImpl<UserAccount>(this as UserAccount, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is UserAccount&&(identical(other.hasDipSubscription, hasDipSubscription) || other.hasDipSubscription == hasDipSubscription)&&(identical(other.name, name) || other.name == name)&&(identical(other.email, email) || other.email == email)&&(identical(other.vpnExpirationDate, vpnExpirationDate) || other.vpnExpirationDate == vpnExpirationDate)&&const DeepCollectionEquality().equals(other.dedicatedIpServers, dedicatedIpServers));
}


@override
int get hashCode => Object.hash(runtimeType,hasDipSubscription,name,email,vpnExpirationDate,const DeepCollectionEquality().hash(dedicatedIpServers));

@override
String toString() {
  return 'UserAccount(hasDipSubscription: $hasDipSubscription, name: $name, email: $email, vpnExpirationDate: $vpnExpirationDate, dedicatedIpServers: $dedicatedIpServers)';
}


}

/// @nodoc
abstract mixin class $UserAccountCopyWith<$Res>  {
  factory $UserAccountCopyWith(UserAccount value, $Res Function(UserAccount) _then) = _$UserAccountCopyWithImpl;
@useResult
$Res call({
 bool hasDipSubscription, String name, String email, DateTime? vpnExpirationDate, List<CountryServersGroup>? dedicatedIpServers
});




}
/// @nodoc
class _$UserAccountCopyWithImpl<$Res>
    implements $UserAccountCopyWith<$Res> {
  _$UserAccountCopyWithImpl(this._self, this._then);

  final UserAccount _self;
  final $Res Function(UserAccount) _then;

/// Create a copy of UserAccount
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? hasDipSubscription = null,Object? name = null,Object? email = null,Object? vpnExpirationDate = freezed,Object? dedicatedIpServers = freezed,}) {
  return _then(_self.copyWith(
hasDipSubscription: null == hasDipSubscription ? _self.hasDipSubscription : hasDipSubscription // ignore: cast_nullable_to_non_nullable
as bool,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,email: null == email ? _self.email : email // ignore: cast_nullable_to_non_nullable
as String,vpnExpirationDate: freezed == vpnExpirationDate ? _self.vpnExpirationDate : vpnExpirationDate // ignore: cast_nullable_to_non_nullable
as DateTime?,dedicatedIpServers: freezed == dedicatedIpServers ? _self.dedicatedIpServers : dedicatedIpServers // ignore: cast_nullable_to_non_nullable
as List<CountryServersGroup>?,
  ));
}

}


/// @nodoc


class _UserAccount extends UserAccount {
  const _UserAccount({required this.hasDipSubscription, required this.name, required this.email, required this.vpnExpirationDate, required final  List<CountryServersGroup>? dedicatedIpServers}): _dedicatedIpServers = dedicatedIpServers,super._();
  

@override final  bool hasDipSubscription;
@override final  String name;
@override final  String email;
@override final  DateTime? vpnExpirationDate;
 final  List<CountryServersGroup>? _dedicatedIpServers;
@override List<CountryServersGroup>? get dedicatedIpServers {
  final value = _dedicatedIpServers;
  if (value == null) return null;
  if (_dedicatedIpServers is EqualUnmodifiableListView) return _dedicatedIpServers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(value);
}


/// Create a copy of UserAccount
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$UserAccountCopyWith<_UserAccount> get copyWith => __$UserAccountCopyWithImpl<_UserAccount>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _UserAccount&&(identical(other.hasDipSubscription, hasDipSubscription) || other.hasDipSubscription == hasDipSubscription)&&(identical(other.name, name) || other.name == name)&&(identical(other.email, email) || other.email == email)&&(identical(other.vpnExpirationDate, vpnExpirationDate) || other.vpnExpirationDate == vpnExpirationDate)&&const DeepCollectionEquality().equals(other._dedicatedIpServers, _dedicatedIpServers));
}


@override
int get hashCode => Object.hash(runtimeType,hasDipSubscription,name,email,vpnExpirationDate,const DeepCollectionEquality().hash(_dedicatedIpServers));

@override
String toString() {
  return 'UserAccount(hasDipSubscription: $hasDipSubscription, name: $name, email: $email, vpnExpirationDate: $vpnExpirationDate, dedicatedIpServers: $dedicatedIpServers)';
}


}

/// @nodoc
abstract mixin class _$UserAccountCopyWith<$Res> implements $UserAccountCopyWith<$Res> {
  factory _$UserAccountCopyWith(_UserAccount value, $Res Function(_UserAccount) _then) = __$UserAccountCopyWithImpl;
@override @useResult
$Res call({
 bool hasDipSubscription, String name, String email, DateTime? vpnExpirationDate, List<CountryServersGroup>? dedicatedIpServers
});




}
/// @nodoc
class __$UserAccountCopyWithImpl<$Res>
    implements _$UserAccountCopyWith<$Res> {
  __$UserAccountCopyWithImpl(this._self, this._then);

  final _UserAccount _self;
  final $Res Function(_UserAccount) _then;

/// Create a copy of UserAccount
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? hasDipSubscription = null,Object? name = null,Object? email = null,Object? vpnExpirationDate = freezed,Object? dedicatedIpServers = freezed,}) {
  return _then(_UserAccount(
hasDipSubscription: null == hasDipSubscription ? _self.hasDipSubscription : hasDipSubscription // ignore: cast_nullable_to_non_nullable
as bool,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,email: null == email ? _self.email : email // ignore: cast_nullable_to_non_nullable
as String,vpnExpirationDate: freezed == vpnExpirationDate ? _self.vpnExpirationDate : vpnExpirationDate // ignore: cast_nullable_to_non_nullable
as DateTime?,dedicatedIpServers: freezed == dedicatedIpServers ? _self._dedicatedIpServers : dedicatedIpServers // ignore: cast_nullable_to_non_nullable
as List<CountryServersGroup>?,
  ));
}


}

// dart format on
