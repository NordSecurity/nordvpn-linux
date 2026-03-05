import 'package:flutter/material.dart';

// --- 1. Core Colors ---
// This class defines the base color palette using hexadecimal values.
class AppCoreColors {
  AppCoreColors._(); // Private constructor to prevent instantiation.

  static const Color transparent = Color(0x00000000); // Represents a fully transparent color.

  // Neutral shades from 0 (white) to 1000 (darkest black).
  static const Color neutral0 = Color(0xFFFFFFFF);
  static const Color neutral100 = Color(0xFFF7F7F8);
  static const Color neutral150 = Color(0xFFEDEDED);
  static const Color neutral200 = Color(0xFFE2E2E4);
  static const Color neutral300 = Color(0xFFC8C9CB);
  static const Color neutral400 = Color(0xFFB2B2B3);
  static const Color neutral500 = Color(0xFF909192);
  static const Color neutral600 = Color(0xFF696A6D);
  static const Color neutral700 = Color(0xFF4F5054);
  static const Color neutral800 = Color(0xFF3E3F42);
  static const Color neutral900 = Color(0xFF2A2A2D);
  static const Color neutral950 = Color(0xFF1D1E20);
  static const Color neutral1000 = Color(0xFF141415);

  // Blue shades.
  static const Color blue100 = Color(0xFFF3F7FC);
  static const Color blue200 = Color(0xFFD4E2F7);
  static const Color blue300 = Color(0xFFB5CDF5);
  static const Color blue400 = Color(0xFF8CAEF8);
  static const Color blue500 = Color(0xFF6B90FA);
  static const Color blue600 = Color(0xFF3E5FFF);
  static const Color blue700 = Color(0xFF243DCC);
  static const Color blue800 = Color(0xFF263482);
  static const Color blue900 = Color(0xFF22294F);
  static const Color blue950 = Color(0xFF1A1F3D);
  static const Color blue1000 = Color(0xFF12162B);

  // Green shades.
  static const Color green100 = Color(0xFFECF9EE);
  static const Color green200 = Color(0xFFB7F2C5);
  static const Color green300 = Color(0xFF81E4A2);
  static const Color green400 = Color(0xFF37C871);
  static const Color green500 = Color(0xFF0EA464);
  static const Color green600 = Color(0xFF0A8550);
  static const Color green700 = Color(0xFF075F3C);
  static const Color green800 = Color(0xFF05472B);
  static const Color green900 = Color(0xFF043420);
  static const Color green950 = Color(0xFF032617);
  static const Color green1000 = Color(0xFF02180E);

  // Yellow shades.
  static const Color yellow100 = Color(0xFFFFF6DB);
  static const Color yellow200 = Color(0xFFFEE071);
  static const Color yellow300 = Color(0xFFFAC900);
  static const Color yellow400 = Color(0xFFD1A900);
  static const Color yellow500 = Color(0xFFAE8604);
  static const Color yellow600 = Color(0xFF8E6C10);
  static const Color yellow700 = Color(0xFF654A0B);
  static const Color yellow800 = Color(0xFF4E3709);
  static const Color yellow900 = Color(0xFF3C2A07);
  static const Color yellow950 = Color(0xFF2D2006);
  static const Color yellow1000 = Color(0xFF1B1509);

  // Red shades.
  static const Color red100 = Color(0xFFFCEFEE);
  static const Color red200 = Color(0xFFF9D7D3);
  static const Color red300 = Color(0xFFF6BEB9);
  static const Color red400 = Color(0xFFF29086);
  static const Color red500 = Color(0xFFEC6255);
  static const Color red600 = Color(0xFFE02F1F);
  static const Color red700 = Color(0xFF9E1C10);
  static const Color red800 = Color(0xFF771209);
  static const Color red900 = Color(0xFF5A0E07);
  static const Color red950 = Color(0xFF450C07);
  static const Color red1000 = Color(0xFF2F0704);
}

// --- 2. Semantic Colors ---
// These classes define colors used for specific UI purposes, organized by theme mode.
// They reference the core colors defined in AppCoreColors for consistency.
class AppSemanticColorsLight {
  AppSemanticColorsLight._(); // Private constructor to prevent instantiation.

