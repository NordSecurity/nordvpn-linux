// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'consent_screen_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$ConsentScreenThemeTailorMixin on ThemeExtension<ConsentScreenTheme> {
  Color get overlayColor;
  TextStyle get bodyTextStyle;
  TextStyle get titleTextStyle;
  TextStyle get titleBarTextStyle;
  double get width;
  double get height;
  double get padding;
  TextStyle get listItemTitle;
  TextStyle get listItemSubtitle;
  double get titleBarWidth;

  @override
  ConsentScreenTheme copyWith({
    Color? overlayColor,
    TextStyle? bodyTextStyle,
    TextStyle? titleTextStyle,
    TextStyle? titleBarTextStyle,
    double? width,
    double? height,
    double? padding,
    TextStyle? listItemTitle,
    TextStyle? listItemSubtitle,
    double? titleBarWidth,
  }) {
    return ConsentScreenTheme(
      overlayColor: overlayColor ?? this.overlayColor,
      bodyTextStyle: bodyTextStyle ?? this.bodyTextStyle,
      titleTextStyle: titleTextStyle ?? this.titleTextStyle,
      titleBarTextStyle: titleBarTextStyle ?? this.titleBarTextStyle,
      width: width ?? this.width,
      height: height ?? this.height,
      padding: padding ?? this.padding,
      listItemTitle: listItemTitle ?? this.listItemTitle,
      listItemSubtitle: listItemSubtitle ?? this.listItemSubtitle,
      titleBarWidth: titleBarWidth ?? this.titleBarWidth,
    );
  }

  @override
  ConsentScreenTheme lerp(
    covariant ThemeExtension<ConsentScreenTheme>? other,
    double t,
  ) {
    if (other is! ConsentScreenTheme) return this as ConsentScreenTheme;
    return ConsentScreenTheme(
      overlayColor: Color.lerp(overlayColor, other.overlayColor, t)!,
      bodyTextStyle: TextStyle.lerp(bodyTextStyle, other.bodyTextStyle, t)!,
      titleTextStyle: TextStyle.lerp(titleTextStyle, other.titleTextStyle, t)!,
      titleBarTextStyle: TextStyle.lerp(
        titleBarTextStyle,
        other.titleBarTextStyle,
        t,
      )!,
      width: t < 0.5 ? width : other.width,
      height: t < 0.5 ? height : other.height,
      padding: t < 0.5 ? padding : other.padding,
      listItemTitle: TextStyle.lerp(listItemTitle, other.listItemTitle, t)!,
      listItemSubtitle: TextStyle.lerp(
        listItemSubtitle,
        other.listItemSubtitle,
        t,
      )!,
      titleBarWidth: t < 0.5 ? titleBarWidth : other.titleBarWidth,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is ConsentScreenTheme &&
            const DeepCollectionEquality().equals(
              overlayColor,
              other.overlayColor,
            ) &&
            const DeepCollectionEquality().equals(
              bodyTextStyle,
              other.bodyTextStyle,
            ) &&
            const DeepCollectionEquality().equals(
              titleTextStyle,
              other.titleTextStyle,
            ) &&
            const DeepCollectionEquality().equals(
              titleBarTextStyle,
              other.titleBarTextStyle,
            ) &&
            const DeepCollectionEquality().equals(width, other.width) &&
            const DeepCollectionEquality().equals(height, other.height) &&
            const DeepCollectionEquality().equals(padding, other.padding) &&
            const DeepCollectionEquality().equals(
              listItemTitle,
              other.listItemTitle,
            ) &&
            const DeepCollectionEquality().equals(
              listItemSubtitle,
              other.listItemSubtitle,
            ) &&
            const DeepCollectionEquality().equals(
              titleBarWidth,
              other.titleBarWidth,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(overlayColor),
      const DeepCollectionEquality().hash(bodyTextStyle),
      const DeepCollectionEquality().hash(titleTextStyle),
      const DeepCollectionEquality().hash(titleBarTextStyle),
      const DeepCollectionEquality().hash(width),
      const DeepCollectionEquality().hash(height),
      const DeepCollectionEquality().hash(padding),
      const DeepCollectionEquality().hash(listItemTitle),
      const DeepCollectionEquality().hash(listItemSubtitle),
      const DeepCollectionEquality().hash(titleBarWidth),
    );
  }
}

extension ConsentScreenThemeBuildContextProps on BuildContext {
  ConsentScreenTheme get consentScreenTheme =>
      Theme.of(this).extension<ConsentScreenTheme>()!;
  Color get overlayColor => consentScreenTheme.overlayColor;
  TextStyle get bodyTextStyle => consentScreenTheme.bodyTextStyle;
  TextStyle get titleTextStyle => consentScreenTheme.titleTextStyle;
  TextStyle get titleBarTextStyle => consentScreenTheme.titleBarTextStyle;
  double get width => consentScreenTheme.width;
  double get height => consentScreenTheme.height;
  double get padding => consentScreenTheme.padding;
  TextStyle get listItemTitle => consentScreenTheme.listItemTitle;
  TextStyle get listItemSubtitle => consentScreenTheme.listItemSubtitle;
  double get titleBarWidth => consentScreenTheme.titleBarWidth;
}
