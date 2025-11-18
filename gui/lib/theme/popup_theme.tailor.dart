// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'popup_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$PopupThemeTailorMixin on ThemeExtension<PopupTheme> {
  double get widgetWidth;
  BorderRadius get widgetRadius;
  double get contentAllPadding;
  double get xButtonAllPadding;
  double get gapBetweenElements;
  double get verticalElementSpacing;
  double get buttonHeight;
  double get singleButtonMinWidth;
  EdgeInsetsGeometry get buttonPadding;
  Color get primaryButtonBackgroundColor;
  Color get secondaryButtonBackgroundColor;
  TextStyle get textPrimary;
  TextStyle get textSecondary;

  @override
  PopupTheme copyWith({
    double? widgetWidth,
    BorderRadius? widgetRadius,
    double? contentAllPadding,
    double? xButtonAllPadding,
    double? gapBetweenElements,
    double? verticalElementSpacing,
    double? buttonHeight,
    double? singleButtonMinWidth,
    EdgeInsetsGeometry? buttonPadding,
    Color? primaryButtonBackgroundColor,
    Color? secondaryButtonBackgroundColor,
    TextStyle? textPrimary,
    TextStyle? textSecondary,
  }) {
    return PopupTheme(
      widgetWidth: widgetWidth ?? this.widgetWidth,
      widgetRadius: widgetRadius ?? this.widgetRadius,
      contentAllPadding: contentAllPadding ?? this.contentAllPadding,
      xButtonAllPadding: xButtonAllPadding ?? this.xButtonAllPadding,
      gapBetweenElements: gapBetweenElements ?? this.gapBetweenElements,
      verticalElementSpacing:
          verticalElementSpacing ?? this.verticalElementSpacing,
      buttonHeight: buttonHeight ?? this.buttonHeight,
      singleButtonMinWidth: singleButtonMinWidth ?? this.singleButtonMinWidth,
      buttonPadding: buttonPadding ?? this.buttonPadding,
      primaryButtonBackgroundColor:
          primaryButtonBackgroundColor ?? this.primaryButtonBackgroundColor,
      secondaryButtonBackgroundColor:
          secondaryButtonBackgroundColor ?? this.secondaryButtonBackgroundColor,
      textPrimary: textPrimary ?? this.textPrimary,
      textSecondary: textSecondary ?? this.textSecondary,
    );
  }

  @override
  PopupTheme lerp(covariant ThemeExtension<PopupTheme>? other, double t) {
    if (other is! PopupTheme) return this as PopupTheme;
    return PopupTheme(
      widgetWidth: t < 0.5 ? widgetWidth : other.widgetWidth,
      widgetRadius: t < 0.5 ? widgetRadius : other.widgetRadius,
      contentAllPadding: t < 0.5 ? contentAllPadding : other.contentAllPadding,
      xButtonAllPadding: t < 0.5 ? xButtonAllPadding : other.xButtonAllPadding,
      gapBetweenElements: t < 0.5
          ? gapBetweenElements
          : other.gapBetweenElements,
      verticalElementSpacing: t < 0.5
          ? verticalElementSpacing
          : other.verticalElementSpacing,
      buttonHeight: t < 0.5 ? buttonHeight : other.buttonHeight,
      singleButtonMinWidth: t < 0.5
          ? singleButtonMinWidth
          : other.singleButtonMinWidth,
      buttonPadding: t < 0.5 ? buttonPadding : other.buttonPadding,
      primaryButtonBackgroundColor: Color.lerp(
        primaryButtonBackgroundColor,
        other.primaryButtonBackgroundColor,
        t,
      )!,
      secondaryButtonBackgroundColor: Color.lerp(
        secondaryButtonBackgroundColor,
        other.secondaryButtonBackgroundColor,
        t,
      )!,
      textPrimary: TextStyle.lerp(textPrimary, other.textPrimary, t)!,
      textSecondary: TextStyle.lerp(textSecondary, other.textSecondary, t)!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is PopupTheme &&
            const DeepCollectionEquality().equals(
              widgetWidth,
              other.widgetWidth,
            ) &&
            const DeepCollectionEquality().equals(
              widgetRadius,
              other.widgetRadius,
            ) &&
            const DeepCollectionEquality().equals(
              contentAllPadding,
              other.contentAllPadding,
            ) &&
            const DeepCollectionEquality().equals(
              xButtonAllPadding,
              other.xButtonAllPadding,
            ) &&
            const DeepCollectionEquality().equals(
              gapBetweenElements,
              other.gapBetweenElements,
            ) &&
            const DeepCollectionEquality().equals(
              verticalElementSpacing,
              other.verticalElementSpacing,
            ) &&
            const DeepCollectionEquality().equals(
              buttonHeight,
              other.buttonHeight,
            ) &&
            const DeepCollectionEquality().equals(
              singleButtonMinWidth,
              other.singleButtonMinWidth,
            ) &&
            const DeepCollectionEquality().equals(
              buttonPadding,
              other.buttonPadding,
            ) &&
            const DeepCollectionEquality().equals(
              primaryButtonBackgroundColor,
              other.primaryButtonBackgroundColor,
            ) &&
            const DeepCollectionEquality().equals(
              secondaryButtonBackgroundColor,
              other.secondaryButtonBackgroundColor,
            ) &&
            const DeepCollectionEquality().equals(
              textPrimary,
              other.textPrimary,
            ) &&
            const DeepCollectionEquality().equals(
              textSecondary,
              other.textSecondary,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(widgetWidth),
      const DeepCollectionEquality().hash(widgetRadius),
      const DeepCollectionEquality().hash(contentAllPadding),
      const DeepCollectionEquality().hash(xButtonAllPadding),
      const DeepCollectionEquality().hash(gapBetweenElements),
      const DeepCollectionEquality().hash(verticalElementSpacing),
      const DeepCollectionEquality().hash(buttonHeight),
      const DeepCollectionEquality().hash(singleButtonMinWidth),
      const DeepCollectionEquality().hash(buttonPadding),
      const DeepCollectionEquality().hash(primaryButtonBackgroundColor),
      const DeepCollectionEquality().hash(secondaryButtonBackgroundColor),
      const DeepCollectionEquality().hash(textPrimary),
      const DeepCollectionEquality().hash(textSecondary),
    );
  }
}

extension PopupThemeBuildContextProps on BuildContext {
  PopupTheme get popupTheme => Theme.of(this).extension<PopupTheme>()!;
  double get widgetWidth => popupTheme.widgetWidth;
  BorderRadius get widgetRadius => popupTheme.widgetRadius;
  double get contentAllPadding => popupTheme.contentAllPadding;
  double get xButtonAllPadding => popupTheme.xButtonAllPadding;
  double get gapBetweenElements => popupTheme.gapBetweenElements;
  double get verticalElementSpacing => popupTheme.verticalElementSpacing;
  double get buttonHeight => popupTheme.buttonHeight;
  double get singleButtonMinWidth => popupTheme.singleButtonMinWidth;
  EdgeInsetsGeometry get buttonPadding => popupTheme.buttonPadding;
  Color get primaryButtonBackgroundColor =>
      popupTheme.primaryButtonBackgroundColor;
  Color get secondaryButtonBackgroundColor =>
      popupTheme.secondaryButtonBackgroundColor;
  TextStyle get textPrimary => popupTheme.textPrimary;
  TextStyle get textSecondary => popupTheme.textSecondary;
}