  // Background colors for the light theme.
  static const Color bgPrimary = AppCoreColors.neutral100;
  static const Color bgSecondary = AppCoreColors.neutral0;
  static const Color bgTertiary = AppCoreColors.neutral150;
  static const Color bgAccent = AppCoreColors.blue600;
  static const Color bgAccentSubtle = AppCoreColors.blue100;
  static const Color bgDisabled = AppCoreColors.neutral300;
  static const Color bgSuccess = AppCoreColors.green600;
  static const Color bgCritical = AppCoreColors.red600;
  static const Color bgWarning = AppCoreColors.yellow300;
  static const Color bgSuccessSubtle = AppCoreColors.green100;
  static const Color bgWarningSubtle = AppCoreColors.yellow100;
  static const Color bgCriticalSubtle = AppCoreColors.red100;
  static const Color bgPrimaryActive = AppCoreColors.neutral150;
  static const Color bgSecondaryActive = AppCoreColors.neutral100;
  static const Color bgAccentActive = AppCoreColors.blue700;
  static const Color bgOverlay = Color(0x80141415);
  static const Color bgGlass = Color(0xB3F7F7F8); // rgba(247, 247, 248, 0.7)
  static const Color bgInverse = AppCoreColors.neutral950;
  static const Color bgInverseActive = AppCoreColors.neutral800;
  static const Color bgGradientPrimaryStart = AppCoreColors.neutral100; // Maps to core neutral-100
  static const Color bgGradientPrimaryEnd = Color(0x00F7F7F8); // #F7F7F8 with 0% opacity
  static const Color bgGradientSecondaryStart = AppCoreColors.neutral0; // Maps to core neutral-0
  static const Color bgGradientSecondaryEnd = Color(0x00FFFFFF); // #FFFFFF with 0% opacity
  static const Color bgSkeletonStart = AppCoreColors.neutral150; // Maps to core neutral-150
  static const Color bgSkeletonEnd = AppCoreColors.neutral200; // Maps to core neutral-200
  static const Color bgChartData1 = Color(0xFF6387EE); // #6387EE
  static const Color bgChartData2 = Color(0xFFEC6255); // #EC6255
  static const Color bgChartData3 = Color(0xFFBA8555); // #BA8555
  static const Color bgChartData4 = Color(0xFFE85E83); // #E85E83
  static const Color bgChartData5 = Color(0xFFE4700C); // #E4700C
  static const Color bgChartData6 = Color(0xFF3A9E85); // #3A9E85
  static const Color bgChartData7 = Color(0xFF967CC2); // #967CC2
  static const Color bgChartData8 = Color(0xFF909192); // #909192

  // Border colors for the light theme.
  static const Color borderPrimary = AppCoreColors.neutral300;
  static const Color borderSecondary = AppCoreColors.neutral200;
  static const Color borderInput = AppCoreColors.neutral500;
  static const Color borderAccent = AppCoreColors.blue600;
  static const Color borderSuccess = AppCoreColors.green400;
  static const Color borderWarning = AppCoreColors.yellow300;
  static const Color borderCritical = AppCoreColors.red400;
  static const Color borderAccentActive = AppCoreColors.blue700;

  // Text colors for the light theme.
  static const Color textPrimary = AppCoreColors.neutral900;
  static const Color textSecondary = AppCoreColors.neutral600;
  static const Color textAccent = AppCoreColors.blue600;
  static const Color textPrimaryOnColor = AppCoreColors.neutral0;
  static const Color textSecondaryOnColor = AppCoreColors.neutral100;
  static const Color textDisabled = AppCoreColors.neutral400;
  static const Color textSuccess = AppCoreColors.green700;
  static const Color textWarning = AppCoreColors.yellow700;
  static const Color textCritical = AppCoreColors.red700;
  static const Color textAccentActive = AppCoreColors.blue700;
}

class AppSemanticColorsDark {
  AppSemanticColorsDark._(); // Private constructor to prevent instantiation.

