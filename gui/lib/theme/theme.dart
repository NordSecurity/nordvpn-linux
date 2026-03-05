import 'dart:ui';

import 'package:flutter/material.dart';
import 'package:nordvpn/theme/allow_list_theme.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/aurora_design.dart';
import 'package:nordvpn/theme/autoconnect_panel_theme.dart';
import 'package:nordvpn/theme/consent_screen_theme.dart';
import 'package:nordvpn/theme/copy_field_theme.dart';
import 'package:nordvpn/theme/custom_dns_theme.dart';
import 'package:nordvpn/theme/dropdown_theme.dart';
import 'package:nordvpn/theme/error_screen_theme.dart';
import 'package:nordvpn/theme/inline_loading_indicator_theme.dart';
import 'package:nordvpn/theme/input_theme.dart';
import 'package:nordvpn/theme/interactive_list_view_theme.dart';
import 'package:nordvpn/theme/loading_indicator_theme.dart';
import 'package:nordvpn/theme/login_form_theme.dart';
import 'package:nordvpn/theme/on_off_switch_theme.dart';
import 'package:nordvpn/theme/radio_button_theme.dart';
import 'package:nordvpn/theme/servers_list_theme.dart';
import 'package:nordvpn/theme/settings_theme.dart';
import 'package:nordvpn/theme/support_link_theme.dart';
import 'package:nordvpn/theme/vpn_status_card_theme.dart';
import 'package:nordvpn/theme/popup_theme.dart';

ThemeData lightTheme() {
  return NordVpnTheme(ThemeMode.light).data();
}

ThemeData darkTheme() {
  return NordVpnTheme(ThemeMode.dark).data();
}

final class NordVpnTheme {
  final ThemeMode mode;

  final SemanticColors semanticColors;

  NordVpnTheme(this.mode) : semanticColors = mode == ThemeMode.dark ? SemanticColors(true) : SemanticColors(false);

  ThemeData data() {
    final data = mode == ThemeMode.light
        ? ThemeData.light(useMaterial3: true)
        : ThemeData.dark(useMaterial3: true);
    return data.copyWith(
      scaffoldBackgroundColor: semanticColors.bgPrimary,
      disabledColor: semanticColors.textDisabled,
      navigationRailTheme: _navigationRailTheme(),
      navigationBarTheme: _navigationBarTheme(),
      tabBarTheme: _tabBarTheme(),
      appBarTheme: _appBarTheme(),
      colorScheme: _colorScheme(),
      elevatedButtonTheme: _elevatedButtonThemeData(),
      inputDecorationTheme: _inputDecorationTheme(),
      textButtonTheme: _textButtonThemeData(),
      outlinedButtonTheme: _outlinedButtonThemeData(),
      iconButtonTheme: _iconButtonThemeData(),
      dividerTheme: _dividerTheme(),
      checkboxTheme: _checkboxThemeData(),
      tooltipTheme: _tooltipThemeData(),
      extensions: [
        _appThemeExt(),
        _vpnStatusCardThemeExt(),
        _serversListThemeExt(),
        _settingsThemeExt(),
        _onOffSwitchThemeExt(),
        _radioButtonThemeExt(),
        _inputThemeExt(),
        _loadingIndicatorThemeExt(),
        _copyFieldThemeExt(),
        _supportLinkThemeExt(),
        _loginFormThemeExt(),
        _inlineLoadingIndicatorThemeExt(),
        _autoconnectPanelTheme(),
        _customDnsThemeExt(),
        _dropdownThemeExt(),
        _allowListThemeExt(),
        _interactiveListViewThemeExt(),
        _errorScreenThemeExt(),
        _consentScreenThemeExt(),
        _popupThemeExt(),
      ],
    );
  }

