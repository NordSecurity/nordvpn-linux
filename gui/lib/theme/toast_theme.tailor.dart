// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'toast_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$ToastThemeTailorMixin on ThemeExtension<ToastTheme> {
  TextStyle get messageTextStyle;
  Color get backgroundColor;
  double get spacing;
  BorderRadius get borderRadius;
  double get widgetWidth;
  double get widgetHeight;
  double get widgetPositionRight;
  double get widgetPositionBottom;
  EdgeInsets get closeButtonPadding;
  double get borderWidth;
  Color get borderColor;

  @override
  ToastTheme copyWith({
    TextStyle? messageTextStyle,
    Color? backgroundColor,
    double? spacing,
    BorderRadius? borderRadius,
    double? widgetWidth,
    double? widgetHeight,
    double? widgetPositionRight,
    double? widgetPositionBottom,
    EdgeInsets? closeButtonPadding,
    double? borderWidth,
    Color? borderColor,
  }) {
    return ToastTheme(
      messageTextStyle: messageTextStyle ?? this.messageTextStyle,
      backgroundColor: backgroundColor ?? this.backgroundColor,
      spacing: spacing ?? this.spacing,
      borderRadius: borderRadius ?? this.borderRadius,
      widgetWidth: widgetWidth ?? this.widgetWidth,
      widgetHeight: widgetHeight ?? this.widgetHeight,
      widgetPositionRight: widgetPositionRight ?? this.widgetPositionRight,
      widgetPositionBottom: widgetPositionBottom ?? this.widgetPositionBottom,
      closeButtonPadding: closeButtonPadding ?? this.closeButtonPadding,
      borderWidth: borderWidth ?? this.borderWidth,
      borderColor: borderColor ?? this.borderColor,
    );
  }

  @override
  ToastTheme lerp(covariant ThemeExtension<ToastTheme>? other, double t) {
    if (other is! ToastTheme) return this as ToastTheme;
    return ToastTheme(
      messageTextStyle: TextStyle.lerp(
        messageTextStyle,
        other.messageTextStyle,
        t,
      )!,
      backgroundColor: Color.lerp(backgroundColor, other.backgroundColor, t)!,
      spacing: t < 0.5 ? spacing : other.spacing,
      borderRadius: t < 0.5 ? borderRadius : other.borderRadius,
      widgetWidth: t < 0.5 ? widgetWidth : other.widgetWidth,
      widgetHeight: t < 0.5 ? widgetHeight : other.widgetHeight,
      widgetPositionRight: t < 0.5
          ? widgetPositionRight
          : other.widgetPositionRight,
      widgetPositionBottom: t < 0.5
          ? widgetPositionBottom
          : other.widgetPositionBottom,
      closeButtonPadding: t < 0.5
          ? closeButtonPadding
          : other.closeButtonPadding,
      borderWidth: t < 0.5 ? borderWidth : other.borderWidth,
      borderColor: Color.lerp(borderColor, other.borderColor, t)!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is ToastTheme &&
            const DeepCollectionEquality().equals(
              messageTextStyle,
              other.messageTextStyle,
            ) &&
            const DeepCollectionEquality().equals(
              backgroundColor,
              other.backgroundColor,
            ) &&
            const DeepCollectionEquality().equals(spacing, other.spacing) &&
            const DeepCollectionEquality().equals(
              borderRadius,
              other.borderRadius,
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
              widgetPositionRight,
              other.widgetPositionRight,
            ) &&
            const DeepCollectionEquality().equals(
              widgetPositionBottom,
              other.widgetPositionBottom,
            ) &&
            const DeepCollectionEquality().equals(
              closeButtonPadding,
              other.closeButtonPadding,
            ) &&
            const DeepCollectionEquality().equals(
              borderWidth,
              other.borderWidth,
            ) &&
            const DeepCollectionEquality().equals(
              borderColor,
              other.borderColor,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(messageTextStyle),
      const DeepCollectionEquality().hash(backgroundColor),
      const DeepCollectionEquality().hash(spacing),
      const DeepCollectionEquality().hash(borderRadius),
      const DeepCollectionEquality().hash(widgetWidth),
      const DeepCollectionEquality().hash(widgetHeight),
      const DeepCollectionEquality().hash(widgetPositionRight),
      const DeepCollectionEquality().hash(widgetPositionBottom),
      const DeepCollectionEquality().hash(closeButtonPadding),
      const DeepCollectionEquality().hash(borderWidth),
      const DeepCollectionEquality().hash(borderColor),
    );
  }
}

extension ToastThemeBuildContextProps on BuildContext {
  ToastTheme get toastTheme => Theme.of(this).extension<ToastTheme>()!;
  TextStyle get messageTextStyle => toastTheme.messageTextStyle;
  Color get backgroundColor => toastTheme.backgroundColor;
  double get spacing => toastTheme.spacing;
  BorderRadius get borderRadius => toastTheme.borderRadius;
  double get widgetWidth => toastTheme.widgetWidth;
  double get widgetHeight => toastTheme.widgetHeight;
  double get widgetPositionRight => toastTheme.widgetPositionRight;
  double get widgetPositionBottom => toastTheme.widgetPositionBottom;
  EdgeInsets get closeButtonPadding => toastTheme.closeButtonPadding;
  double get borderWidth => toastTheme.borderWidth;
  Color get borderColor => toastTheme.borderColor;
}
