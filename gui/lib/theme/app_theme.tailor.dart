// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'app_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$AppThemeTailorMixin on ThemeExtension<AppTheme> {
  double get borderRadiusLarge;
  double get borderRadiusMedium;
  double get borderRadiusSmall;
  double get padding;
  double get margin;
  double get outerPadding;
  Color get borderColor;
  double get verticalSpaceSmall;
  double get verticalSpaceMedium;
  double get verticalSpaceLarge;
  double get horizontalSpaceSmall;
  double get horizontalSpace;
  Color get textErrorColor;
  Color get successColor;
  double get flagsBorderSize;
  Color get overlayBackgroundColor;
  double get trailingIconSize;
  Color get backgroundColor;
  Color get areaBackgroundColor;
  Color get dividerColor;
  double get disabledOpacity;
  TextStyle get captionStrong;
  TextStyle get caption;
  TextStyle get captionRegularGray171;
  TextStyle get bodyStrong;
  TextStyle get body;
  TextStyle get subtitleStrong;
  TextStyle get linkButton;
  TextStyle get title;
  TextStyle get linkNormal;
  TextStyle get linkSmall;
  TextStyle get textDisabled;
  Color get area;

  @override
  AppTheme copyWith({
    double? borderRadiusLarge,
    double? borderRadiusMedium,
    double? borderRadiusSmall,
    double? padding,
    double? margin,
    double? outerPadding,
    Color? borderColor,
    double? verticalSpaceSmall,
    double? verticalSpaceMedium,
    double? verticalSpaceLarge,
    double? horizontalSpaceSmall,
    double? horizontalSpace,
    Color? textErrorColor,
    Color? successColor,
    double? flagsBorderSize,
    Color? overlayBackgroundColor,
    double? trailingIconSize,
    Color? backgroundColor,
    Color? areaBackgroundColor,
    Color? dividerColor,
    double? disabledOpacity,
    TextStyle? captionStrong,
    TextStyle? caption,
    TextStyle? captionRegularGray171,
    TextStyle? bodyStrong,
    TextStyle? body,
    TextStyle? subtitleStrong,
    TextStyle? linkButton,
    TextStyle? title,
    TextStyle? linkNormal,
    TextStyle? linkSmall,
    TextStyle? textDisabled,
    Color? area,
  }) {
    return AppTheme(
      borderRadiusLarge: borderRadiusLarge ?? this.borderRadiusLarge,
      borderRadiusMedium: borderRadiusMedium ?? this.borderRadiusMedium,
      borderRadiusSmall: borderRadiusSmall ?? this.borderRadiusSmall,
      padding: padding ?? this.padding,
      margin: margin ?? this.margin,
      outerPadding: outerPadding ?? this.outerPadding,
      borderColor: borderColor ?? this.borderColor,
      verticalSpaceSmall: verticalSpaceSmall ?? this.verticalSpaceSmall,
      verticalSpaceMedium: verticalSpaceMedium ?? this.verticalSpaceMedium,
      verticalSpaceLarge: verticalSpaceLarge ?? this.verticalSpaceLarge,
      horizontalSpaceSmall: horizontalSpaceSmall ?? this.horizontalSpaceSmall,
      horizontalSpace: horizontalSpace ?? this.horizontalSpace,
      textErrorColor: textErrorColor ?? this.textErrorColor,
      successColor: successColor ?? this.successColor,
      flagsBorderSize: flagsBorderSize ?? this.flagsBorderSize,
      overlayBackgroundColor:
          overlayBackgroundColor ?? this.overlayBackgroundColor,
      trailingIconSize: trailingIconSize ?? this.trailingIconSize,
      backgroundColor: backgroundColor ?? this.backgroundColor,
      areaBackgroundColor: areaBackgroundColor ?? this.areaBackgroundColor,
      dividerColor: dividerColor ?? this.dividerColor,
      disabledOpacity: disabledOpacity ?? this.disabledOpacity,
      captionStrong: captionStrong ?? this.captionStrong,
      caption: caption ?? this.caption,
      captionRegularGray171:
          captionRegularGray171 ?? this.captionRegularGray171,
      bodyStrong: bodyStrong ?? this.bodyStrong,
      body: body ?? this.body,
      subtitleStrong: subtitleStrong ?? this.subtitleStrong,
      linkButton: linkButton ?? this.linkButton,
      title: title ?? this.title,
      linkNormal: linkNormal ?? this.linkNormal,
      linkSmall: linkSmall ?? this.linkSmall,
      textDisabled: textDisabled ?? this.textDisabled,
      area: area ?? this.area,
    );
  }

  @override
  AppTheme lerp(covariant ThemeExtension<AppTheme>? other, double t) {
    if (other is! AppTheme) return this as AppTheme;
    return AppTheme(
      borderRadiusLarge: t < 0.5 ? borderRadiusLarge : other.borderRadiusLarge,
      borderRadiusMedium: t < 0.5
          ? borderRadiusMedium
          : other.borderRadiusMedium,
      borderRadiusSmall: t < 0.5 ? borderRadiusSmall : other.borderRadiusSmall,
      padding: t < 0.5 ? padding : other.padding,
      margin: t < 0.5 ? margin : other.margin,
      outerPadding: t < 0.5 ? outerPadding : other.outerPadding,
      borderColor: Color.lerp(borderColor, other.borderColor, t)!,
      verticalSpaceSmall: t < 0.5
          ? verticalSpaceSmall
          : other.verticalSpaceSmall,
      verticalSpaceMedium: t < 0.5
          ? verticalSpaceMedium
          : other.verticalSpaceMedium,
      verticalSpaceLarge: t < 0.5
          ? verticalSpaceLarge
          : other.verticalSpaceLarge,
      horizontalSpaceSmall: t < 0.5
          ? horizontalSpaceSmall
          : other.horizontalSpaceSmall,
      horizontalSpace: t < 0.5 ? horizontalSpace : other.horizontalSpace,
      textErrorColor: Color.lerp(textErrorColor, other.textErrorColor, t)!,
      successColor: Color.lerp(successColor, other.successColor, t)!,
      flagsBorderSize: t < 0.5 ? flagsBorderSize : other.flagsBorderSize,
      overlayBackgroundColor: Color.lerp(
        overlayBackgroundColor,
        other.overlayBackgroundColor,
        t,
      )!,
      trailingIconSize: t < 0.5 ? trailingIconSize : other.trailingIconSize,
      backgroundColor: Color.lerp(backgroundColor, other.backgroundColor, t)!,
      areaBackgroundColor: Color.lerp(
        areaBackgroundColor,
        other.areaBackgroundColor,
        t,
      )!,
      dividerColor: Color.lerp(dividerColor, other.dividerColor, t)!,
      disabledOpacity: t < 0.5 ? disabledOpacity : other.disabledOpacity,
      captionStrong: TextStyle.lerp(captionStrong, other.captionStrong, t)!,
      caption: TextStyle.lerp(caption, other.caption, t)!,
      captionRegularGray171: TextStyle.lerp(
        captionRegularGray171,
        other.captionRegularGray171,
        t,
      )!,
      bodyStrong: TextStyle.lerp(bodyStrong, other.bodyStrong, t)!,
      body: TextStyle.lerp(body, other.body, t)!,
      subtitleStrong: TextStyle.lerp(subtitleStrong, other.subtitleStrong, t)!,
      linkButton: TextStyle.lerp(linkButton, other.linkButton, t)!,
      title: TextStyle.lerp(title, other.title, t)!,
      linkNormal: TextStyle.lerp(linkNormal, other.linkNormal, t)!,
      linkSmall: TextStyle.lerp(linkSmall, other.linkSmall, t)!,
      textDisabled: TextStyle.lerp(textDisabled, other.textDisabled, t)!,
      area: Color.lerp(area, other.area, t)!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is AppTheme &&
            const DeepCollectionEquality().equals(
              borderRadiusLarge,
              other.borderRadiusLarge,
            ) &&
            const DeepCollectionEquality().equals(
              borderRadiusMedium,
              other.borderRadiusMedium,
            ) &&
            const DeepCollectionEquality().equals(
              borderRadiusSmall,
              other.borderRadiusSmall,
            ) &&
            const DeepCollectionEquality().equals(padding, other.padding) &&
            const DeepCollectionEquality().equals(margin, other.margin) &&
            const DeepCollectionEquality().equals(
              outerPadding,
              other.outerPadding,
            ) &&
            const DeepCollectionEquality().equals(
              borderColor,
              other.borderColor,
            ) &&
            const DeepCollectionEquality().equals(
              verticalSpaceSmall,
              other.verticalSpaceSmall,
            ) &&
            const DeepCollectionEquality().equals(
              verticalSpaceMedium,
              other.verticalSpaceMedium,
            ) &&
            const DeepCollectionEquality().equals(
              verticalSpaceLarge,
              other.verticalSpaceLarge,
            ) &&
            const DeepCollectionEquality().equals(
              horizontalSpaceSmall,
              other.horizontalSpaceSmall,
            ) &&
            const DeepCollectionEquality().equals(
              horizontalSpace,
              other.horizontalSpace,
            ) &&
            const DeepCollectionEquality().equals(
              textErrorColor,
              other.textErrorColor,
            ) &&
            const DeepCollectionEquality().equals(
              successColor,
              other.successColor,
            ) &&
            const DeepCollectionEquality().equals(
              flagsBorderSize,
              other.flagsBorderSize,
            ) &&
            const DeepCollectionEquality().equals(
              overlayBackgroundColor,
              other.overlayBackgroundColor,
            ) &&
            const DeepCollectionEquality().equals(
              trailingIconSize,
              other.trailingIconSize,
            ) &&
            const DeepCollectionEquality().equals(
              backgroundColor,
              other.backgroundColor,
            ) &&
            const DeepCollectionEquality().equals(
              areaBackgroundColor,
              other.areaBackgroundColor,
            ) &&
            const DeepCollectionEquality().equals(
              dividerColor,
              other.dividerColor,
            ) &&
            const DeepCollectionEquality().equals(
              disabledOpacity,
              other.disabledOpacity,
            ) &&
            const DeepCollectionEquality().equals(
              captionStrong,
              other.captionStrong,
            ) &&
            const DeepCollectionEquality().equals(caption, other.caption) &&
            const DeepCollectionEquality().equals(
              captionRegularGray171,
              other.captionRegularGray171,
            ) &&
            const DeepCollectionEquality().equals(
              bodyStrong,
              other.bodyStrong,
            ) &&
            const DeepCollectionEquality().equals(body, other.body) &&
            const DeepCollectionEquality().equals(
              subtitleStrong,
              other.subtitleStrong,
            ) &&
            const DeepCollectionEquality().equals(
              linkButton,
              other.linkButton,
            ) &&
            const DeepCollectionEquality().equals(title, other.title) &&
            const DeepCollectionEquality().equals(
              linkNormal,
              other.linkNormal,
            ) &&
            const DeepCollectionEquality().equals(linkSmall, other.linkSmall) &&
            const DeepCollectionEquality().equals(
              textDisabled,
              other.textDisabled,
            ) &&
            const DeepCollectionEquality().equals(area, other.area));
  }

  @override
  int get hashCode {
    return Object.hashAll([
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(borderRadiusLarge),
      const DeepCollectionEquality().hash(borderRadiusMedium),
      const DeepCollectionEquality().hash(borderRadiusSmall),
      const DeepCollectionEquality().hash(padding),
      const DeepCollectionEquality().hash(margin),
      const DeepCollectionEquality().hash(outerPadding),
      const DeepCollectionEquality().hash(borderColor),
      const DeepCollectionEquality().hash(verticalSpaceSmall),
      const DeepCollectionEquality().hash(verticalSpaceMedium),
      const DeepCollectionEquality().hash(verticalSpaceLarge),
      const DeepCollectionEquality().hash(horizontalSpaceSmall),
      const DeepCollectionEquality().hash(horizontalSpace),
      const DeepCollectionEquality().hash(textErrorColor),
      const DeepCollectionEquality().hash(successColor),
      const DeepCollectionEquality().hash(flagsBorderSize),
      const DeepCollectionEquality().hash(overlayBackgroundColor),
      const DeepCollectionEquality().hash(trailingIconSize),
      const DeepCollectionEquality().hash(backgroundColor),
      const DeepCollectionEquality().hash(areaBackgroundColor),
      const DeepCollectionEquality().hash(dividerColor),
      const DeepCollectionEquality().hash(disabledOpacity),
      const DeepCollectionEquality().hash(captionStrong),
      const DeepCollectionEquality().hash(caption),
      const DeepCollectionEquality().hash(captionRegularGray171),
      const DeepCollectionEquality().hash(bodyStrong),
      const DeepCollectionEquality().hash(body),
      const DeepCollectionEquality().hash(subtitleStrong),
      const DeepCollectionEquality().hash(linkButton),
      const DeepCollectionEquality().hash(title),
      const DeepCollectionEquality().hash(linkNormal),
      const DeepCollectionEquality().hash(linkSmall),
      const DeepCollectionEquality().hash(textDisabled),
      const DeepCollectionEquality().hash(area),
    ]);
  }
}

