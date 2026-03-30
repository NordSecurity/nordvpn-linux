// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'connection_card_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$ConnectionCardThemeTailorMixin on ThemeExtension<ConnectionCardTheme> {
  double get height;
  double get maxConnectButtonWidth;
  TextStyle get primaryFont;
  ButtonStyle get secureMyConnectionButtonStyle;
  ButtonStyle get cancelButtonStyle;
  EdgeInsetsGeometry get connectionCardPadding;
  double get smallSpacing;
  double get mediumSpacing;
  ConnectionCardLabelTheme get labelTheme;
  ConnectionCardIconTheme get iconTheme;

  @override
  ConnectionCardTheme copyWith({
    double? height,
    double? maxConnectButtonWidth,
    TextStyle? primaryFont,
    ButtonStyle? secureMyConnectionButtonStyle,
    ButtonStyle? cancelButtonStyle,
    EdgeInsetsGeometry? connectionCardPadding,
    double? smallSpacing,
    double? mediumSpacing,
    ConnectionCardLabelTheme? labelTheme,
    ConnectionCardIconTheme? iconTheme,
  }) {
    return ConnectionCardTheme(
      height: height ?? this.height,
      maxConnectButtonWidth:
          maxConnectButtonWidth ?? this.maxConnectButtonWidth,
      primaryFont: primaryFont ?? this.primaryFont,
      secureMyConnectionButtonStyle:
          secureMyConnectionButtonStyle ?? this.secureMyConnectionButtonStyle,
      cancelButtonStyle: cancelButtonStyle ?? this.cancelButtonStyle,
      connectionCardPadding:
          connectionCardPadding ?? this.connectionCardPadding,
      smallSpacing: smallSpacing ?? this.smallSpacing,
      mediumSpacing: mediumSpacing ?? this.mediumSpacing,
      labelTheme: labelTheme ?? this.labelTheme,
      iconTheme: iconTheme ?? this.iconTheme,
    );
  }

  @override
  ConnectionCardTheme lerp(
    covariant ThemeExtension<ConnectionCardTheme>? other,
    double t,
  ) {
    if (other is! ConnectionCardTheme) return this as ConnectionCardTheme;
    return ConnectionCardTheme(
      height: t < 0.5 ? height : other.height,
      maxConnectButtonWidth: t < 0.5
          ? maxConnectButtonWidth
          : other.maxConnectButtonWidth,
      primaryFont: TextStyle.lerp(primaryFont, other.primaryFont, t)!,
      secureMyConnectionButtonStyle: t < 0.5
          ? secureMyConnectionButtonStyle
          : other.secureMyConnectionButtonStyle,
      cancelButtonStyle: t < 0.5 ? cancelButtonStyle : other.cancelButtonStyle,
      connectionCardPadding: t < 0.5
          ? connectionCardPadding
          : other.connectionCardPadding,
      smallSpacing: t < 0.5 ? smallSpacing : other.smallSpacing,
      mediumSpacing: t < 0.5 ? mediumSpacing : other.mediumSpacing,
      labelTheme: t < 0.5 ? labelTheme : other.labelTheme,
      iconTheme: t < 0.5 ? iconTheme : other.iconTheme,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is ConnectionCardTheme &&
            const DeepCollectionEquality().equals(height, other.height) &&
            const DeepCollectionEquality().equals(
              maxConnectButtonWidth,
              other.maxConnectButtonWidth,
            ) &&
            const DeepCollectionEquality().equals(
              primaryFont,
              other.primaryFont,
            ) &&
            const DeepCollectionEquality().equals(
              secureMyConnectionButtonStyle,
              other.secureMyConnectionButtonStyle,
            ) &&
            const DeepCollectionEquality().equals(
              cancelButtonStyle,
              other.cancelButtonStyle,
            ) &&
            const DeepCollectionEquality().equals(
              connectionCardPadding,
              other.connectionCardPadding,
            ) &&
            const DeepCollectionEquality().equals(
              smallSpacing,
              other.smallSpacing,
            ) &&
            const DeepCollectionEquality().equals(
              mediumSpacing,
              other.mediumSpacing,
            ) &&
            const DeepCollectionEquality().equals(
              labelTheme,
              other.labelTheme,
            ) &&
            const DeepCollectionEquality().equals(iconTheme, other.iconTheme));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(height),
      const DeepCollectionEquality().hash(maxConnectButtonWidth),
      const DeepCollectionEquality().hash(primaryFont),
      const DeepCollectionEquality().hash(secureMyConnectionButtonStyle),
      const DeepCollectionEquality().hash(cancelButtonStyle),
      const DeepCollectionEquality().hash(connectionCardPadding),
      const DeepCollectionEquality().hash(smallSpacing),
      const DeepCollectionEquality().hash(mediumSpacing),
      const DeepCollectionEquality().hash(labelTheme),
      const DeepCollectionEquality().hash(iconTheme),
    );
  }
}

extension ConnectionCardThemeBuildContextProps on BuildContext {
  ConnectionCardTheme get connectionCardTheme =>
      Theme.of(this).extension<ConnectionCardTheme>()!;
  double get height => connectionCardTheme.height;
  double get maxConnectButtonWidth => connectionCardTheme.maxConnectButtonWidth;
  TextStyle get primaryFont => connectionCardTheme.primaryFont;
  ButtonStyle get secureMyConnectionButtonStyle =>
      connectionCardTheme.secureMyConnectionButtonStyle;
  ButtonStyle get cancelButtonStyle => connectionCardTheme.cancelButtonStyle;
  EdgeInsetsGeometry get connectionCardPadding =>
      connectionCardTheme.connectionCardPadding;
  double get smallSpacing => connectionCardTheme.smallSpacing;
  double get mediumSpacing => connectionCardTheme.mediumSpacing;
  ConnectionCardLabelTheme get labelTheme => connectionCardTheme.labelTheme;
  ConnectionCardIconTheme get iconTheme => connectionCardTheme.iconTheme;
}
