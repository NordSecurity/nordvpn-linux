// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'toast_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$ToastThemeTailorMixin on ThemeExtension<ToastTheme> {
  TextStyle get toastMessageTextStyle;
  Color get toastBackgroundColor;
  double get toastSpacing;
  BorderRadius get toastBorderRadius;
  double get widgetWidth;
  double get widgetHeight;
  EdgeInsets get toastCloseButtonPadding;
  double get toastBorderWidth;
  Color get toastBorderColor;

  @override
  ToastTheme copyWith({
    TextStyle? toastMessageTextStyle,
    Color? toastBackgroundColor,
    double? toastSpacing,
    BorderRadius? toastBorderRadius,
    double? widgetWidth,
    double? widgetHeight,
    EdgeInsets? toastCloseButtonPadding,
    double? toastBorderWidth,
    Color? toastBorderColor,
  }) {
    return ToastTheme(
      toastMessageTextStyle:
          toastMessageTextStyle ?? this.toastMessageTextStyle,
      toastBackgroundColor: toastBackgroundColor ?? this.toastBackgroundColor,
      toastSpacing: toastSpacing ?? this.toastSpacing,
      toastBorderRadius: toastBorderRadius ?? this.toastBorderRadius,
      widgetWidth: widgetWidth ?? this.widgetWidth,
      widgetHeight: widgetHeight ?? this.widgetHeight,
      toastCloseButtonPadding:
          toastCloseButtonPadding ?? this.toastCloseButtonPadding,
      toastBorderWidth: toastBorderWidth ?? this.toastBorderWidth,
      toastBorderColor: toastBorderColor ?? this.toastBorderColor,
    );
  }

  @override
  ToastTheme lerp(covariant ThemeExtension<ToastTheme>? other, double t) {
    if (other is! ToastTheme) return this as ToastTheme;
    return ToastTheme(
      toastMessageTextStyle: TextStyle.lerp(
        toastMessageTextStyle,
        other.toastMessageTextStyle,
        t,
      )!,
      toastBackgroundColor: Color.lerp(
        toastBackgroundColor,
        other.toastBackgroundColor,
        t,
      )!,
      toastSpacing: t < 0.5 ? toastSpacing : other.toastSpacing,
      toastBorderRadius: t < 0.5 ? toastBorderRadius : other.toastBorderRadius,
      widgetWidth: t < 0.5 ? widgetWidth : other.widgetWidth,
      widgetHeight: t < 0.5 ? widgetHeight : other.widgetHeight,
      toastCloseButtonPadding: t < 0.5
          ? toastCloseButtonPadding
          : other.toastCloseButtonPadding,
      toastBorderWidth: t < 0.5 ? toastBorderWidth : other.toastBorderWidth,
      toastBorderColor: Color.lerp(
        toastBorderColor,
        other.toastBorderColor,
        t,
      )!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is ToastTheme &&
            const DeepCollectionEquality().equals(
              toastMessageTextStyle,
              other.toastMessageTextStyle,
            ) &&
            const DeepCollectionEquality().equals(
              toastBackgroundColor,
              other.toastBackgroundColor,
            ) &&
            const DeepCollectionEquality().equals(
              toastSpacing,
              other.toastSpacing,
            ) &&
            const DeepCollectionEquality().equals(
              toastBorderRadius,
              other.toastBorderRadius,
            ) &&
            const DeepCollectionEquality().equals(
              widgetWidth,
              other.widgetWidth,
            ) &&
            const DeepCollectionEquality().equals(
              widgetHeight,
              other.widgetHeight,
            ) &&
            const DeepCollectionEquality().equals(
              toastCloseButtonPadding,
              other.toastCloseButtonPadding,
            ) &&
            const DeepCollectionEquality().equals(
              toastBorderWidth,
              other.toastBorderWidth,
            ) &&
            const DeepCollectionEquality().equals(
              toastBorderColor,
              other.toastBorderColor,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(toastMessageTextStyle),
      const DeepCollectionEquality().hash(toastBackgroundColor),
      const DeepCollectionEquality().hash(toastSpacing),
      const DeepCollectionEquality().hash(toastBorderRadius),
      const DeepCollectionEquality().hash(widgetWidth),
      const DeepCollectionEquality().hash(widgetHeight),
      const DeepCollectionEquality().hash(toastCloseButtonPadding),
      const DeepCollectionEquality().hash(toastBorderWidth),
      const DeepCollectionEquality().hash(toastBorderColor),
    );
  }
}

extension ToastThemeBuildContextProps on BuildContext {
  ToastTheme get toastTheme => Theme.of(this).extension<ToastTheme>()!;
  TextStyle get toastMessageTextStyle => toastTheme.toastMessageTextStyle;
  Color get toastBackgroundColor => toastTheme.toastBackgroundColor;
  double get toastSpacing => toastTheme.toastSpacing;
  BorderRadius get toastBorderRadius => toastTheme.toastBorderRadius;
  double get widgetWidth => toastTheme.widgetWidth;
  double get widgetHeight => toastTheme.widgetHeight;
  EdgeInsets get toastCloseButtonPadding => toastTheme.toastCloseButtonPadding;
  double get toastBorderWidth => toastTheme.toastBorderWidth;
  Color get toastBorderColor => toastTheme.toastBorderColor;
}
