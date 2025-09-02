import 'package:flutter/material.dart';
import 'package:nordvpn/theme/allow_list_theme.dart';
import 'package:nordvpn/theme/app_theme.dart';
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
import 'package:nordvpn/theme/ux_colors.dart';
import 'package:nordvpn/theme/ux_fonts.dart';
import 'package:nordvpn/theme/vpn_status_card_theme.dart';

ThemeData lightTheme() {
  return NordVpnTheme(ThemeMode.light).data();
}

ThemeData darkTheme() {
  return NordVpnTheme(ThemeMode.dark).data();
}

final class NordVpnTheme {
  final ThemeMode mode;
  final UXColors uxColors;
  final UXFonts uxFonts;

  NordVpnTheme(this.mode) : uxColors = UXColors(mode), uxFonts = UXFonts(mode);

  ThemeData data() {
    final data = mode == ThemeMode.light
        ? ThemeData.light(useMaterial3: true)
        : ThemeData.dark(useMaterial3: true);
    return data.copyWith(
      scaffoldBackgroundColor: uxColors.backgroundSecondary,
      disabledColor: uxColors.textDisabled,
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
      backgroundColor: uxColors.backgroundSecondary,
      labelType: NavigationRailLabelType.all,
      useIndicator: true,
      selectedLabelTextStyle: uxFonts.body,
      unselectedLabelTextStyle: uxFonts.body,
      indicatorColor: uxColors.fillGreySecondary,
    );
  }

  NavigationBarThemeData _navigationBarTheme() {
    return NavigationBarThemeData(
      indicatorColor: uxColors.fillGreySecondary,
      backgroundColor: uxColors.backgroundSecondary,
      surfaceTintColor: Colors.transparent,
      labelTextStyle: WidgetStateProperty.all(uxFonts.body),
    );
  }

  TabBarThemeData _tabBarTheme() {
    return TabBarThemeData(
      indicator: UnderlineTabIndicator(
        borderSide: BorderSide(width: 2, color: uxColors.fillAccentPrimary),
      ),
      tabAlignment: TabAlignment.start,
      indicatorSize: TabBarIndicatorSize.tab,
      indicatorColor: uxColors.fillAccentPrimary,
      dividerColor: Colors.transparent,
      dividerHeight: 0,
      labelStyle: uxFonts.caption.copyWith(color: uxColors.textPrimary),
      unselectedLabelStyle: uxFonts.captionTransparent_60,
    );
  }

  AppBarTheme _appBarTheme() {
    return AppBarTheme(
      toolbarHeight: 70,
      shape: Border(bottom: BorderSide(color: uxColors.strokeMedium, width: 1)),
      surfaceTintColor: Colors.transparent,
      centerTitle: true,
      backgroundColor: uxColors.backgroundPrimary,
    );
  }

  ColorScheme _colorScheme() {
    return ColorScheme.fromSeed(
      surface: uxColors.backgroundPrimary,
      seedColor: uxColors.fillAccentPrimary,
    );
  }