  NavigationRailThemeData _navigationRailTheme() {
    return NavigationRailThemeData(
      backgroundColor: semanticColors.bgPrimary,
      labelType: NavigationRailLabelType.all,
      useIndicator: true,
      selectedLabelTextStyle: AppTypography.body.copyWith(
        color: semanticColors.textPrimary,
      ),
      unselectedLabelTextStyle: AppTypography.body.copyWith(
        color: semanticColors.textPrimary,
      ),
      indicatorColor: semanticColors.bgSecondary,
    );
  }

  NavigationBarThemeData _navigationBarTheme() {
    return NavigationBarThemeData(
      indicatorColor: semanticColors.bgSecondary,
      backgroundColor: semanticColors.bgPrimary,
      surfaceTintColor: AppCoreColors.transparent,
      labelTextStyle: WidgetStateProperty.all(
        AppTypography.body.copyWith(
          color: semanticColors.textPrimary,
        ),
      ),
    );
  }

  TabBarThemeData _tabBarTheme() {
    return TabBarThemeData(
      indicator: UnderlineTabIndicator(
        borderSide: BorderSide(width: 2, color: semanticColors.bgAccent),
      ),
      tabAlignment: TabAlignment.start,
      indicatorSize: TabBarIndicatorSize.tab,
      indicatorColor: semanticColors.bgAccent,
      dividerColor: AppCoreColors.transparent,
      dividerHeight: 0,
      labelStyle: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
      unselectedLabelStyle: AppTypography.subBody.copyWith(
        color: semanticColors.textSecondary,
      ),
    );
  }

  AppBarTheme _appBarTheme() {
    return AppBarTheme(
      toolbarHeight: 70,
      shape: Border(
        bottom: BorderSide(
          color: semanticColors.borderSecondary,
          width: 1,
        ),
      ),
      surfaceTintColor: AppCoreColors.transparent,
      centerTitle: true,
      backgroundColor: semanticColors.bgSecondary,
    );
  }

  ColorScheme _colorScheme() {
    return ColorScheme.fromSeed(
      surface: semanticColors.bgSecondary,
      seedColor: semanticColors.bgAccent,
    );
  }