  // Background colors for the dark theme.
  static const Color bgPrimary = AppCoreColors.neutral1000;
  static const Color bgSecondary = AppCoreColors.neutral950;
  static const Color bgTertiary = AppCoreColors.neutral900;
  static const Color bgAccent = AppCoreColors.blue600;
  static const Color bgAccentSubtle = AppCoreColors.blue900;
  static const Color bgDisabled = AppCoreColors.neutral800;
  static const Color bgSuccess = AppCoreColors.green600;
  static const Color bgCritical = AppCoreColors.red700;
  static const Color bgWarning = AppCoreColors.yellow600;
  static const Color bgSuccessSubtle = AppCoreColors.green900;
  static const Color bgWarningSubtle = AppCoreColors.yellow900;
  static const Color bgCriticalSubtle = AppCoreColors.red900;
  static const Color bgPrimaryActive = AppCoreColors.neutral950;
  static const Color bgSecondaryActive = AppCoreColors.neutral900;
  static const Color bgAccentActive = AppCoreColors.blue700;
  static const Color bgOverlay = Color(0x80141415);
  static const Color bgGlass = Color(0xB3141415); // rgba(20, 20, 21, 0.7)
  static const Color bgInverse = AppCoreColors.neutral0;
  static const Color bgInverseActive = AppCoreColors.neutral200;
  static const Color bgGradientPrimaryStart = AppCoreColors.neutral1000; // Maps to core neutral-1000
  static const Color bgGradientPrimaryEnd = Color(0x00141415); // #141415 with 0% opacity
  static const Color bgGradientSecondaryStart = AppCoreColors.neutral950; // Maps to core neutral-950
  static const Color bgGradientSecondaryEnd = Color(0x001D1E20); // #1D1E20 with 0% opacity
  static const Color bgSkeletonStart = AppCoreColors.neutral900; // Maps to core neutral-900
  static const Color bgSkeletonEnd = AppCoreColors.neutral800; // Maps to core neutral-800
  static const Color bgChartData1 = Color(0xFF6387EE); // #6387EE
  static const Color bgChartData2 = Color(0xFFEC6255); // #EC6255
  static const Color bgChartData3 = Color(0xFFBA8555); // #BA8555
  static const Color bgChartData4 = Color(0xFFE85E83); // #E85E83
  static const Color bgChartData5 = Color(0xFFE4700C); // #E4700C
  static const Color bgChartData6 = Color(0xFF3A9E85); // #3A9E85
  static const Color bgChartData7 = Color(0xFF967CC2); // #967CC2
  static const Color bgChartData8 = Color(0xFF909192); // #909192

  // Border colors for the dark theme.
  static const Color borderPrimary = AppCoreColors.neutral700;
  static const Color borderSecondary = AppCoreColors.neutral800;
  static const Color borderInput = AppCoreColors.neutral600;
  static const Color borderAccent = AppCoreColors.blue600;
  static const Color borderSuccess = AppCoreColors.green600;
  static const Color borderWarning = AppCoreColors.yellow600;
  static const Color borderCritical = AppCoreColors.red600;
  static const Color borderAccentActive = AppCoreColors.blue700;

  // Text colors for the dark theme.
  static const Color textPrimary = AppCoreColors.neutral0;
  static const Color textSecondary = AppCoreColors.neutral500;
  static const Color textAccent = AppCoreColors.blue500;
  static const Color textPrimaryOnColor = AppCoreColors.neutral0;
  static const Color textSecondaryOnColor = AppCoreColors.neutral100;
  static const Color textDisabled = AppCoreColors.neutral700;
  static const Color textSuccess = AppCoreColors.green400;
  static const Color textWarning = AppCoreColors.yellow300;
  static const Color textCritical = AppCoreColors.red400;
  static const Color textAccentActive = AppCoreColors.blue400;
}

class SemanticColors {
  final bool isDark;

  const SemanticColors(this.isDark);

