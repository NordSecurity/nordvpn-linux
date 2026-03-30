// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'context_menu_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$ContextMenuThemeTailorMixin on ThemeExtension<ContextMenuTheme> {
  double get menuWidth;
  BorderRadius get menuRadius;
  EdgeInsets get menuPadding;
  Color get menuColor;
  Color get menuBorderColor;
  double get menuBorderWidth;
  double get itemHeight;
  EdgeInsets get itemPadding;
  Color get itemHoverColor;
  TextStyle get itemTextStyle;

  @override
  ContextMenuTheme copyWith({
    double? menuWidth,
    BorderRadius? menuRadius,
    EdgeInsets? menuPadding,
    Color? menuColor,
    Color? menuBorderColor,
    double? menuBorderWidth,
    double? itemHeight,
    EdgeInsets? itemPadding,
    Color? itemHoverColor,
    TextStyle? itemTextStyle,
  }) {
    return ContextMenuTheme(
      menuWidth: menuWidth ?? this.menuWidth,
      menuRadius: menuRadius ?? this.menuRadius,
      menuPadding: menuPadding ?? this.menuPadding,
      menuColor: menuColor ?? this.menuColor,
      menuBorderColor: menuBorderColor ?? this.menuBorderColor,
      menuBorderWidth: menuBorderWidth ?? this.menuBorderWidth,
      itemHeight: itemHeight ?? this.itemHeight,
      itemPadding: itemPadding ?? this.itemPadding,
      itemHoverColor: itemHoverColor ?? this.itemHoverColor,
      itemTextStyle: itemTextStyle ?? this.itemTextStyle,
    );
  }

  @override
  ContextMenuTheme lerp(
    covariant ThemeExtension<ContextMenuTheme>? other,
    double t,
  ) {
    if (other is! ContextMenuTheme) return this as ContextMenuTheme;
    return ContextMenuTheme(
      menuWidth: t < 0.5 ? menuWidth : other.menuWidth,
      menuRadius: t < 0.5 ? menuRadius : other.menuRadius,
      menuPadding: t < 0.5 ? menuPadding : other.menuPadding,
      menuColor: Color.lerp(menuColor, other.menuColor, t)!,
      menuBorderColor: Color.lerp(menuBorderColor, other.menuBorderColor, t)!,
      menuBorderWidth: t < 0.5 ? menuBorderWidth : other.menuBorderWidth,
      itemHeight: t < 0.5 ? itemHeight : other.itemHeight,
      itemPadding: t < 0.5 ? itemPadding : other.itemPadding,
      itemHoverColor: Color.lerp(itemHoverColor, other.itemHoverColor, t)!,
      itemTextStyle: TextStyle.lerp(itemTextStyle, other.itemTextStyle, t)!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is ContextMenuTheme &&
            const DeepCollectionEquality().equals(menuWidth, other.menuWidth) &&
            const DeepCollectionEquality().equals(
              menuRadius,
              other.menuRadius,
            ) &&
            const DeepCollectionEquality().equals(
              menuPadding,
              other.menuPadding,
            ) &&
            const DeepCollectionEquality().equals(menuColor, other.menuColor) &&
            const DeepCollectionEquality().equals(
              menuBorderColor,
              other.menuBorderColor,
            ) &&
            const DeepCollectionEquality().equals(
              menuBorderWidth,
              other.menuBorderWidth,
            ) &&
            const DeepCollectionEquality().equals(
              itemHeight,
              other.itemHeight,
            ) &&
            const DeepCollectionEquality().equals(
              itemPadding,
              other.itemPadding,
            ) &&
            const DeepCollectionEquality().equals(
              itemHoverColor,
              other.itemHoverColor,
            ) &&
            const DeepCollectionEquality().equals(
              itemTextStyle,
              other.itemTextStyle,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(menuWidth),
      const DeepCollectionEquality().hash(menuRadius),
      const DeepCollectionEquality().hash(menuPadding),
      const DeepCollectionEquality().hash(menuColor),
      const DeepCollectionEquality().hash(menuBorderColor),
      const DeepCollectionEquality().hash(menuBorderWidth),
      const DeepCollectionEquality().hash(itemHeight),
      const DeepCollectionEquality().hash(itemPadding),
      const DeepCollectionEquality().hash(itemHoverColor),
      const DeepCollectionEquality().hash(itemTextStyle),
    );
  }
}

extension ContextMenuThemeBuildContextProps on BuildContext {
  ContextMenuTheme get contextMenuTheme =>
      Theme.of(this).extension<ContextMenuTheme>()!;
  double get menuWidth => contextMenuTheme.menuWidth;
  BorderRadius get menuRadius => contextMenuTheme.menuRadius;
  EdgeInsets get menuPadding => contextMenuTheme.menuPadding;
  Color get menuColor => contextMenuTheme.menuColor;
  Color get menuBorderColor => contextMenuTheme.menuBorderColor;
  double get menuBorderWidth => contextMenuTheme.menuBorderWidth;
  double get itemHeight => contextMenuTheme.itemHeight;
  EdgeInsets get itemPadding => contextMenuTheme.itemPadding;
  Color get itemHoverColor => contextMenuTheme.itemHoverColor;
  TextStyle get itemTextStyle => contextMenuTheme.itemTextStyle;
}
