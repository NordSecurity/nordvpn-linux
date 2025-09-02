// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'interactive_list_view_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$InteractiveListViewThemeTailorMixin
    on ThemeExtension<InteractiveListViewTheme> {
  double get borderRadius;
  Color get borderColor;
  Color get focusBorderColor;
  double get borderWidth;

  @override
  InteractiveListViewTheme copyWith({
    double? borderRadius,
    Color? borderColor,
    Color? focusBorderColor,
    double? borderWidth,
  }) {
    return InteractiveListViewTheme(
      borderRadius: borderRadius ?? this.borderRadius,
      borderColor: borderColor ?? this.borderColor,
      focusBorderColor: focusBorderColor ?? this.focusBorderColor,
      borderWidth: borderWidth ?? this.borderWidth,
    );
  }

  @override
  InteractiveListViewTheme lerp(
    covariant ThemeExtension<InteractiveListViewTheme>? other,
    double t,
  ) {
    if (other is! InteractiveListViewTheme)
      return this as InteractiveListViewTheme;
    return InteractiveListViewTheme(
      borderRadius: t < 0.5 ? borderRadius : other.borderRadius,
      borderColor: Color.lerp(borderColor, other.borderColor, t)!,
      focusBorderColor: Color.lerp(
        focusBorderColor,
        other.focusBorderColor,
        t,
      )!,
      borderWidth: t < 0.5 ? borderWidth : other.borderWidth,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is InteractiveListViewTheme &&
            const DeepCollectionEquality().equals(
              borderRadius,
              other.borderRadius,
            ) &&
            const DeepCollectionEquality().equals(
              borderColor,
              other.borderColor,
            ) &&
            const DeepCollectionEquality().equals(
              focusBorderColor,
              other.focusBorderColor,
            ) &&
            const DeepCollectionEquality().equals(
              borderWidth,
              other.borderWidth,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(borderRadius),
      const DeepCollectionEquality().hash(borderColor),
      const DeepCollectionEquality().hash(focusBorderColor),
      const DeepCollectionEquality().hash(borderWidth),
    );
  }
}

extension InteractiveListViewThemeBuildContextProps on BuildContext {
  InteractiveListViewTheme get interactiveListViewTheme =>
      Theme.of(this).extension<InteractiveListViewTheme>()!;
  double get borderRadius => interactiveListViewTheme.borderRadius;
  Color get borderColor => interactiveListViewTheme.borderColor;
  Color get focusBorderColor => interactiveListViewTheme.focusBorderColor;
  double get borderWidth => interactiveListViewTheme.borderWidth;
}