  Color get bgPrimary => isDark ? AppSemanticColorsDark.bgPrimary : AppSemanticColorsLight.bgPrimary;
  Color get bgSecondary => isDark ? AppSemanticColorsDark.bgSecondary : AppSemanticColorsLight.bgSecondary;
  Color get bgTertiary => isDark ? AppSemanticColorsDark.bgTertiary : AppSemanticColorsLight.bgTertiary;
  Color get bgAccent => isDark ? AppSemanticColorsDark.bgAccent : AppSemanticColorsLight.bgAccent;
  Color get bgAccentSubtle => isDark ? AppSemanticColorsDark.bgAccentSubtle : AppSemanticColorsLight.bgAccentSubtle;
  Color get bgDisabled => isDark ? AppSemanticColorsDark.bgDisabled : AppSemanticColorsLight.bgDisabled;
  Color get bgSuccess => isDark ? AppSemanticColorsDark.bgSuccess : AppSemanticColorsLight.bgSuccess;
  Color get bgCritical => isDark ? AppSemanticColorsDark.bgCritical : AppSemanticColorsLight.bgCritical;
  Color get bgWarning => isDark ? AppSemanticColorsDark.bgWarning : AppSemanticColorsLight.bgWarning;
  Color get bgSuccessSubtle => isDark ? AppSemanticColorsDark.bgSuccessSubtle : AppSemanticColorsLight.bgSuccessSubtle;
  Color get bgWarningSubtle => isDark ? AppSemanticColorsDark.bgWarningSubtle : AppSemanticColorsLight.bgWarningSubtle;
  Color get bgCriticalSubtle => isDark ? AppSemanticColorsDark.bgCriticalSubtle : AppSemanticColorsLight.bgCriticalSubtle;
  Color get bgPrimaryActive => isDark ? AppSemanticColorsDark.bgPrimaryActive : AppSemanticColorsLight.bgPrimaryActive;
  Color get bgSecondaryActive => isDark ? AppSemanticColorsDark.bgSecondaryActive : AppSemanticColorsLight.bgSecondaryActive;
  Color get bgAccentActive => isDark ? AppSemanticColorsDark.bgAccentActive : AppSemanticColorsLight.bgAccentActive;
  Color get bgOverlay => isDark ? AppSemanticColorsDark.bgOverlay : AppSemanticColorsLight.bgOverlay;
  Color get bgGlass => isDark ? AppSemanticColorsDark.bgGlass : AppSemanticColorsLight.bgGlass;
  Color get bgInverse => isDark ? AppSemanticColorsDark.bgInverse : AppSemanticColorsLight.bgInverse;
  Color get bgInverseActive => isDark ? AppSemanticColorsDark.bgInverseActive : AppSemanticColorsLight.bgInverseActive;
  Color get bgGradientPrimaryStart => isDark ? AppSemanticColorsDark.bgGradientPrimaryStart : AppSemanticColorsLight.bgGradientPrimaryStart;
  Color get bgGradientPrimaryEnd => isDark ? AppSemanticColorsDark.bgGradientPrimaryEnd : AppSemanticColorsLight.bgGradientPrimaryEnd;
  Color get bgGradientSecondaryStart => isDark ? AppSemanticColorsDark.bgGradientSecondaryStart : AppSemanticColorsLight.bgGradientSecondaryStart;
  Color get bgGradientSecondaryEnd => isDark ? AppSemanticColorsDark.bgGradientSecondaryEnd : AppSemanticColorsLight.bgGradientSecondaryEnd;
  Color get bgSkeletonStart => isDark ? AppSemanticColorsDark.bgSkeletonStart : AppSemanticColorsLight.bgSkeletonStart;
  Color get bgSkeletonEnd => isDark ? AppSemanticColorsDark.bgSkeletonEnd : AppSemanticColorsLight.bgSkeletonEnd;
  Color get bgChartData1 => isDark ? AppSemanticColorsDark.bgChartData1 : AppSemanticColorsLight.bgChartData1;
  Color get bgChartData2 => isDark ? AppSemanticColorsDark.bgChartData2 : AppSemanticColorsLight.bgChartData2;
  Color get bgChartData3 => isDark ? AppSemanticColorsDark.bgChartData3 : AppSemanticColorsLight.bgChartData3;
  Color get bgChartData4 => isDark ? AppSemanticColorsDark.bgChartData4 : AppSemanticColorsLight.bgChartData4;
  Color get bgChartData5 => isDark ? AppSemanticColorsDark.bgChartData5 : AppSemanticColorsLight.bgChartData5;
  Color get bgChartData6 => isDark ? AppSemanticColorsDark.bgChartData6 : AppSemanticColorsLight.bgChartData6;
  Color get bgChartData7 => isDark ? AppSemanticColorsDark.bgChartData7 : AppSemanticColorsLight.bgChartData7;
  Color get bgChartData8 => isDark ? AppSemanticColorsDark.bgChartData8 : AppSemanticColorsLight.bgChartData8;

  Color get borderPrimary => isDark ? AppSemanticColorsDark.borderPrimary : AppSemanticColorsLight.borderPrimary;
  Color get borderSecondary => isDark ? AppSemanticColorsDark.borderSecondary : AppSemanticColorsLight.borderSecondary;
  Color get borderInput => isDark ? AppSemanticColorsDark.borderInput : AppSemanticColorsLight.borderInput;
  Color get borderAccent => isDark ? AppSemanticColorsDark.borderAccent : AppSemanticColorsLight.borderAccent;
  Color get borderSuccess => isDark ? AppSemanticColorsDark.borderSuccess : AppSemanticColorsLight.borderSuccess;
  Color get borderWarning => isDark ? AppSemanticColorsDark.borderWarning : AppSemanticColorsLight.borderWarning;
  Color get borderCritical => isDark ? AppSemanticColorsDark.borderCritical : AppSemanticColorsLight.borderCritical;
  Color get borderAccentActive => isDark ? AppSemanticColorsDark.borderAccentActive : AppSemanticColorsLight.borderAccentActive;