  ElevatedButtonThemeData _elevatedButtonThemeData() {
    return ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
        textStyle: AppTypography.subHeading.copyWith(
          color: semanticColors.textPrimary,
        ),
        padding: const EdgeInsets.symmetric(horizontal: 25.0, vertical: 16),
        backgroundColor: semanticColors.bgAccent,
        foregroundColor: semanticColors.textPrimaryOnColor,
        disabledBackgroundColor: semanticColors.bgDisabled,
        disabledForegroundColor: semanticColors.textDisabled,
      ),
    );
  }

  InputDecorationTheme _inputDecorationTheme() {
    final border = OutlineInputBorder(
      borderRadius: BorderRadius.circular(8),
      borderSide: BorderSide(
        color: semanticColors.borderSecondary,
        width: 1.0,
      ),
    );

    return InputDecorationTheme(
      constraints: const BoxConstraints(maxHeight: 50),
      isDense: true,
      contentPadding: const EdgeInsets.all(10),
      border: border,
      focusedBorder: border,
      enabledBorder: border,
      hintStyle: AppTypography.body.copyWith(
        color: semanticColors.textSecondary,
      ),
      floatingLabelBehavior: FloatingLabelBehavior.never,
    );
  }

  TextButtonThemeData _textButtonThemeData() {
    return TextButtonThemeData(
      style: TextButton.styleFrom(
        side: BorderSide.none,
        overlayColor: AppCoreColors.transparent,
        textStyle: AppTypography.body.copyWith(
          color: semanticColors.textPrimary,
        ),
        foregroundColor: semanticColors.textAccent,
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 16),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(4),
          side: BorderSide(
            color: semanticColors.borderPrimary,
            width: 1,
          ),
        ),
      ),
    );
  }

  OutlinedButtonThemeData _outlinedButtonThemeData() {
    return OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        side: BorderSide(color: semanticColors.borderPrimary),
        textStyle: AppTypography.subHeading.copyWith(
          color: semanticColors.textPrimary,
        ),
        foregroundColor: semanticColors.textPrimary,
        padding: const EdgeInsets.symmetric(horizontal: 25.0, vertical: 16),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
        disabledBackgroundColor: semanticColors.bgDisabled,
        disabledForegroundColor: semanticColors.textDisabled,
      ),
    );
  }

  IconButtonThemeData _iconButtonThemeData() {
    return IconButtonThemeData(
      style: IconButton.styleFrom(padding: const EdgeInsets.all(10)),
    );
  }

  DividerThemeData _dividerTheme() {
    return DividerThemeData(
      color: semanticColors.borderSecondary,
      thickness: 1,
    );
  }

  CheckboxThemeData _checkboxThemeData() {
    return CheckboxThemeData(
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
      overlayColor: WidgetStateProperty.all(AppCoreColors.transparent),
      checkColor: WidgetStateProperty.resolveWith<Color?>((states) {
        if (states.contains(WidgetState.selected) &&
            states.contains(WidgetState.disabled)) {
          return semanticColors.bgDisabled;
        }
        if (states.contains(WidgetState.selected)) {
          return semanticColors.bgSecondary;
        }
        if (states.contains(WidgetState.disabled)) {
          return semanticColors.bgDisabled;
        }
        return semanticColors.bgSecondary;
      }),
      fillColor: WidgetStateProperty.resolveWith<Color?>((states) {
        if (states.contains(WidgetState.selected) &&
            states.contains(WidgetState.disabled)) {
          return semanticColors.bgDisabled;
        }
        if (states.contains(WidgetState.selected)) {
          return semanticColors.bgAccent;
        }
        if (states.contains(WidgetState.disabled)) {
          return semanticColors.bgSecondary;
        }
        return semanticColors.bgSecondary;
      }),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
      side: WidgetStateBorderSide.resolveWith((states) {
        if (states.contains(WidgetState.selected) &&
            states.contains(WidgetState.disabled)) {
          return BorderSide(width: 1, color: semanticColors.bgDisabled);
        }
        if (states.contains(WidgetState.selected)) {
          return BorderSide(width: 1, color: semanticColors.bgAccent);
        }
        if (states.contains(WidgetState.disabled)) {
          return BorderSide(width: 1, color: semanticColors.bgDisabled);
        }
        return BorderSide(width: 1, color: semanticColors.borderInput);
      }),
    );
  }

  AppTheme _appThemeExt() {
    return AppTheme(
      borderRadiusLarge: 16,
      borderRadiusMedium: 8,
      borderRadiusSmall: 4,
      padding: 10,
      margin: 8,
      outerPadding: 16,
      borderColor: semanticColors.borderSecondary,
      verticalSpaceVerySmall: 2,
      verticalSpaceSmall: 8,
      verticalSpaceMedium: 16,
      verticalSpaceLarge: 24,
      verticalSpaceExtraLarge: 48,
      horizontalSpaceSmall: 8,
      horizontalSpace: 16,
      textErrorColor: semanticColors.textCritical,
      successColor: semanticColors.textSuccess,
      flagsBorderSize: 2,
      overlayBackgroundColor: AppCoreColors.neutral1000.withAlpha(127),
      caption: AppTypography.subBody.copyWith(
        color: semanticColors.textSecondary,
      ),
      captionRegularGray171: TextStyle(
        fontSize: 12,
        color: semanticColors.textSecondary,
        fontWeight: FontWeight.w400,
      ),
      captionStrong: AppTypography.captionMedium.copyWith(
        color: semanticColors.textPrimary,
      ),
      bodyStrong: AppTypography.subHeading.copyWith(
        color: semanticColors.textPrimary,
      ),
      body: AppTypography.body.copyWith(
        color: semanticColors.textPrimary,
      ),
      subtitleStrong: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
      linkButton: TextStyle(
        fontSize: 12,
        color: semanticColors.bgAccent,
        fontWeight: FontWeight.w400,
      ),
      title: AppTypography.display.copyWith(
        color: semanticColors.textPrimary,
      ),
      trailingIconSize: 32,
      backgroundColor: semanticColors.bgSecondary,
      areaBackgroundColor: semanticColors.bgSecondaryActive,
      area: semanticColors.bgSecondaryActive,
      dividerColor: semanticColors.borderSecondary,
      disabledOpacity: 0.5,
      linkNormal: AppTypography.subBody.copyWith(
        color: semanticColors.textAccent,
      ),
      linkSmall: AppTypography.caption.copyWith(
        color: semanticColors.textAccent,
      ),
      textDisabled: AppTypography.body.copyWith(
        color: semanticColors.textDisabled,
      ),
    );
  }

  VpnStatusCardTheme _vpnStatusCardThemeExt() {
    return VpnStatusCardTheme(
      height: 150,
      maxConnectButtonWidth: 408,
      primaryFont: AppTypography.captionMedium.copyWith(
        color: semanticColors.textPrimary,
      ),
      secondaryFont: AppTypography.subHeading.copyWith(
        color: semanticColors.textPrimary,
      ),
      iconSize: 40,
    );
  }

  ServersListTheme _serversListThemeExt() {
    return ServersListTheme(
      flagSize: 32,
      loaderSize: 28,
      listItemHeight: 44,
      paddingSearchGroupsLabel: const EdgeInsets.symmetric(
        horizontal: 32,
        vertical: 8,
      ),
      searchHintStyle: AppTypography.body.copyWith(
        color: semanticColors.textSecondary,
      ),
      searchErrorStyle: AppTypography.body.copyWith(
        color: semanticColors.textCritical,
      ),
      obfuscationSearchWarningStyle: AppTypography.body.copyWith(
        color: semanticColors.textSecondary,
      ),
      obfuscatedItemBackgroundColor: semanticColors.bgSecondaryActive,
      horizontalSpace: 16,
    );
  }

  SettingsTheme _settingsThemeExt() {
    return SettingsTheme(
      currentPageNameStyle: AppTypography.heading.copyWith(
        color: semanticColors.textPrimary,
      ),
      parentPageStyle: AppTypography.heading.copyWith(
        color: semanticColors.textSecondary,
      ),
      itemTitleStyle: AppTypography.body.copyWith(
        color: semanticColors.textPrimary,
      ),
      itemSubtitleStyle: AppTypography.subBody.copyWith(
        color: semanticColors.textSecondary,
      ),
      vpnStatusStyle: AppTypography.captionMedium.copyWith(
        color: semanticColors.textPrimary,
      ),
      textInputWidth: 220,
      otherProductsTitle: AppTypography.body.copyWith(
        color: semanticColors.textPrimary,
      ),
      otherProductsSubtitle: AppTypography.subBody.copyWith(
        color: semanticColors.textSecondary,
      ),
      fwMarkInputSize: 250,
      itemPadding: EdgeInsets.symmetric(horizontal: 4, vertical: 16),
    );
  }

  OnOffSwitchTheme _onOffSwitchThemeExt() {
    return OnOffSwitchTheme(
      label: OnOffLabelTheme(
        width: 32,
        paddingRight: 10,
        textStyle: AppTypography.body.copyWith(
          color: semanticColors.textPrimary,
        ),
        disabledTextStyle: AppTypography.body.copyWith(
          color: semanticColors.textDisabled,
        ),
      ),
      slider: OnOffSliderTheme(
        width: 38,
        height: 18,
        on: SwitchOnOffProps(
          leftOffset: 21,
          rightOffset: 0,
          color: AppCoreColors.neutral0,
          borderColor: semanticColors.bgAccent,
          backgroundColor: semanticColors.bgAccent,
        ),
        off: SwitchOnOffProps(
          leftOffset: 0.8,
          rightOffset: 21,
          color: semanticColors.textPrimary,
          borderColor: semanticColors.borderInput,
          backgroundColor: semanticColors.bgSecondary,
        ),
        disabledOn: SwitchOnOffProps(
          leftOffset: 0.8,
          rightOffset: 21,
          color: semanticColors.bgTertiary,
          borderColor: semanticColors.bgDisabled,
          backgroundColor: semanticColors.bgDisabled,
        ),
        disabledOff: SwitchOnOffProps(
          leftOffset: 0.8,
          rightOffset: 21,
          color: semanticColors.bgDisabled,
          borderColor: semanticColors.borderPrimary,
          backgroundColor: AppCoreColors.transparent,
        ),
        bottomOffset: 2.1,
        topOffset: 2.1,
        borderRadius: 999.0,
      ),
    );
  }

  RadioButtonTheme _radioButtonThemeExt() {
    return RadioButtonTheme(
      padding: 10,
      label: RadioLabelTheme(width: 10, paddingLeft: 10),
      radio: RadioStyle(
        borderWidth: 6,
        width: 18,
        height: 18,
        on: RadioOnOffProps(
          fillColor: semanticColors.bgAccent,
          borderColor: semanticColors.bgAccent,
          dotColor: AppCoreColors.neutral0,
          dotHeight: 6,
          dotWidth: 6,
          borderWidth: 6,
        ),
        off: RadioOnOffProps(
          fillColor: semanticColors.bgSecondary,
          borderColor: semanticColors.borderInput,
          dotColor: AppCoreColors.transparent,
          dotHeight: 0,
          dotWidth: 0,
          borderWidth: 1,
        ),
      ),
    );
  }

  InputTheme _inputThemeExt() {
    return InputTheme(
      height: 56,
      errorStyle: AppTypography.subBody.copyWith(
        color: semanticColors.textCritical,
      ),
      textStyle: AppTypography.body.copyWith(
        color: semanticColors.textPrimary,
      ),
      enabled: EnabledStyle(
        borderColor: semanticColors.borderPrimary,
        borderWidth: 1,
      ),
      focused: FocusedStyle(
        borderColor: semanticColors.bgAccent,
        borderWidth: 2,
      ),
      error: ErrorStyle(
        borderColor: semanticColors.borderCritical,
        borderWidth: 1,
      ),
      focusedError: FocusedErrorStyle(
        borderColor: semanticColors.borderCritical,
        borderWidth: 1,
      ),
      icon: IconStyle(
        color: semanticColors.textPrimary,
        hoverColor: AppCoreColors.transparent,
      ),
    );
  }

  LoadingIndicatorTheme _loadingIndicatorThemeExt() {
    return LoadingIndicatorTheme(
      color: semanticColors.bgAccent,
      strokeWidth: 2,
    );
  }

  CopyFieldTheme _copyFieldThemeExt() {
    return CopyFieldTheme(
      borderRadius: 4,
      commandTextStyle: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
      descriptionTextStyle: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
    );
  }

  SupportLinkTheme _supportLinkThemeExt() {
    return SupportLinkTheme(
      textStyle: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
      urlColor: semanticColors.bgAccent,
    );
  }

  LoginFormTheme _loginFormThemeExt() {
    return LoginFormTheme(
      titleStyle: AppTypography.display.copyWith(
        color: semanticColors.textPrimary,
      ),
      checkboxDescStyle: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
      width: 424,
      height: 348,
      progressIndicator: LoginButtonProgressIndicatorTheme(
        height: 16,
        width: 16,
        stroke: 1.5,
        color: semanticColors.bgSecondary,
      ),
    );
  }

  InlineLoadingIndicatorTheme _inlineLoadingIndicatorThemeExt() {
    return InlineLoadingIndicatorTheme(
      width: 15,
      height: 15,
      stroke: 2,
      color: semanticColors.bgAccent,
      alternativeColor: AppCoreColors.neutral0,
    );
  }

  AutoconnectPanelTheme _autoconnectPanelTheme() {
    return AutoconnectPanelTheme(
      primaryFont: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
      secondaryFont: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
      iconSize: 37,
      loaderSize: 35,
    );
  }

  CustomDnsTheme _customDnsThemeExt() {
    return CustomDnsTheme(
      formBackground: semanticColors.bgSecondaryActive,
      dnsInputWidth: 300,
      dividerColor: semanticColors.borderPrimary,
    );
  }

  AllowListTheme _allowListThemeExt() {
    return AllowListTheme(
      labelStyle: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
      addCardBackground: semanticColors.bgSecondaryActive,
      tableItemsStyle: AppTypography.body.copyWith(
        color: semanticColors.textPrimary,
      ),
      tableHeaderStyle: AppTypography.captionMedium.copyWith(
        color: semanticColors.textPrimary,
      ),
      dividerColor: semanticColors.borderPrimary,
      listItemBackgroundColor: semanticColors.bgSecondaryActive,
    );
  }

  DropdownTheme _dropdownThemeExt() {
    return DropdownTheme(
      color: semanticColors.bgSecondary,
      borderRadius: 4,
      borderColor: semanticColors.borderPrimary,
      focusBorderColor: semanticColors.borderAccent,
      errorBorderColor: semanticColors.borderCritical,
      borderWidth: 1,
      horizontalPadding: 8,
    );
  }

  TooltipThemeData _tooltipThemeData() {
    return TooltipThemeData(
      decoration: BoxDecoration(
        color: semanticColors.bgSecondary, //
        border: Border.all(
          color: semanticColors.borderPrimary,
          width: 1,
        ),
        borderRadius: BorderRadius.all(Radius.circular(5)),
      ),
      textStyle: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
    );
  }

  InteractiveListViewTheme _interactiveListViewThemeExt() {
    return InteractiveListViewTheme(
      borderRadius: 4,
      borderColor: semanticColors.borderPrimary,
      focusBorderColor: semanticColors.borderAccent,
      borderWidth: 1,
      verticalSpaceSmall: 8,
    );
  }

  ErrorScreenTheme _errorScreenThemeExt() {
    return ErrorScreenTheme(
      titleTextStyle: AppTypography.subHeading.copyWith(
        color: semanticColors.textPrimary,
      ),
      descriptionTextStyle: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
    );
  }

  ConsentScreenTheme _consentScreenThemeExt() {
    return ConsentScreenTheme(
      width: 600,
      height: 430,
      overlayColor: AppCoreColors.neutral1000.withAlpha(125),
      titleTextStyle: AppTypography.display.copyWith(
        color: semanticColors.textPrimary,
      ),
      bodyTextStyle: AppTypography.body.copyWith(
        color: semanticColors.textPrimary,
      ),
      titleBarTextStyle: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
      padding: 40,
      listItemTitle: AppTypography.body.copyWith(
        color: semanticColors.textPrimary,
      ),
      listItemSubtitle: AppTypography.subBody.copyWith(
        color: semanticColors.textPrimary,
      ),
      titleBarWidth: 45,
    );
  }

  PopupTheme _popupThemeExt() {
    return PopupTheme(
      widgetWidth: 500,
      widgetRadius: BorderRadius.all(Radius.circular(16.0)),
      contentAllPadding: AppSpacing.spacing4,
      xButtonAllPadding: AppSpacing.spacing1,
      gapBetweenElements: AppSpacing.spacing2,
      verticalElementSpacing: AppSpacing.spacing4,
      singleButtonMinWidth: AppSpacing.spacing30,
      textPrimary: AppTypography.subHeading.copyWith(
        color: semanticColors.textPrimary,
      ),
      textSecondary: AppTypography.body.copyWith(
        color: semanticColors.textSecondary,
      ),
    );
  }
}
