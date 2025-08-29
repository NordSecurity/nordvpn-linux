// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'servers_list_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$ServersListThemeTailorMixin on ThemeExtension<ServersListTheme> {
  double get flagSize;
  double get loaderSize;
  double get listItemHeight;
  EdgeInsetsGeometry get paddingSearchGroupsLabel;
  TextStyle get searchHintStyle;
  TextStyle get obfuscationSearchWarningStyle;
  TextStyle get searchErrorStyle;
  Color get obfuscatedItemBackgroundColor;

  @override
  ServersListTheme copyWith({
    double? flagSize,
    double? loaderSize,
    double? listItemHeight,
    EdgeInsetsGeometry? paddingSearchGroupsLabel,
    TextStyle? searchHintStyle,
    TextStyle? obfuscationSearchWarningStyle,
    TextStyle? searchErrorStyle,
    Color? obfuscatedItemBackgroundColor,
  }) {
    return ServersListTheme(
      flagSize: flagSize ?? this.flagSize,
      loaderSize: loaderSize ?? this.loaderSize,
      listItemHeight: listItemHeight ?? this.listItemHeight,
      paddingSearchGroupsLabel:
          paddingSearchGroupsLabel ?? this.paddingSearchGroupsLabel,
      searchHintStyle: searchHintStyle ?? this.searchHintStyle,
      obfuscationSearchWarningStyle:
          obfuscationSearchWarningStyle ?? this.obfuscationSearchWarningStyle,
      searchErrorStyle: searchErrorStyle ?? this.searchErrorStyle,
      obfuscatedItemBackgroundColor:
          obfuscatedItemBackgroundColor ?? this.obfuscatedItemBackgroundColor,
    );
  }

  @override
  ServersListTheme lerp(
    covariant ThemeExtension<ServersListTheme>? other,
    double t,
  ) {
    if (other is! ServersListTheme) return this as ServersListTheme;
    return ServersListTheme(
      flagSize: t < 0.5 ? flagSize : other.flagSize,
      loaderSize: t < 0.5 ? loaderSize : other.loaderSize,
      listItemHeight: t < 0.5 ? listItemHeight : other.listItemHeight,
      paddingSearchGroupsLabel: t < 0.5
          ? paddingSearchGroupsLabel
          : other.paddingSearchGroupsLabel,
      searchHintStyle: TextStyle.lerp(
        searchHintStyle,
        other.searchHintStyle,
        t,
      )!,
      obfuscationSearchWarningStyle: TextStyle.lerp(
        obfuscationSearchWarningStyle,
        other.obfuscationSearchWarningStyle,
        t,
      )!,
      searchErrorStyle: TextStyle.lerp(
        searchErrorStyle,
        other.searchErrorStyle,
        t,
      )!,
      obfuscatedItemBackgroundColor: Color.lerp(
        obfuscatedItemBackgroundColor,
        other.obfuscatedItemBackgroundColor,
        t,
      )!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is ServersListTheme &&
            const DeepCollectionEquality().equals(flagSize, other.flagSize) &&
            const DeepCollectionEquality().equals(
              loaderSize,
              other.loaderSize,
            ) &&
            const DeepCollectionEquality().equals(
              listItemHeight,
              other.listItemHeight,
            ) &&
            const DeepCollectionEquality().equals(
              paddingSearchGroupsLabel,
              other.paddingSearchGroupsLabel,
            ) &&
            const DeepCollectionEquality().equals(
              searchHintStyle,
              other.searchHintStyle,
            ) &&
            const DeepCollectionEquality().equals(
              obfuscationSearchWarningStyle,
              other.obfuscationSearchWarningStyle,
            ) &&
            const DeepCollectionEquality().equals(
              searchErrorStyle,
              other.searchErrorStyle,
            ) &&
            const DeepCollectionEquality().equals(
              obfuscatedItemBackgroundColor,
              other.obfuscatedItemBackgroundColor,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(flagSize),
      const DeepCollectionEquality().hash(loaderSize),
      const DeepCollectionEquality().hash(listItemHeight),
      const DeepCollectionEquality().hash(paddingSearchGroupsLabel),
      const DeepCollectionEquality().hash(searchHintStyle),
      const DeepCollectionEquality().hash(obfuscationSearchWarningStyle),
      const DeepCollectionEquality().hash(searchErrorStyle),
      const DeepCollectionEquality().hash(obfuscatedItemBackgroundColor),
    );
  }
}

extension ServersListThemeBuildContextProps on BuildContext {
  ServersListTheme get serversListTheme =>
      Theme.of(this).extension<ServersListTheme>()!;
  double get flagSize => serversListTheme.flagSize;
  double get loaderSize => serversListTheme.loaderSize;
  double get listItemHeight => serversListTheme.listItemHeight;
  EdgeInsetsGeometry get paddingSearchGroupsLabel =>
      serversListTheme.paddingSearchGroupsLabel;
  TextStyle get searchHintStyle => serversListTheme.searchHintStyle;
  TextStyle get obfuscationSearchWarningStyle =>
      serversListTheme.obfuscationSearchWarningStyle;
  TextStyle get searchErrorStyle => serversListTheme.searchErrorStyle;
  Color get obfuscatedItemBackgroundColor =>
      serversListTheme.obfuscatedItemBackgroundColor;
}