  Color get textPrimary => isDark ? AppSemanticColorsDark.textPrimary : AppSemanticColorsLight.textPrimary;
  Color get textSecondary => isDark ? AppSemanticColorsDark.textSecondary : AppSemanticColorsLight.textSecondary;
  Color get textAccent => isDark ? AppSemanticColorsDark.textAccent : AppSemanticColorsLight.textAccent;
  Color get textPrimaryOnColor => isDark ? AppSemanticColorsDark.textPrimaryOnColor : AppSemanticColorsLight.textPrimaryOnColor;
  Color get textSecondaryOnColor => isDark ? AppSemanticColorsDark.textSecondaryOnColor : AppSemanticColorsLight.textSecondaryOnColor;
  Color get textDisabled => isDark ? AppSemanticColorsDark.textDisabled : AppSemanticColorsLight.textDisabled;
  Color get textSuccess => isDark ? AppSemanticColorsDark.textSuccess : AppSemanticColorsLight.textSuccess;
  Color get textWarning => isDark ? AppSemanticColorsDark.textWarning : AppSemanticColorsLight.textWarning;
  Color get textCritical => isDark ? AppSemanticColorsDark.textCritical : AppSemanticColorsLight.textCritical;
  Color get textAccentActive => isDark ? AppSemanticColorsDark.textAccentActive : AppSemanticColorsLight.textAccentActive;
}

// --- 3. Typography ---
// This class defines typography constants, including font sizes, weights,
// letter spacing, line heights, and pre-defined text styles for various elements.
class AppTypography {
  AppTypography._(); // Private constructor to prevent instantiation.

  // Font Families (assuming 'Inter' is added to your pubspec.yaml file)
  // Example for pubspec.yaml:
  // flutter:
  //   fonts:
  //     - family: Inter
  //       fonts:
  //         - asset: assets/fonts/Inter-Regular.ttf
  //         - asset: assets/fonts/Inter-Medium.ttf
  //           weight: 500
  //         - asset: assets/fonts/Inter-SemiBold.ttf
  //           weight: 600
  static const String fontFamilyHeading = 'Inter';
  static const String fontFamilyBody = 'Inter';

  // Font Sizes (converted from rem to logical pixels, assuming a base font size of 16px).
  static const double fontSize2xs = 0.6875 * 16; // ~11px
  static const double fontSizeXs = 0.75 * 16;    // 12px
  static const double fontSizeSm = 0.875 * 16;   // 14px
  static const double fontSizeMd = 1.0 * 16;     // 16px
  static const double fontSizeLg = 1.125 * 16;   // 18px
  static const double fontSizeXl = 1.25 * 16;    // 20px
  static const double fontSize2xl = 1.375 * 16;  // 22px
  static const double fontSize3xl = 1.5 * 16;    // 24px
  static const double fontSize4xl = 1.625 * 16;  // 26px
  static const double fontSize5xl = 1.75 * 16;   // 28px
  static const double fontSize6xl = 2.0 * 16;    // 32px
  static const double fontSize7xl = 2.5 * 16;    // 40px
  static const double fontSize8xl = 3.0 * 16;    // 48px
  static const double fontSize9xl = 3.5 * 16;    // 56px

  // Font Weights corresponding to the design tokens.
  static const FontWeight fontWeightNormal = FontWeight.w400;
  static const FontWeight fontWeightMedium = FontWeight.w500;
  static const FontWeight fontWeightBold = FontWeight.w600; // Typically Flutter's w600 is "semi-bold"

  // Letter Spacing (relative values in em).
  static const double letterSpacing2xs = -0.047;
  static const double letterSpacingXs = -0.031;
  static const double letterSpacingSm = -0.016;
  static const double letterSpacingMd = 0.0;
  static const double letterSpacingLg = 0.016;
  static const double letterSpacingXl = 0.031;
  static const double letterSpacing2xl = 0.047;

