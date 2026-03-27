// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'nav_rail_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$NavRailThemeTailorMixin on ThemeExtension<NavRailTheme> {
  Color get railBg;
  double get railWidth;
  double get containerWidth;
  double get containerHeight;
  double get betweenIconsGap;
  double get iconsPaddingTop;
  double get iconsMargin;
  BorderRadius get radius;
  Color get selectedItemBg;

  @override
  NavRailTheme copyWith({
    Color? railBg,
    double? railWidth,
    double? containerWidth,
    double? containerHeight,
    double? iconsGap,
    double? iconsPaddingTop,
    double? iconsMargin,
    BorderRadius? radius,
    Color? selectedItemBg,
  }) {
    return NavRailTheme(
      railBg: railBg ?? this.railBg,
      railWidth: railWidth ?? this.railWidth,
      containerWidth: containerWidth ?? this.containerWidth,
      containerHeight: containerHeight ?? this.containerHeight,
      betweenIconsGap: iconsGap ?? this.betweenIconsGap,
      iconsPaddingTop: iconsPaddingTop ?? this.iconsPaddingTop,
      iconsMargin: iconsMargin ?? this.iconsMargin,
      radius: radius ?? this.radius,
      selectedItemBg: selectedItemBg ?? this.selectedItemBg,
    );
  }

  @override
  NavRailTheme lerp(covariant ThemeExtension<NavRailTheme>? other, double t) {
    if (other is! NavRailTheme) return this as NavRailTheme;
    return NavRailTheme(
      railBg: Color.lerp(railBg, other.railBg, t)!,
      railWidth: t < 0.5 ? railWidth : other.railWidth,
      containerWidth: t < 0.5 ? containerWidth : other.containerWidth,
      containerHeight: t < 0.5 ? containerHeight : other.containerHeight,
      betweenIconsGap: t < 0.5 ? betweenIconsGap : other.betweenIconsGap,
      iconsPaddingTop: t < 0.5 ? iconsPaddingTop : other.iconsPaddingTop,
      iconsMargin: t < 0.5 ? iconsMargin : other.iconsMargin,
      radius: t < 0.5 ? radius : other.radius,
      selectedItemBg: Color.lerp(selectedItemBg, other.selectedItemBg, t)!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is NavRailTheme &&
            const DeepCollectionEquality().equals(railBg, other.railBg) &&
            const DeepCollectionEquality().equals(railWidth, other.railWidth) &&
            const DeepCollectionEquality().equals(
              containerWidth,
              other.containerWidth,
            ) &&
            const DeepCollectionEquality().equals(
              containerHeight,
              other.containerHeight,
            ) &&
            const DeepCollectionEquality().equals(
              betweenIconsGap,
              other.betweenIconsGap,
            ) &&
            const DeepCollectionEquality().equals(
              iconsPaddingTop,
              other.iconsPaddingTop,
            ) &&
            const DeepCollectionEquality().equals(
              iconsMargin,
              other.iconsMargin,
            ) &&
            const DeepCollectionEquality().equals(radius, other.radius) &&
            const DeepCollectionEquality().equals(
              selectedItemBg,
              other.selectedItemBg,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(railBg),
      const DeepCollectionEquality().hash(railWidth),
      const DeepCollectionEquality().hash(containerWidth),
      const DeepCollectionEquality().hash(containerHeight),
      const DeepCollectionEquality().hash(betweenIconsGap),
      const DeepCollectionEquality().hash(iconsPaddingTop),
      const DeepCollectionEquality().hash(iconsMargin),
      const DeepCollectionEquality().hash(radius),
      const DeepCollectionEquality().hash(selectedItemBg),
    );
  }
}

extension NavRailThemeBuildContextProps on BuildContext {
  NavRailTheme get navRailTheme => Theme.of(this).extension<NavRailTheme>()!;
  Color get railBg => navRailTheme.railBg;
  double get railWidth => navRailTheme.railWidth;
  double get containerWidth => navRailTheme.containerWidth;
  double get containerHeight => navRailTheme.containerHeight;
  double get iconsGap => navRailTheme.betweenIconsGap;
  double get iconsPaddingTop => navRailTheme.iconsPaddingTop;
  double get iconsMargin => navRailTheme.iconsMargin;
  BorderRadius get radius => navRailTheme.radius;
  Color get selectedItemBg => navRailTheme.selectedItemBg;
}
