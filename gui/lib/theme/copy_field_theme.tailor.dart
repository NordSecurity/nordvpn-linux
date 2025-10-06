// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'copy_field_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$CopyFieldThemeTailorMixin on ThemeExtension<CopyFieldTheme> {
  double get borderRadius;
  TextStyle get commandTextStyle;
  TextStyle get descriptionTextStyle;

  @override
  CopyFieldTheme copyWith({
    double? borderRadius,
    TextStyle? commandTextStyle,
    TextStyle? descriptionTextStyle,
  }) {
    return CopyFieldTheme(
      borderRadius: borderRadius ?? this.borderRadius,
      commandTextStyle: commandTextStyle ?? this.commandTextStyle,
      descriptionTextStyle: descriptionTextStyle ?? this.descriptionTextStyle,
    );
  }

  @override
  CopyFieldTheme lerp(
    covariant ThemeExtension<CopyFieldTheme>? other,
    double t,
  ) {
    if (other is! CopyFieldTheme) return this as CopyFieldTheme;
    return CopyFieldTheme(
      borderRadius: t < 0.5 ? borderRadius : other.borderRadius,
      commandTextStyle: TextStyle.lerp(
        commandTextStyle,
        other.commandTextStyle,
        t,
      )!,
      descriptionTextStyle: TextStyle.lerp(
        descriptionTextStyle,
        other.descriptionTextStyle,
        t,
      )!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is CopyFieldTheme &&
            const DeepCollectionEquality().equals(
              borderRadius,
              other.borderRadius,
            ) &&
            const DeepCollectionEquality().equals(
              commandTextStyle,
              other.commandTextStyle,
            ) &&
            const DeepCollectionEquality().equals(
              descriptionTextStyle,
              other.descriptionTextStyle,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(borderRadius),
      const DeepCollectionEquality().hash(commandTextStyle),
      const DeepCollectionEquality().hash(descriptionTextStyle),
    );
  }
}

extension CopyFieldThemeBuildContextProps on BuildContext {
  CopyFieldTheme get copyFieldTheme =>
      Theme.of(this).extension<CopyFieldTheme>()!;
  double get borderRadius => copyFieldTheme.borderRadius;
  TextStyle get commandTextStyle => copyFieldTheme.commandTextStyle;
  TextStyle get descriptionTextStyle => copyFieldTheme.descriptionTextStyle;
}