  // Pre-defined Text Styles.
  // Line heights are calculated as a factor (pixel height / effective font size).
  static const TextStyle display = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize5xl,
    letterSpacing: letterSpacingSm,
    height: 36 / (1.75 * 16), // 36px line-height
  );

  static const TextStyle heading = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize3xl,
    letterSpacing: letterSpacingMd,
    height: 32 / (1.5 * 16), // 32px line-height
  );

  static const TextStyle subHeading = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightMedium,
    fontSize: fontSizeMd,
    letterSpacing: letterSpacingMd,
    height: 22 / (1.0 * 16), // 22px line-height
  );

  static const TextStyle body = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightNormal,
    fontSize: fontSizeMd,
    letterSpacing: letterSpacingMd,
    height: 22 / (1.0 * 16), // 22px line-height
  );

  static const TextStyle subBody = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightNormal,
    fontSize: fontSizeSm,
    letterSpacing: letterSpacingMd,
    height: 20 / (0.875 * 16), // 20px line-height
  );

  static const TextStyle caption = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightNormal,
    fontSize: fontSizeXs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.75 * 16), // 18px line-height
  );

  static const TextStyle captionMedium = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightMedium,
    fontSize: fontSizeXs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.75 * 16), // 18px line-height
  );

  // Additional heading styles.
  static const TextStyle heading2xl = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize7xl,
    letterSpacing: letterSpacingSm,
    height: 48 / (2.5 * 16),
  );
  static const TextStyle headingXl = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize6xl,
    letterSpacing: letterSpacingSm,
    height: 40 / (2.0 * 16),
  );
  static const TextStyle headingLg = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize5xl,
    letterSpacing: letterSpacingSm,
    height: 36 / (1.75 * 16),
  );
  static const TextStyle headingMd = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize4xl,
    letterSpacing: letterSpacingMd,
    height: 34 / (1.625 * 16),
  );
  static const TextStyle headingSm = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize2xl,
    letterSpacing: letterSpacingMd,
    height: 30 / (1.375 * 16),
  );

  // Additional body styles.
  static const TextStyle bodyLg = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightNormal,
    fontSize: fontSizeLg,
    letterSpacing: letterSpacingMd,
    height: 24 / (1.125 * 16),
  );
  static const TextStyle bodyLgMedium = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightMedium,
    fontSize: fontSizeLg,
    letterSpacing: letterSpacingMd,
    height: 24 / (1.125 * 16),
  );
  static const TextStyle bodyLgBold = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightBold,
    fontSize: fontSizeLg,
    letterSpacing: letterSpacingMd,
    height: 24 / (1.125 * 16),
  );
  static const TextStyle bodyMd = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightNormal,
    fontSize: fontSizeMd,
    letterSpacing: letterSpacingMd,
    height: 22 / (1.0 * 16),
  );
  static const TextStyle bodyMdMedium = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightMedium,
    fontSize: fontSizeMd,
    letterSpacing: letterSpacingMd,
    height: 22 / (1.0 * 16),
  );
  static const TextStyle bodyMdBold = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: FontWeight.bold,
    fontSize: fontSizeMd,
    letterSpacing: letterSpacingMd,
    height: 22 / (1.0 * 16),
  );
  static const TextStyle bodySm = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightNormal,
    fontSize: fontSizeSm,
    letterSpacing: letterSpacingMd,
    height: 20 / (0.875 * 16),
  );
  static const TextStyle bodySmMedium = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightMedium,
    fontSize: fontSizeSm,
    letterSpacing: letterSpacingMd,
    height: 20 / (0.875 * 16),
  );
  static const TextStyle bodySmBold = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: FontWeight.bold,
    fontSize: fontSizeSm,
    letterSpacing: letterSpacingMd,
    height: 20 / (0.875 * 16),
  );
  static const TextStyle bodyXs = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: FontWeight.normal,
    fontSize: fontSizeXs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.75 * 16),
  );
  static const TextStyle bodyXsMedium = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightMedium,
    fontSize: fontSizeXs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.75 * 16),
  );
  static const TextStyle bodyXsBold = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: FontWeight.bold,
    fontSize: fontSizeXs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.75 * 16),
  );
  static const TextStyle body2xs = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: FontWeight.normal,
    fontSize: fontSize2xs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.6875 * 16),
  );
  static const TextStyle body2xsMedium = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightMedium,
    fontSize: fontSize2xs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.6875 * 16),
  );
  static const TextStyle body2xsBold = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: FontWeight.bold,
    fontSize: fontSize2xs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.6875 * 16),
  );
}

// --- 4. Spacing ---
// This class defines common spacing values, converted to logical pixels.
class AppSpacing {
  AppSpacing._(); // Private constructor to prevent instantiation.

  static const double spacing0 = 0.0;
  static const double spacing0_5 = 2.0;
  static const double spacing1 = 4.0;
  static const double spacing2 = 8.0;
  static const double spacing2_5 = 10.0;
  static const double spacing3 = 12.0;
  static const double spacing4 = 16.0;
  static const double spacing5 = 20.0;
  static const double spacing6 = 24.0;
  static const double spacing7 = 28.0;
  static const double spacing8 = 32.0;
  static const double spacing10 = 40.0;
  static const double spacing12 = 48.0;
  static const double spacing16 = 64.0;
  static const double spacing20 = 80.0;
  static const double spacing30 = 120.0;
}

// --- 5. Border Radius ---
// This class defines common border radius values for rounded corners.
class AppBorderRadius {
  AppBorderRadius._(); // Private constructor to prevent instantiation.

