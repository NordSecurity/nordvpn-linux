// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'loading_indicator_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$LoadingIndicatorThemeTailorMixin
    on ThemeExtension<LoadingIndicatorTheme> {
  Color get color;
  double get strokeWidth;

  @override
  LoadingIndicatorTheme copyWith({Color? color, double? strokeWidth}) {
    return LoadingIndicatorTheme(
      color: color ?? this.color,
      strokeWidth: strokeWidth ?? this.strokeWidth,
    );
  }

  @override
  LoadingIndicatorTheme lerp(
    covariant ThemeExtension<LoadingIndicatorTheme>? other,
    double t,
  ) {
    if (other is! LoadingIndicatorTheme) return this as LoadingIndicatorTheme;
    return LoadingIndicatorTheme(
      color: Color.lerp(color, other.color, t)!,
      strokeWidth: t < 0.5 ? strokeWidth : other.strokeWidth,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is LoadingIndicatorTheme &&
            const DeepCollectionEquality().equals(color, other.color) &&
            const DeepCollectionEquality().equals(
              strokeWidth,
              other.strokeWidth,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(color),
      const DeepCollectionEquality().hash(strokeWidth),
    );
  }
}

extension LoadingIndicatorThemeBuildContextProps on BuildContext {
  LoadingIndicatorTheme get loadingIndicatorTheme =>
      Theme.of(this).extension<LoadingIndicatorTheme>()!;
  Color get color => loadingIndicatorTheme.color;
  double get strokeWidth => loadingIndicatorTheme.strokeWidth;
}
