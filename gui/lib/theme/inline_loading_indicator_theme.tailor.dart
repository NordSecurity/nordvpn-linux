// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'inline_loading_indicator_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$InlineLoadingIndicatorThemeTailorMixin
    on ThemeExtension<InlineLoadingIndicatorTheme> {
  double get width;
  double get height;
  double get stroke;
  Color get color;
  Color get alternativeColor;

  @override
  InlineLoadingIndicatorTheme copyWith({
    double? width,
    double? height,
    double? stroke,
    Color? color,
    Color? alternativeColor,
  }) {
    return InlineLoadingIndicatorTheme(
      width: width ?? this.width,
      height: height ?? this.height,
      stroke: stroke ?? this.stroke,
      color: color ?? this.color,
      alternativeColor: alternativeColor ?? this.alternativeColor,
    );
  }

  @override
  InlineLoadingIndicatorTheme lerp(
    covariant ThemeExtension<InlineLoadingIndicatorTheme>? other,
    double t,
  ) {
    if (other is! InlineLoadingIndicatorTheme)
      return this as InlineLoadingIndicatorTheme;
    return InlineLoadingIndicatorTheme(
      width: t < 0.5 ? width : other.width,
      height: t < 0.5 ? height : other.height,
      stroke: t < 0.5 ? stroke : other.stroke,
      color: Color.lerp(color, other.color, t)!,
      alternativeColor: Color.lerp(
        alternativeColor,
        other.alternativeColor,
        t,
      )!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is InlineLoadingIndicatorTheme &&
            const DeepCollectionEquality().equals(width, other.width) &&
            const DeepCollectionEquality().equals(height, other.height) &&
            const DeepCollectionEquality().equals(stroke, other.stroke) &&
            const DeepCollectionEquality().equals(color, other.color) &&
            const DeepCollectionEquality().equals(
              alternativeColor,
              other.alternativeColor,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(width),
      const DeepCollectionEquality().hash(height),
      const DeepCollectionEquality().hash(stroke),
      const DeepCollectionEquality().hash(color),
      const DeepCollectionEquality().hash(alternativeColor),
    );
  }
}

extension InlineLoadingIndicatorThemeBuildContextProps on BuildContext {
  InlineLoadingIndicatorTheme get inlineLoadingIndicatorTheme =>
      Theme.of(this).extension<InlineLoadingIndicatorTheme>()!;
  double get width => inlineLoadingIndicatorTheme.width;
  double get height => inlineLoadingIndicatorTheme.height;
  double get stroke => inlineLoadingIndicatorTheme.stroke;
  Color get color => inlineLoadingIndicatorTheme.color;
  Color get alternativeColor => inlineLoadingIndicatorTheme.alternativeColor;
}