  static const BorderRadius none = BorderRadius.zero;
  static const BorderRadius xs = BorderRadius.all(Radius.circular(3.0));
  static const BorderRadius sm = BorderRadius.all(Radius.circular(6.0));
  static const BorderRadius md = BorderRadius.all(Radius.circular(12.0));
  static const BorderRadius lg = BorderRadius.all(Radius.circular(20.0));
  static const BorderRadius full = BorderRadius.all(Radius.circular(9999.0)); 
}

// --- 6. Border Width ---
// This class defines common border width values.
class AppBorderWidth {
  AppBorderWidth._(); // Private constructor to prevent instantiation.

  static const double none = 0.0;
  static const double sm = 0.5;
  static const double md = 1.0;
  static const double lg = 2.0;
  static const double xl = 3.0;
}

// --- 7. Box Shadows ---
// This class defines box shadow configurations for light and dark themes.
class AppBoxShadows {
  AppBoxShadows._(); // Private constructor to prevent instantiation.

  // Light theme shadows.
  static List<BoxShadow> lightNone = []; // An empty list represents no shadow.
  static List<BoxShadow> lightSm = const [
    BoxShadow(offset: Offset(0, 1), blurRadius: 2, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.1)),
  ];
  static List<BoxShadow> lightMd = const [
    BoxShadow(offset: Offset(0, 2), blurRadius: 4, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.16)),
  ];
  static List<BoxShadow> lightLg = const [
    BoxShadow(offset: Offset(0, 4), blurRadius: 8, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.16)),
  ];
  static List<BoxShadow> lightLgReverse = const [
    BoxShadow(offset: Offset(0, -4), blurRadius: 8, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.15)),
  ];
  static List<BoxShadow> lightModal = const [
    BoxShadow(offset: Offset(0, 4), blurRadius: 8, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.16)),
  ];
  static List<BoxShadow> lightPopover = const [
    BoxShadow(offset: Offset(0, 2), blurRadius: 4, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.16)),
  ];
  static List<BoxShadow> lightBottomSheet = const [
    BoxShadow(offset: Offset(0, -4), blurRadius: 8, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.16)),
  ];
  static List<BoxShadow> lightAccentMd = const [
    BoxShadow(offset: Offset(0, 0), blurRadius: 8, spreadRadius: 0, color: Color.fromRGBO(62, 95, 255, 0)),
  ];
  static List<BoxShadow> lightBevel = const [
    BoxShadow(offset: Offset(0, 2), blurRadius: 8, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.04)),
  ];
  static List<BoxShadow> lightAppIconFeature = const [
    BoxShadow(offset: Offset(0, 2), blurRadius: 3, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.2)),
  ];
  static List<BoxShadow> lightFocus = const [
    BoxShadow(offset: Offset(0, 0), blurRadius: 0, spreadRadius: 2, color: Color.fromRGBO(255, 255, 255, 1)),
    BoxShadow(offset: Offset(0, 0), blurRadius: 0, spreadRadius: 4, color: Color.fromRGBO(62, 95, 255, 1)),
  ];
  static List<BoxShadow> lightFocusInset = const [
    BoxShadow(offset: Offset(0, 0), blurRadius: 0, spreadRadius: 2, color: Color.fromRGBO(62, 95, 255, 1)),
    BoxShadow(offset: Offset(0, 0), blurRadius: 0, spreadRadius: 4, color: Color.fromRGBO(255, 255, 255, 1)),
  ];

  // Dark theme shadows.
  static List<BoxShadow> darkSm = const [
    BoxShadow(offset: Offset(0, 1), blurRadius: 2, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.7)),
  ];
  static List<BoxShadow> darkMd = const [
    BoxShadow(offset: Offset(0, 2), blurRadius: 4, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.7)),
  ];
  static List<BoxShadow> darkLg = const [
    BoxShadow(offset: Offset(0, 4), blurRadius: 8, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.7)),
  ];
  static List<BoxShadow> darkLgReverse = const [
    BoxShadow(offset: Offset(0, -4), blurRadius: 8, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.7)),
  ];
  static List<BoxShadow> darkModal = const [
    BoxShadow(offset: Offset(0, 4), blurRadius: 8, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.7)),
  ];
  static List<BoxShadow> darkPopover = const [
    BoxShadow(offset: Offset(0, 2), blurRadius: 4, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.7)),
  ];
  static List<BoxShadow> darkBottomSheet = const [
    BoxShadow(offset: Offset(0, -4), blurRadius: 8, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.7)),
  ];
  static List<BoxShadow> darkAccentMd = const [
    BoxShadow(offset: Offset(0, 0), blurRadius: 8, spreadRadius: 0, color: Color.fromRGBO(62, 95, 255, 0.4)),
  ];
  static List<BoxShadow> darkBevel = const [
    BoxShadow(offset: Offset(0, 0.5), blurRadius: 0, spreadRadius: 0, color: Color.fromRGBO(255, 255, 255, 0.16)), // Note: Inset shadow - implementation handled by Container decoration
  ];
  static List<BoxShadow> darkAppIconFeature = const [
    BoxShadow(offset: Offset(0, 2), blurRadius: 3, spreadRadius: 0, color: Color.fromRGBO(0, 0, 0, 0.2)),
  ];
  static List<BoxShadow> darkFocus = const [
    BoxShadow(offset: Offset(0, 0), blurRadius: 0, spreadRadius: 2, color: Color.fromRGBO(255, 255, 255, 1)),
    BoxShadow(offset: Offset(0, 0), blurRadius: 0, spreadRadius: 4, color: Color.fromRGBO(62, 95, 255, 1)),
  ];
  static List<BoxShadow> darkFocusInset = const [
    BoxShadow(offset: Offset(0, 0), blurRadius: 0, spreadRadius: 2, color: Color.fromRGBO(62, 95, 255, 1)),
    BoxShadow(offset: Offset(0, 0), blurRadius: 0, spreadRadius: 4, color: Color.fromRGBO(255, 255, 255, 1)),
  ];
}

