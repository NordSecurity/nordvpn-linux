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

ThemeData lightTheme() {
  return NordVpnTheme(ThemeMode.light).data();
}

ThemeData darkTheme() {
  return NordVpnTheme(ThemeMode.dark).data();
}

final class NordVpnTheme {
  final ThemeMode mode;

  final AppDesign design;

  NordVpnTheme(this.mode) : design = AppDesign(mode);

  ThemeData data() {
    final data = mode == ThemeMode.light
        ? ThemeData.light(useMaterial3: true)
        : ThemeData.dark(useMaterial3: true);
    return data.copyWith(
      scaffoldBackgroundColor: design.semanticColors.bgPrimary,
      disabledColor: design.semanticColors.textDisabled,
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
      ],
    );
  }

  NavigationRailThemeData _navigationRailTheme() {
    return NavigationRailThemeData(
      backgroundColor: design.semanticColors.bgPrimary,
      labelType: NavigationRailLabelType.all,
      useIndicator: true,
      selectedLabelTextStyle: design.typography.body.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      unselectedLabelTextStyle: design.typography.body.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      indicatorColor: design.semanticColors.bgSecondary,
    );
  }

  NavigationBarThemeData _navigationBarTheme() {
    return NavigationBarThemeData(
      indicatorColor: design.semanticColors.bgSecondary,
      backgroundColor: design.semanticColors.bgPrimary,
      surfaceTintColor: design.colors.transparent,
      labelTextStyle: WidgetStateProperty.all(design.typography.body.copyWith(
        color: design.semanticColors.textPrimary,
      )),
    );
  }

  TabBarThemeData _tabBarTheme() {
    return TabBarThemeData(
      indicator: UnderlineTabIndicator(
        borderSide: BorderSide(width: 2, color: design.semanticColors.bgAccent),
      ),
      tabAlignment: TabAlignment.start,
      indicatorSize: TabBarIndicatorSize.tab,
      indicatorColor: design.semanticColors.bgAccent,
      dividerColor: design.colors.transparent,
      dividerHeight: 0,
      labelStyle: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      unselectedLabelStyle: design.typography.subBody.copyWith(
        color: design.semanticColors.textSecondary,
      ),
    );
  }

  AppBarTheme _appBarTheme() {
    return AppBarTheme(
      toolbarHeight: 70,
      shape: Border(
        bottom: BorderSide(
          color: design.semanticColors.borderSecondary,
          width: 1,
        ),
      ),
      surfaceTintColor: design.colors.transparent,
      centerTitle: true,
      backgroundColor: design.semanticColors.bgSecondary,
    );
  }

  ColorScheme _colorScheme() {
    return ColorScheme.fromSeed(
      surface: design.semanticColors.bgSecondary,
      seedColor: design.semanticColors.bgAccent,
    );
  }

  ElevatedButtonThemeData _elevatedButtonThemeData() {
    return ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
        textStyle: design.typography.subHeading.copyWith(
          color: design.semanticColors.textPrimary,
        ),
        padding: const EdgeInsets.symmetric(horizontal: 25.0, vertical: 16),
        backgroundColor: design.semanticColors.bgAccent,
        foregroundColor: design.semanticColors.textPrimaryOnColor,
        disabledBackgroundColor: design.semanticColors.bgDisabled,
        disabledForegroundColor: design.semanticColors.textDisabled,
      ),
    );
  }

  InputDecorationTheme _inputDecorationTheme() {
    final border = OutlineInputBorder(
      borderRadius: BorderRadius.circular(8),
      borderSide: BorderSide(color: design.semanticColors.borderSecondary, width: 1.0),
    );

    return InputDecorationTheme(
      constraints: const BoxConstraints(maxHeight: 50),
      isDense: true,
      contentPadding: const EdgeInsets.all(10),
      border: border,
      focusedBorder: border,
      enabledBorder: border,
      hintStyle: design.typography.body.copyWith(
        color: design.semanticColors.textSecondary,
      ),
      floatingLabelBehavior: FloatingLabelBehavior.never,
    );
  }

  TextButtonThemeData _textButtonThemeData() {
    return TextButtonThemeData(
      style: TextButton.styleFrom(
        side: BorderSide.none,
        overlayColor: design.colors.transparent,
        textStyle: design.typography.body.copyWith(
          color: design.semanticColors.textPrimary,
        ),
        foregroundColor: design.semanticColors.textAccent,
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 16),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(4),
          side: BorderSide(
            color: design.semanticColors.borderPrimary,
            width: 1,
          ),
        ),
      ),
    );
  }

  OutlinedButtonThemeData _outlinedButtonThemeData() {
    return OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        side: BorderSide(color: design.semanticColors.borderPrimary),
        textStyle: design.typography.subHeading.copyWith(
          color: design.semanticColors.textPrimary,
        ),
        foregroundColor: design.semanticColors.textPrimary,
        padding: const EdgeInsets.symmetric(horizontal: 25.0, vertical: 16),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
        disabledBackgroundColor: design.semanticColors.bgDisabled,
        disabledForegroundColor: design.semanticColors.textDisabled,
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
      color: design.semanticColors.borderSecondary,
      thickness: 1,
    );
  }

  CheckboxThemeData _checkboxThemeData() {
    return CheckboxThemeData(
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
      overlayColor: WidgetStateProperty.all(design.colors.transparent),
      checkColor: WidgetStateProperty.resolveWith<Color?>((states) {
        if (states.contains(WidgetState.selected) &&
            states.contains(WidgetState.disabled)) {
          return design.semanticColors.bgDisabled;
        }
        if (states.contains(WidgetState.selected)) {
          return design.semanticColors.bgSecondary;
        }
        if (states.contains(WidgetState.disabled)) {
          return design.semanticColors.bgDisabled;
        }
        return design.semanticColors.bgSecondary;
      }),
      fillColor: WidgetStateProperty.resolveWith<Color?>((states) {
        if (states.contains(WidgetState.selected) &&
            states.contains(WidgetState.disabled)) {
          return design.semanticColors.bgDisabled;
        }
        if (states.contains(WidgetState.selected)) {
          return design.semanticColors.bgAccent;
        }
        if (states.contains(WidgetState.disabled)) {
          return design.semanticColors.bgSecondary;
        }
        return design.semanticColors.bgSecondary;
      }),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
      side: WidgetStateBorderSide.resolveWith((states) {
        if (states.contains(WidgetState.selected) &&
            states.contains(WidgetState.disabled)) {
          return BorderSide(width: 1, color: design.semanticColors.bgDisabled);
        }
        if (states.contains(WidgetState.selected)) {
          return BorderSide(width: 1, color: design.semanticColors.bgAccent);
        }
        if (states.contains(WidgetState.disabled)) {
          return BorderSide(width: 1, color: design.semanticColors.bgDisabled);
        }
        return BorderSide(width: 1, color: design.semanticColors.borderInput);
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
      borderColor: design.semanticColors.borderSecondary,
      verticalSpaceSmall: 8,
      verticalSpaceMedium: 16,
      verticalSpaceLarge: 24,
      horizontalSpaceSmall: 8,
      horizontalSpace: 16,
      textErrorColor: design.semanticColors.textCritical,
      successColor: design.semanticColors.textSuccess,
      flagsBorderSize: 2,
      overlayBackgroundColor: design.colors.neutral1000.withAlpha(127),
      caption: design.typography.subBody.copyWith(
        color: design.semanticColors.textSecondary,
      ),
      captionRegularGray171: TextStyle(
        fontSize: 12,
        color: design.semanticColors.textSecondary,
        fontWeight: FontWeight.w400,
      ),
      captionStrong: design.typography.captionMedium.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      bodyStrong: design.typography.subHeading.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      body: design.typography.body.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      subtitleStrong: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      linkButton: TextStyle(
        fontSize: 12,
        color: design.semanticColors.bgAccent,
        fontWeight: FontWeight.w400,
      ),
      title: design.typography.display.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      trailingIconSize: 32,
      backgroundColor: design.semanticColors.bgSecondary,
      areaBackgroundColor: design.semanticColors.bgSecondaryActive,
      area: design.semanticColors.bgSecondaryActive,
      dividerColor: design.semanticColors.borderSecondary,
      disabledOpacity: 0.5,
      linkNormal: design.typography.body.copyWith(
        color: design.semanticColors.textAccent,
      ),
      linkSmall: design.typography.subBody.copyWith(
        color: design.semanticColors.textAccent,
      ),
      textDisabled: design.typography.body.copyWith(
        color: design.semanticColors.textDisabled,
      ),
    );
  }

  VpnStatusCardTheme _vpnStatusCardThemeExt() {
    return VpnStatusCardTheme(
      height: 150,
      maxConnectButtonWidth: 408,
      primaryFont: design.typography.captionMedium.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      secondaryFont: design.typography.subHeading.copyWith(
        color: design.semanticColors.textPrimary,
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
      searchHintStyle: design.typography.body.copyWith(
        color: design.semanticColors.textSecondary,
      ),
      searchErrorStyle: design.typography.body.copyWith(
        color: design.semanticColors.textCritical,
      ),
      obfuscationSearchWarningStyle: design.typography.body.copyWith(
        color: design.semanticColors.textSecondary,
      ),
      obfuscatedItemBackgroundColor: design.semanticColors.bgSecondaryActive,
    );
  }

  SettingsTheme _settingsThemeExt() {
    return SettingsTheme(
      currentPageNameStyle: design.typography.heading.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      parentPageStyle: design.typography.heading.copyWith(
        color: design.semanticColors.textSecondary,
      ),
      itemTitleStyle: design.typography.body.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      itemSubtitleStyle: design.typography.subBody.copyWith(
        color: design.semanticColors.textSecondary,
      ),
      vpnStatusStyle: design.typography.captionMedium.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      textInputWidth: 220,
      otherProductsTitle: design.typography.body.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      otherProductsSubtitle: design.typography.body.copyWith(
        color: design.semanticColors.textSecondary,
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
        textStyle: design.typography.body.copyWith(
          color: design.semanticColors.textPrimary,
        ),
        disabledTextStyle: design.typography.body.copyWith(
          color: design.semanticColors.textDisabled,
        ),
      ),
      slider: OnOffSliderTheme(
        width: 38,
        height: 18,
        on: SwitchOnOffProps(
          leftOffset: 21,
          rightOffset: 0,
          color: design.colors.neutral0,
          borderColor: design.semanticColors.bgAccent,
          backgroundColor: design.semanticColors.bgAccent,
        ),
        off: SwitchOnOffProps(
          leftOffset: 0.8,
          rightOffset: 21,
          color: design.semanticColors.textPrimary,
          borderColor: design.semanticColors.borderInput,
          backgroundColor: design.semanticColors.bgSecondary,
        ),
        disabledOn: SwitchOnOffProps(
          leftOffset: 0.8,
          rightOffset: 21,
          color: design.semanticColors.bgTertiary,
          borderColor: design.semanticColors.bgDisabled,
          backgroundColor: design.semanticColors.bgDisabled,
        ),
        disabledOff: SwitchOnOffProps(
          leftOffset: 0.8,
          rightOffset: 21,
          color: design.semanticColors.bgDisabled,
          borderColor: design.semanticColors.borderPrimary,
          backgroundColor: design.colors.transparent,
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
          fillColor: design.semanticColors.bgAccent,
          borderColor: design.semanticColors.bgAccent,
          dotColor: design.colors.neutral0,
          dotHeight: 6,
          dotWidth: 6,
          borderWidth: 6,
        ),
        off: RadioOnOffProps(
          fillColor: design.semanticColors.bgSecondary,
          borderColor: design.semanticColors.borderInput,
          dotColor: design.colors.transparent,
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
      errorStyle: design.typography.subBody.copyWith(
        color: design.semanticColors.textCritical,
      ),
      textStyle: design.typography.body.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      enabled: EnabledStyle(
        borderColor: design.semanticColors.borderPrimary,
        borderWidth: 1,
      ),
      focused: FocusedStyle(
        borderColor: design.semanticColors.bgAccent,
        borderWidth: 2,
      ),
      error: ErrorStyle(
        borderColor: design.semanticColors.borderCritical,
        borderWidth: 1,
      ),
      focusedError: FocusedErrorStyle(
        borderColor: design.semanticColors.borderCritical,
        borderWidth: 1,
      ),
      icon: IconStyle(
        color: design.semanticColors.textPrimary,
        hoverColor: design.colors.transparent,
      ),
    );
  }

  LoadingIndicatorTheme _loadingIndicatorThemeExt() {
    return LoadingIndicatorTheme(
      color: design.semanticColors.bgAccent,
      strokeWidth: 2,
    );
  }

  CopyFieldTheme _copyFieldThemeExt() {
    return CopyFieldTheme(
      borderRadius: 4,
      commandTextStyle: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      descriptionTextStyle: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
    );
  }

  SupportLinkTheme _supportLinkThemeExt() {
    return SupportLinkTheme(
      textStyle: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      urlColor: design.semanticColors.bgAccent,
    );
  }

  LoginFormTheme _loginFormThemeExt() {
    return LoginFormTheme(
      titleStyle: design.typography.display.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      checkboxDescStyle: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      width: 424,
      height: 348,
      progressIndicator: LoginButtonProgressIndicatorTheme(
        height: 16,
        width: 16,
        stroke: 1.5,
        color: design.semanticColors.bgSecondary,
      ),
    );
  }

  InlineLoadingIndicatorTheme _inlineLoadingIndicatorThemeExt() {
    return InlineLoadingIndicatorTheme(
      width: 15,
      height: 15,
      stroke: 2,
      color: design.semanticColors.bgAccent,
      alternativeColor: design.colors.neutral0,
    );
  }

  AutoconnectPanelTheme _autoconnectPanelTheme() {
    return AutoconnectPanelTheme(
      primaryFont: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      secondaryFont: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      iconSize: 37,
      loaderSize: 35,
    );
  }

  CustomDnsTheme _customDnsThemeExt() {
    return CustomDnsTheme(
      formBackground: design.semanticColors.bgSecondaryActive,
      dnsInputWidth: 300,
      dividerColor: design.semanticColors.borderPrimary,
    );
  }

  AllowListTheme _allowListThemeExt() {
    return AllowListTheme(
      labelStyle: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      addCardBackground: design.semanticColors.bgSecondaryActive,
      tableItemsStyle: design.typography.body.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      tableHeaderStyle: design.typography.captionMedium.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      dividerColor: design.semanticColors.borderPrimary,
      listItemBackgroundColor: design.semanticColors.bgSecondaryActive,
    );
  }

  DropdownTheme _dropdownThemeExt() {
    return DropdownTheme(
      color: design.semanticColors.bgSecondary,
      borderRadius: 4,
      borderColor: design.semanticColors.borderPrimary,
      focusBorderColor: design.semanticColors.borderAccent,
      errorBorderColor: design.semanticColors.borderCritical,
      borderWidth: 1,
      horizontalPadding: 8,
    );
  }

  TooltipThemeData _tooltipThemeData() {
    return TooltipThemeData(
      decoration: BoxDecoration(
        color: design.semanticColors.bgSecondary, //
        border: Border.all(
          color: design.semanticColors.borderPrimary,
          width: 1,
        ),
        borderRadius: BorderRadius.all(Radius.circular(5)),
      ),
      textStyle: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
    );
  }

  InteractiveListViewTheme _interactiveListViewThemeExt() {
    return InteractiveListViewTheme(
      borderRadius: 4,
      borderColor: design.semanticColors.borderPrimary,
      focusBorderColor: design.semanticColors.borderAccent,
      borderWidth: 1,
    );
  }

  ErrorScreenTheme _errorScreenThemeExt() {
    return ErrorScreenTheme(
      titleTextStyle: design.typography.subHeading.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      descriptionTextStyle: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
    );
  }

  ConsentScreenTheme _consentScreenThemeExt() {
    return ConsentScreenTheme(
      width: 600,
      height: 430,
      overlayColor: design.colors.neutral1000.withAlpha(125),
      titleTextStyle: design.typography.display.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      bodyTextStyle: design.typography.body.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      titleBarTextStyle: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      padding: 40,
      listItemTitle: design.typography.body.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      listItemSubtitle: design.typography.subBody.copyWith(
        color: design.semanticColors.textPrimary,
      ),
      titleBarWidth: 45,
    );
  }
}
