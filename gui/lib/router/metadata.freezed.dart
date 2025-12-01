// dart format width=80
// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'metadata.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;
/// @nodoc
mixin _$RouteMetadata {

 AppRoute get route; Widget get screen; String? get displayName;  Function(BuildContext)? get onPressed;
/// Create a copy of RouteMetadata
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RouteMetadataCopyWith<RouteMetadata> get copyWith => _$RouteMetadataCopyWithImpl<RouteMetadata>(this as RouteMetadata, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RouteMetadata&&(identical(other.route, route) || other.route == route)&&(identical(other.screen, screen) || other.screen == screen)&&(identical(other.displayName, displayName) || other.displayName == displayName)&&(identical(other.onPressed, onPressed) || other.onPressed == onPressed));
}


@override
int get hashCode => Object.hash(runtimeType,route,screen,displayName,onPressed);

@override
String toString() {
  return 'RouteMetadata(route: $route, screen: $screen, displayName: $displayName, onPressed: $onPressed)';
}


}

/// @nodoc
abstract mixin class $RouteMetadataCopyWith<$Res>  {
  factory $RouteMetadataCopyWith(RouteMetadata value, $Res Function(RouteMetadata) _then) = _$RouteMetadataCopyWithImpl;
@useResult
$Res call({
 AppRoute route, Widget screen, String? displayName,  Function(BuildContext)? onPressed
});




}
/// @nodoc
class _$RouteMetadataCopyWithImpl<$Res>
    implements $RouteMetadataCopyWith<$Res> {
  _$RouteMetadataCopyWithImpl(this._self, this._then);

  final RouteMetadata _self;
  final $Res Function(RouteMetadata) _then;

/// Create a copy of RouteMetadata
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? route = null,Object? screen = null,Object? displayName = freezed,Object? onPressed = freezed,}) {
  return _then(_self.copyWith(
route: null == route ? _self.route : route // ignore: cast_nullable_to_non_nullable
as AppRoute,screen: null == screen ? _self.screen : screen // ignore: cast_nullable_to_non_nullable
as Widget,displayName: freezed == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String?,onPressed: freezed == onPressed ? _self.onPressed : onPressed // ignore: cast_nullable_to_non_nullable
as  Function(BuildContext)?,
  ));
}

}


/// @nodoc


class _RouteMetadata extends RouteMetadata {
  const _RouteMetadata({required this.route, required this.screen, this.displayName, this.onPressed}): super._();
  

@override final  AppRoute route;
@override final  Widget screen;
@override final  String? displayName;
@override final   Function(BuildContext)? onPressed;

/// Create a copy of RouteMetadata
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RouteMetadataCopyWith<_RouteMetadata> get copyWith => __$RouteMetadataCopyWithImpl<_RouteMetadata>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _RouteMetadata&&(identical(other.route, route) || other.route == route)&&(identical(other.screen, screen) || other.screen == screen)&&(identical(other.displayName, displayName) || other.displayName == displayName)&&(identical(other.onPressed, onPressed) || other.onPressed == onPressed));
}


@override
int get hashCode => Object.hash(runtimeType,route,screen,displayName,onPressed);

@override
String toString() {
  return 'RouteMetadata(route: $route, screen: $screen, displayName: $displayName, onPressed: $onPressed)';
}


}

/// @nodoc
abstract mixin class _$RouteMetadataCopyWith<$Res> implements $RouteMetadataCopyWith<$Res> {
  factory _$RouteMetadataCopyWith(_RouteMetadata value, $Res Function(_RouteMetadata) _then) = __$RouteMetadataCopyWithImpl;
@override @useResult
$Res call({
 AppRoute route, Widget screen, String? displayName,  Function(BuildContext)? onPressed
});




}
/// @nodoc
class __$RouteMetadataCopyWithImpl<$Res>
    implements _$RouteMetadataCopyWith<$Res> {
  __$RouteMetadataCopyWithImpl(this._self, this._then);

  final _RouteMetadata _self;
  final $Res Function(_RouteMetadata) _then;

/// Create a copy of RouteMetadata
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? route = null,Object? screen = null,Object? displayName = freezed,Object? onPressed = freezed,}) {
  return _then(_RouteMetadata(
route: null == route ? _self.route : route // ignore: cast_nullable_to_non_nullable
as AppRoute,screen: null == screen ? _self.screen : screen // ignore: cast_nullable_to_non_nullable
as Widget,displayName: freezed == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String?,onPressed: freezed == onPressed ? _self.onPressed : onPressed // ignore: cast_nullable_to_non_nullable
as  Function(BuildContext)?,
  ));
}


}

// dart format on
