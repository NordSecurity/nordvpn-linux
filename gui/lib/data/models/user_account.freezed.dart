// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
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

 bool get hasDipSubscription; String get name; String get email; DateTime? get vpnExpirationDate; List<CountryServersGroup>? get dedicatedIpServers; DateTime? get createdOn;
/// Create a copy of UserAccount
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$UserAccountCopyWith<UserAccount> get copyWith => _$UserAccountCopyWithImpl<UserAccount>(this as UserAccount, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is UserAccount&&(identical(other.hasDipSubscription, hasDipSubscription) || other.hasDipSubscription == hasDipSubscription)&&(identical(other.name, name) || other.name == name)&&(identical(other.email, email) || other.email == email)&&(identical(other.vpnExpirationDate, vpnExpirationDate) || other.vpnExpirationDate == vpnExpirationDate)&&const DeepCollectionEquality().equals(other.dedicatedIpServers, dedicatedIpServers)&&(identical(other.createdOn, createdOn) || other.createdOn == createdOn));
}


@override
int get hashCode => Object.hash(runtimeType,hasDipSubscription,name,email,vpnExpirationDate,const DeepCollectionEquality().hash(dedicatedIpServers),createdOn);

@override
String toString() {
  return 'UserAccount(hasDipSubscription: $hasDipSubscription, name: $name, email: $email, vpnExpirationDate: $vpnExpirationDate, dedicatedIpServers: $dedicatedIpServers, createdOn: $createdOn)';
}


}

/// @nodoc
abstract mixin class $UserAccountCopyWith<$Res>  {
  factory $UserAccountCopyWith(UserAccount value, $Res Function(UserAccount) _then) = _$UserAccountCopyWithImpl;
@useResult
$Res call({
 bool hasDipSubscription, String name, String email, DateTime? vpnExpirationDate, List<CountryServersGroup>? dedicatedIpServers, DateTime? createdOn
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
@pragma('vm:prefer-inline') @override $Res call({Object? hasDipSubscription = null,Object? name = null,Object? email = null,Object? vpnExpirationDate = freezed,Object? dedicatedIpServers = freezed,Object? createdOn = freezed,}) {
  return _then(_self.copyWith(
hasDipSubscription: null == hasDipSubscription ? _self.hasDipSubscription : hasDipSubscription // ignore: cast_nullable_to_non_nullable
as bool,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,email: null == email ? _self.email : email // ignore: cast_nullable_to_non_nullable
as String,vpnExpirationDate: freezed == vpnExpirationDate ? _self.vpnExpirationDate : vpnExpirationDate // ignore: cast_nullable_to_non_nullable
as DateTime?,dedicatedIpServers: freezed == dedicatedIpServers ? _self.dedicatedIpServers : dedicatedIpServers // ignore: cast_nullable_to_non_nullable
as List<CountryServersGroup>?,createdOn: freezed == createdOn ? _self.createdOn : createdOn // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}

}


/// Adds pattern-matching-related methods to [UserAccount].
extension UserAccountPatterns on UserAccount {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _UserAccount value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _UserAccount() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _UserAccount value)  $default,){
final _that = this;
switch (_that) {
case _UserAccount():
return $default(_that);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _UserAccount value)?  $default,){
final _that = this;
switch (_that) {
case _UserAccount() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( bool hasDipSubscription,  String name,  String email,  DateTime? vpnExpirationDate,  List<CountryServersGroup>? dedicatedIpServers,  DateTime? createdOn)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _UserAccount() when $default != null:
return $default(_that.hasDipSubscription,_that.name,_that.email,_that.vpnExpirationDate,_that.dedicatedIpServers,_that.createdOn);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( bool hasDipSubscription,  String name,  String email,  DateTime? vpnExpirationDate,  List<CountryServersGroup>? dedicatedIpServers,  DateTime? createdOn)  $default,) {final _that = this;
switch (_that) {
case _UserAccount():
return $default(_that.hasDipSubscription,_that.name,_that.email,_that.vpnExpirationDate,_that.dedicatedIpServers,_that.createdOn);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( bool hasDipSubscription,  String name,  String email,  DateTime? vpnExpirationDate,  List<CountryServersGroup>? dedicatedIpServers,  DateTime? createdOn)?  $default,) {final _that = this;
switch (_that) {
case _UserAccount() when $default != null:
return $default(_that.hasDipSubscription,_that.name,_that.email,_that.vpnExpirationDate,_that.dedicatedIpServers,_that.createdOn);case _:
  return null;

}
}

}

/// @nodoc


class _UserAccount extends UserAccount {
  const _UserAccount({required this.hasDipSubscription, required this.name, required this.email, required this.vpnExpirationDate, required final  List<CountryServersGroup>? dedicatedIpServers, required this.createdOn}): _dedicatedIpServers = dedicatedIpServers,super._();
  

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

@override final  DateTime? createdOn;

/// Create a copy of UserAccount
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$UserAccountCopyWith<_UserAccount> get copyWith => __$UserAccountCopyWithImpl<_UserAccount>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _UserAccount&&(identical(other.hasDipSubscription, hasDipSubscription) || other.hasDipSubscription == hasDipSubscription)&&(identical(other.name, name) || other.name == name)&&(identical(other.email, email) || other.email == email)&&(identical(other.vpnExpirationDate, vpnExpirationDate) || other.vpnExpirationDate == vpnExpirationDate)&&const DeepCollectionEquality().equals(other._dedicatedIpServers, _dedicatedIpServers)&&(identical(other.createdOn, createdOn) || other.createdOn == createdOn));
}


@override
int get hashCode => Object.hash(runtimeType,hasDipSubscription,name,email,vpnExpirationDate,const DeepCollectionEquality().hash(_dedicatedIpServers),createdOn);

@override
String toString() {
  return 'UserAccount(hasDipSubscription: $hasDipSubscription, name: $name, email: $email, vpnExpirationDate: $vpnExpirationDate, dedicatedIpServers: $dedicatedIpServers, createdOn: $createdOn)';
}


}

/// @nodoc
abstract mixin class _$UserAccountCopyWith<$Res> implements $UserAccountCopyWith<$Res> {
  factory _$UserAccountCopyWith(_UserAccount value, $Res Function(_UserAccount) _then) = __$UserAccountCopyWithImpl;
@override @useResult
$Res call({
 bool hasDipSubscription, String name, String email, DateTime? vpnExpirationDate, List<CountryServersGroup>? dedicatedIpServers, DateTime? createdOn
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
@override @pragma('vm:prefer-inline') $Res call({Object? hasDipSubscription = null,Object? name = null,Object? email = null,Object? vpnExpirationDate = freezed,Object? dedicatedIpServers = freezed,Object? createdOn = freezed,}) {
  return _then(_UserAccount(
hasDipSubscription: null == hasDipSubscription ? _self.hasDipSubscription : hasDipSubscription // ignore: cast_nullable_to_non_nullable
as bool,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,email: null == email ? _self.email : email // ignore: cast_nullable_to_non_nullable
as String,vpnExpirationDate: freezed == vpnExpirationDate ? _self.vpnExpirationDate : vpnExpirationDate // ignore: cast_nullable_to_non_nullable
as DateTime?,dedicatedIpServers: freezed == dedicatedIpServers ? _self._dedicatedIpServers : dedicatedIpServers // ignore: cast_nullable_to_non_nullable
as List<CountryServersGroup>?,createdOn: freezed == createdOn ? _self.createdOn : createdOn // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}


}

// dart format on