  ElevatedButtonThemeData _elevatedButtonThemeData() {
    return ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
        textStyle: uxFonts.bodyStrong,
        padding: const EdgeInsets.symmetric(horizontal: 25.0, vertical: 16),
        backgroundColor: uxColors.fillAccentPrimary,
        foregroundColor: uxColors.textOnAccent,
        disabledBackgroundColor: uxColors.fillGreyDisabled,
        disabledForegroundColor: uxColors.textDisabled,
      ),
    );
  }

  InputDecorationTheme _inputDecorationTheme() {
    final border = OutlineInputBorder(
      borderRadius: BorderRadius.circular(8),
      borderSide: BorderSide(color: uxColors.strokeMedium, width: 1.0),
    );

    return InputDecorationTheme(
      constraints: const BoxConstraints(maxHeight: 50),
      isDense: true,
      contentPadding: const EdgeInsets.all(10),
      border: border,
      focusedBorder: border,
      enabledBorder: border,
      hintStyle: uxFonts.body.copyWith(color: uxColors.textSecondary),
      floatingLabelBehavior: FloatingLabelBehavior.never,
    );
  }

  TextButtonThemeData _textButtonThemeData() {
    return TextButtonThemeData(
      style: TextButton.styleFrom(
        side: BorderSide.none,
        overlayColor: Colors.transparent,
        textStyle: uxFonts.body,
        foregroundColor: uxColors.textAccentPrimary,
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 16),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(4),
          side: BorderSide(color: uxColors.strokeSoft, width: 1),
        ),
      ),
    );
  }

  OutlinedButtonThemeData _outlinedButtonThemeData() {
    return OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        side: BorderSide(color: uxColors.strokeSoft),
        textStyle: uxFonts.bodyStrong,
        foregroundColor: uxColors.textPrimary,
        padding: const EdgeInsets.symmetric(horizontal: 25.0, vertical: 16),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
        disabledBackgroundColor: uxColors.fillGreyDisabled,
        disabledForegroundColor: uxColors.textDisabled,
      ),
    );
  }

  IconButtonThemeData _iconButtonThemeData() {
    return IconButtonThemeData(
      style: IconButton.styleFrom(padding: const EdgeInsets.all(10)),
    );
  }

  DividerThemeData _dividerTheme() {
    return DividerThemeData(color: uxColors.strokeDivider, thickness: 1);
  }

  CheckboxThemeData _checkboxThemeData() {
    return CheckboxThemeData(
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
      overlayColor: WidgetStateProperty.all(Colors.transparent),
      checkColor: WidgetStateProperty.resolveWith<Color?>((states) {
        if (states.contains(WidgetState.selected) &&
            states.contains(WidgetState.disabled)) {
          return uxColors.fillGreyDisabled;
        }
        if (states.contains(WidgetState.selected)) {
          return uxColors.fillGreyPrimary;
        }
        if (states.contains(WidgetState.disabled)) {
          return uxColors.fillGreyDisabled;
        }
        return uxColors.fillGreyPrimary;
      }),
      fillColor: WidgetStateProperty.resolveWith<Color?>((states) {
        if (states.contains(WidgetState.selected) &&
            states.contains(WidgetState.disabled)) {
          return uxColors.fillGreyDisabled;
        }
        if (states.contains(WidgetState.selected)) {
          return uxColors.fillAccentPrimary;
        }
        if (states.contains(WidgetState.disabled)) {
          return uxColors.fillGreyPrimary;
        }
        return uxColors.fillGreyPrimary;
      }),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
      side: WidgetStateBorderSide.resolveWith((states) {
        if (states.contains(WidgetState.selected) &&
            states.contains(WidgetState.disabled)) {
          return BorderSide(width: 1, color: uxColors.fillGreyDisabled);
        }
        if (states.contains(WidgetState.selected)) {
          return BorderSide(width: 1, color: uxColors.fillAccentPrimary);
        }
        if (states.contains(WidgetState.disabled)) {
          return BorderSide(width: 1, color: uxColors.fillGreyDisabled);
        }
        return BorderSide(width: 1, color: uxColors.strokeControlPrimary);
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
      borderColor: uxColors.strokeMedium,
      verticalSpaceSmall: 8,
      verticalSpaceMedium: 16,
      verticalSpaceLarge: 24,
      horizontalSpaceSmall: 8,
      horizontalSpace: 16,
      textErrorColor: uxColors.textCaution,
      successColor: uxColors.textSuccess,
      flagsBorderSize: 2,
      overlayBackgroundColor: uxColors.backgroundOverlay,
      caption: uxFonts.caption,
      captionRegularGray171: TextStyle(
        fontSize: 12,
        color: uxColors.textSecondary,
        fontWeight: FontWeight.w400,
      ),
      captionStrong: uxFonts.captionStrong,
      bodyStrong: uxFonts.bodyStrong,
      body: uxFonts.body,
      subtitleStrong: uxFonts.caption,
      linkButton: TextStyle(
        fontSize: 12,
        color: uxColors.fillAccentPrimary,
        fontWeight: FontWeight.w400,
      ),
      title: uxFonts.title,
      trailingIconSize: 32,
      backgroundColor: uxColors.backgroundPrimary,
      areaBackgroundColor: uxColors.fillGreyTertiary,
      area: uxColors.fillGreyQuaternary,
      dividerColor: uxColors.strokeMedium,
      disabledOpacity: 0.5,
      linkNormal: uxFonts.linkNormal,
      linkSmall: uxFonts.linkSmall,
      textDisabled: uxFonts.textDisabled,
    );
  }

  VpnStatusCardTheme _vpnStatusCardThemeExt() {
    return VpnStatusCardTheme(
      height: 150,
      maxConnectButtonWidth: 408,
      primaryFont: uxFonts.captionStrong,
      secondaryFont: uxFonts.bodyStrong,
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
      searchHintStyle: uxFonts.body.copyWith(color: uxColors.textSecondary),
      searchErrorStyle: uxFonts.body.copyWith(color: uxColors.textCaution),
      obfuscationSearchWarningStyle: uxFonts.body.copyWith(
        color: uxColors.textSecondary,
      ),
      obfuscatedItemBackgroundColor: uxColors.fillGreyQuaternary,
    );
  }

  SettingsTheme _settingsThemeExt() {
    return SettingsTheme(
      currentPageNameStyle: uxFonts.subtitle,
      parentPageStyle: uxFonts.subtitle.copyWith(color: uxColors.textSecondary),
      itemTitleStyle: uxFonts.body,
      itemSubtitleStyle: uxFonts.caption,
      vpnStatusStyle: uxFonts.captionStrong,
      textInputWidth: 220,
      otherProductsTitle: uxFonts.body,
      otherProductsSubtitle: uxFonts.body.copyWith(
        color: uxColors.textSecondary,
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
        textStyle: uxFonts.body,
        disabledTextStyle: uxFonts.body.copyWith(color: uxColors.textDisabled),
      ),
      slider: OnOffSliderTheme(
        width: 38,
        height: 18,
        on: SwitchOnOffProps(
          leftOffset: 21,
          rightOffset: 0,
          color: uxColors.fillWhiteFixed,
          borderColor: uxColors.fillAccentPrimary,
          backgroundColor: uxColors.fillAccentPrimary,
        ),
        off: SwitchOnOffProps(
          leftOffset: 0.8,
          rightOffset: 21,
          color: uxColors.strokeControlPrimary,
          borderColor: uxColors.strokeControlPrimary,
          backgroundColor: uxColors.fillGreyPrimary,
        ),
        disabledOn: SwitchOnOffProps(
          leftOffset: 0.8,
          rightOffset: 21,
          color: uxColors.fillGreyDisabled,
          borderColor: uxColors.fillGreyDisabled,
          backgroundColor: uxColors.fillGreyDisabled,
        ),
        disabledOff: SwitchOnOffProps(
          leftOffset: 0.8,
          rightOffset: 21,
          color: uxColors.fillGreyDisabled,
          borderColor: uxColors.strokeDisabled,
          backgroundColor: Colors.transparent,
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
          fillColor: uxColors.fillAccentPrimary,
          borderColor: uxColors.fillAccentPrimary,
          dotColor: uxColors.fillWhiteFixed,
          dotHeight: 6,
          dotWidth: 6,
          borderWidth: 6,
        ),
        off: RadioOnOffProps(
          fillColor: uxColors.fillGreyPrimary,
          borderColor: uxColors.strokeControlPrimary,
          dotColor: Colors.transparent,
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
      errorStyle: uxFonts.bodyCaution,
      textStyle: uxFonts.body,
      enabled: EnabledStyle(borderColor: uxColors.strokeSoft, borderWidth: 1),
      focused: FocusedStyle(
        borderColor: uxColors.fillAccentPrimary,
        borderWidth: 2,
      ),
      error: ErrorStyle(borderColor: uxColors.strokeCaution, borderWidth: 1),
      focusedError: FocusedErrorStyle(
        borderColor: uxColors.strokeCaution,
        borderWidth: 1,
      ),
      icon: IconStyle(
        color: uxColors.iconPrimary,
        hoverColor: Colors.transparent,
      ),
    );
  }

  LoadingIndicatorTheme _loadingIndicatorThemeExt() {
    return LoadingIndicatorTheme(
      color: uxColors.fillAccentPrimary,
      strokeWidth: 2,
    );
  }

  CopyFieldTheme _copyFieldThemeExt() {
    return CopyFieldTheme(
      borderRadius: 4,
      commandTextStyle: uxFonts.caption,
      descriptionTextStyle: uxFonts.caption.copyWith(
        color: uxColors.textPrimary,
      ),
    );
  }

  SupportLinkTheme _supportLinkThemeExt() {
    return SupportLinkTheme(
      textStyle: uxFonts.caption,
      urlColor: uxColors.fillAccentPrimary,
    );
  }

  LoginFormTheme _loginFormThemeExt() {
    return LoginFormTheme(
      titleStyle: uxFonts.title,
      checkboxDescStyle: uxFonts.caption,
      width: 424,
      height: 348,
      progressIndicator: LoginButtonProgressIndicatorTheme(
        height: 16,
        width: 16,
        stroke: 1.5,
        color: uxColors.fillGreyPrimary,
      ),
    );
  }

  InlineLoadingIndicatorTheme _inlineLoadingIndicatorThemeExt() {
    return InlineLoadingIndicatorTheme(
      width: 15,
      height: 15,
      stroke: 2,
      color: uxColors.fillAccentPrimary,
      alternativeColor: uxColors.fillWhiteFixed,
    );
  }

  AutoconnectPanelTheme _autoconnectPanelTheme() {
    return AutoconnectPanelTheme(
      primaryFont: uxFonts.caption,
      secondaryFont: uxFonts.caption,
      iconSize: 37,
      loaderSize: 35,
    );
  }

  CustomDnsTheme _customDnsThemeExt() {
    return CustomDnsTheme(
      formBackground: uxColors.fillGreyQuaternary,
      dnsInputWidth: 300,
      dividerColor: uxColors.strokeSoft,
    );
  }

  AllowListTheme _allowListThemeExt() {
    return AllowListTheme(
      labelStyle: uxFonts.caption,
      addCardBackground: uxColors.fillGreyQuaternary,
      tableItemsStyle: uxFonts.body,
      tableHeaderStyle: uxFonts.captionStrong,
      dividerColor: uxColors.strokeSoft,
      listItemBackgroundColor: uxColors.fillGreyQuaternary,
    );
  }

  DropdownTheme _dropdownThemeExt() {
    return DropdownTheme(
      color: uxColors.fillGreyPrimary,
      borderRadius: 4,
      borderColor: uxColors.strokeSoft,
      focusBorderColor: uxColors.strokeAccent,
      errorBorderColor: uxColors.strokeCaution,
      borderWidth: 1,
      horizontalPadding: 8,
    );
  }

  TooltipThemeData _tooltipThemeData() {
    return TooltipThemeData(
      decoration: BoxDecoration(
        color: uxColors.backgroundSecondary, //
        border: Border.all(color: uxColors.strokeSoft, width: 1),
        borderRadius: BorderRadius.all(Radius.circular(5)),
      ),
      textStyle: uxFonts.caption.copyWith(color: uxColors.textPrimary),
    );
  }

  InteractiveListViewTheme _interactiveListViewThemeExt() {
    return InteractiveListViewTheme(
      borderRadius: 4,
      borderColor: uxColors.strokeSoft,
      focusBorderColor: uxColors.strokeAccent,
      borderWidth: 1,
    );
  }

  ErrorScreenTheme _errorScreenThemeExt() {
    return ErrorScreenTheme(
      titleTextStyle: uxFonts.bodyStrong,
      descriptionTextStyle: uxFonts.caption.copyWith(
        color: uxColors.textPrimary,
      ),
    );
  }

  ConsentScreenTheme _consentScreenThemeExt() {
    return ConsentScreenTheme(
      width: 600,
      height: 400,
      overlayColor: uxColors.backgroundOverlay,
      titleTextStyle: uxFonts.title,
      bodyTextStyle: uxFonts.body,
      titleBarTextStyle: uxFonts.caption.copyWith(color: uxColors.textPrimary),
      padding: 40,
      listItemTitle: uxFonts.body,
      listItemSubtitle: uxFonts.caption,
      titleBarWidth: 45,
    );
  }
}
