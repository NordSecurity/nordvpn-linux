// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'allow_list_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$AllowListThemeTailorMixin on ThemeExtension<AllowListTheme> {
  TextStyle get labelStyle;
  Color get addCardBackground;
  Color get dividerColor;
  Color get listItemBackgroundColor;
  TextStyle get tableHeaderStyle;
  TextStyle get tableItemsStyle;

  @override
  AllowListTheme copyWith({
    TextStyle? labelStyle,
    Color? addCardBackground,
    Color? dividerColor,
    Color? listItemBackgroundColor,
    TextStyle? tableHeaderStyle,
    TextStyle? tableItemsStyle,
  }) {
    return AllowListTheme(
      labelStyle: labelStyle ?? this.labelStyle,
      addCardBackground: addCardBackground ?? this.addCardBackground,
      dividerColor: dividerColor ?? this.dividerColor,
      listItemBackgroundColor:
          listItemBackgroundColor ?? this.listItemBackgroundColor,
      tableHeaderStyle: tableHeaderStyle ?? this.tableHeaderStyle,
      tableItemsStyle: tableItemsStyle ?? this.tableItemsStyle,
    );
  }

  @override
  AllowListTheme lerp(
    covariant ThemeExtension<AllowListTheme>? other,
    double t,
  ) {
    if (other is! AllowListTheme) return this as AllowListTheme;
    return AllowListTheme(
      labelStyle: TextStyle.lerp(labelStyle, other.labelStyle, t)!,
      addCardBackground: Color.lerp(
        addCardBackground,
        other.addCardBackground,
        t,
      )!,
      dividerColor: Color.lerp(dividerColor, other.dividerColor, t)!,
      listItemBackgroundColor: Color.lerp(
        listItemBackgroundColor,
        other.listItemBackgroundColor,
        t,
      )!,
      tableHeaderStyle: TextStyle.lerp(
        tableHeaderStyle,
        other.tableHeaderStyle,
        t,
      )!,
      tableItemsStyle: TextStyle.lerp(
        tableItemsStyle,
        other.tableItemsStyle,
        t,
      )!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is AllowListTheme &&
            const DeepCollectionEquality().equals(
              labelStyle,
              other.labelStyle,
            ) &&
            const DeepCollectionEquality().equals(
              addCardBackground,
              other.addCardBackground,
            ) &&
            const DeepCollectionEquality().equals(
              dividerColor,
              other.dividerColor,
            ) &&
            const DeepCollectionEquality().equals(
              listItemBackgroundColor,
              other.listItemBackgroundColor,
            ) &&
            const DeepCollectionEquality().equals(
              tableHeaderStyle,
              other.tableHeaderStyle,
            ) &&
            const DeepCollectionEquality().equals(
              tableItemsStyle,
              other.tableItemsStyle,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(labelStyle),
      const DeepCollectionEquality().hash(addCardBackground),
      const DeepCollectionEquality().hash(dividerColor),
      const DeepCollectionEquality().hash(listItemBackgroundColor),
      const DeepCollectionEquality().hash(tableHeaderStyle),
      const DeepCollectionEquality().hash(tableItemsStyle),
    );
  }
}

extension AllowListThemeBuildContextProps on BuildContext {
  AllowListTheme get allowListTheme =>
      Theme.of(this).extension<AllowListTheme>()!;
  TextStyle get labelStyle => allowListTheme.labelStyle;
  Color get addCardBackground => allowListTheme.addCardBackground;
  Color get dividerColor => allowListTheme.dividerColor;
  Color get listItemBackgroundColor => allowListTheme.listItemBackgroundColor;
  TextStyle get tableHeaderStyle => allowListTheme.tableHeaderStyle;
  TextStyle get tableItemsStyle => allowListTheme.tableItemsStyle;
}