// --- 8. Opacity ---
// This class defines common opacity values (0.0 to 1.0).
class AppOpacity {
  AppOpacity._(); // Private constructor to prevent instantiation.

  static const double o0 = 0.0;
  static const double o25 = 0.25;
  static const double o50 = 0.5;
  static const double o100 = 1.0;
}

// --- 8.5. Blur ---
// This class defines blur radius values for various blur effects.
class AppBlur {
  AppBlur._(); // Private constructor to prevent instantiation.

  static const double sm = 8;
  static const double md = 16;
  static const double lg = 24;
}

// --- 9. Transitions ---
// This class defines transition durations and timing functions (curves) for animations.
class AppTransitions {
  AppTransitions._(); // Private constructor to prevent instantiation.

  // Durations for animations.
  static const Duration durationDefault = Duration(milliseconds: 250);
  static const Duration durationSlow = Duration(milliseconds: 400);
  static const Duration durationMedium = Duration(milliseconds: 250);
  static const Duration durationFast = Duration(milliseconds: 150);

  // Timing Functions (Curves) - these are approximations for the cubic-bezier values
  // provided in your design tokens, using Flutter's built-in `Curves`.
  static const Curve timingFunctionDefault = Curves.easeInOut; // Corresponds to cubic-bezier(0.4, 0, 0.2, 1)
  static const Curve timingFunctionIn = Curves.easeIn;       // Corresponds to cubic-bezier(0.4, 0, 1, 1)
  static const Curve timingFunctionOut = Curves.easeOut;     // Corresponds to cubic-bezier(0, 0, 0.2, 1)
  static const Curve timingFunctionInOut = Curves.easeInOut; // Corresponds to cubic-bezier(0.4, 0, 0.2, 1)
}


// --- Main AppTheme Class ---
// This class aggregates all design token categories into a single, central theme
// file, making it easy to access all design constants from one place.
class AppDesign {
  AppDesign._(); // Private constructor to prevent instantiation.

  // Access to core color palette.
  static final AppCoreColors colors = AppCoreColors._();

  // Access to semantic colors for light theme.
  static final AppSemanticColorsLight lightColors = AppSemanticColorsLight._();
  // Access to semantic colors for dark theme.
  static final AppSemanticColorsDark darkColors = AppSemanticColorsDark._();

  // Access to typography definitions (font sizes, weights, styles).
  static final AppTypography typography = AppTypography._();

  // Access to spacing values.
  static final AppSpacing spacing = AppSpacing._();

  // Access to border radius values.
  static final AppBorderRadius borderRadius = AppBorderRadius._();

  // Access to border width values.
  static final AppBorderWidth borderWidth = AppBorderWidth._();

  // Access to box shadow configurations.
  static final AppBoxShadows boxShadows = AppBoxShadows._();

  // Access to opacity values.
  static final AppOpacity opacity = AppOpacity._();

  // Access to transition parameters.
  static final AppTransitions transitions = AppTransitions._();

  // Helper method to retrieve semantic colors based on the current brightness (theme mode).
  static dynamic getSemanticColors(Brightness brightness) {
    return brightness == Brightness.light ? lightColors : darkColors;
  }
}