extension AppThemeBuildContextProps on BuildContext {
  AppTheme get appTheme => Theme.of(this).extension<AppTheme>()!;
  double get borderRadiusLarge => appTheme.borderRadiusLarge;
  double get borderRadiusMedium => appTheme.borderRadiusMedium;
  double get borderRadiusSmall => appTheme.borderRadiusSmall;
  double get padding => appTheme.padding;
  double get margin => appTheme.margin;
  double get outerPadding => appTheme.outerPadding;
  Color get borderColor => appTheme.borderColor;
  double get verticalSpaceSmall => appTheme.verticalSpaceSmall;
  double get verticalSpaceMedium => appTheme.verticalSpaceMedium;
  double get verticalSpaceLarge => appTheme.verticalSpaceLarge;
  double get horizontalSpaceSmall => appTheme.horizontalSpaceSmall;
  double get horizontalSpace => appTheme.horizontalSpace;
  Color get textErrorColor => appTheme.textErrorColor;
  Color get successColor => appTheme.successColor;
  double get flagsBorderSize => appTheme.flagsBorderSize;
  Color get overlayBackgroundColor => appTheme.overlayBackgroundColor;
  double get trailingIconSize => appTheme.trailingIconSize;
  Color get backgroundColor => appTheme.backgroundColor;
  Color get areaBackgroundColor => appTheme.areaBackgroundColor;
  Color get dividerColor => appTheme.dividerColor;
  double get disabledOpacity => appTheme.disabledOpacity;
  TextStyle get captionStrong => appTheme.captionStrong;
  TextStyle get caption => appTheme.caption;
  TextStyle get captionRegularGray171 => appTheme.captionRegularGray171;
  TextStyle get bodyStrong => appTheme.bodyStrong;
  TextStyle get body => appTheme.body;
  TextStyle get subtitleStrong => appTheme.subtitleStrong;
  TextStyle get linkButton => appTheme.linkButton;
  TextStyle get title => appTheme.title;
  TextStyle get linkNormal => appTheme.linkNormal;
  TextStyle get linkSmall => appTheme.linkSmall;
  TextStyle get textDisabled => appTheme.textDisabled;
  Color get area => appTheme.area;
}
