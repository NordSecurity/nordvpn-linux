import 'package:flutter/material.dart';

// --- 1. Core Colors ---
// This class defines the base color palette using hexadecimal values.
final class AppCoreColors {
  final Color transparent = Color(
    0x00000000,
  ); // Represents a fully transparent color.

  // Neutral shades from 0 (white) to 1000 (darkest black).
  final Color neutral0 = Color(0xFFFFFFFF);
  final Color neutral100 = Color(0xFFF7F7F8);
  final Color neutral150 = Color(0xFFEDEDED);
  final Color neutral200 = Color(0xFFE2E2E4);
  final Color neutral300 = Color(0xFFC8C9CB);
  final Color neutral400 = Color(0xFFB2B2B3);
  final Color neutral500 = Color(0xFF909192);
  final Color neutral600 = Color(0xFF696A6D);
  final Color neutral700 = Color(0xFF4F5054);
  final Color neutral800 = Color(0xFF3E3F42);
  final Color neutral900 = Color(0xFF2A2A2D);
  final Color neutral950 = Color(0xFF1D1E20);
  final Color neutral1000 = Color(0xFF141415);

  // Blue shades.
  final Color blue100 = Color(0xFFF3F7FC);
  final Color blue200 = Color(0xFFD4E2F7);
  final Color blue300 = Color(0xFFB5CDF5);
  final Color blue400 = Color(0xFF8CAEF8);
  final Color blue500 = Color(0xFF6B90FA);
  final Color blue600 = Color(0xFF3E5FFF);
  final Color blue700 = Color(0xFF243DCC);
  final Color blue800 = Color(0xFF263482);
  final Color blue900 = Color(0xFF22294F);
  final Color blue950 = Color(0xFF1A1F3D);
  final Color blue1000 = Color(0xFF12162B);

  // Green shades.
  final Color green100 = Color(0xFFECF9EE);
  final Color green200 = Color(0xFFB7F2C5);
  final Color green300 = Color(0xFF81E4A2);
  final Color green400 = Color(0xFF37C871);
  final Color green500 = Color(0xFF0EA464);
  final Color green600 = Color(0xFF0A8550);
  final Color green700 = Color(0xFF075F3C);
  final Color green800 = Color(0xFF05472B);
  final Color green900 = Color(0xFF043420);
  final Color green950 = Color(0xFF032617);
  final Color green1000 = Color(0xFF02180E);

  // Yellow shades.
  final Color yellow100 = Color(0xFFFFF6DB);
  final Color yellow200 = Color(0xFFFEE071);
  final Color yellow300 = Color(0xFFFAC900);
  final Color yellow400 = Color(0xFFD1A900);
  final Color yellow500 = Color(0xFFAE8604);
  final Color yellow600 = Color(0xFF8E6C10);
  final Color yellow700 = Color(0xFF654A0B);
  final Color yellow800 = Color(0xFF4E3709);
  final Color yellow900 = Color(0xFF3C2A07);
  final Color yellow950 = Color(0xFF2D2006);
  final Color yellow1000 = Color(0xFF1B1509);

  // Red shades.
  final Color red100 = Color(0xFFFCEFEE);
  final Color red200 = Color(0xFFF9D7D3);
  final Color red300 = Color(0xFFF6BEB9);
  final Color red400 = Color(0xFFF29086);
  final Color red500 = Color(0xFFEC6255);
  final Color red600 = Color(0xFFE02F1F);
  final Color red700 = Color(0xFF9E1C10);
  final Color red800 = Color(0xFF771209);
  final Color red900 = Color(0xFF5A0E07);
  final Color red950 = Color(0xFF450C07);
  final Color red1000 = Color(0xFF2F0704);
}

abstract class SemanticColors {
  // Background colors for the light theme.
  Color get bgPrimary;
  Color get bgSecondary;
  Color get bgTertiary;
  Color get bgAccent;
  Color get bgAccentSubtle;
  Color get bgDisabled;
  Color get bgSuccess;
  Color get bgCritical;
  Color get bgWarning;
  Color get bgSuccessSubtle;
  Color get bgWarningSubtle;
  Color get bgCriticalSubtle;
  Color get bgPrimaryActive;
  Color get bgSecondaryActive;
  Color get bgAccentActive;
  Color get bgOverlay;
  Color get bgInverse;
  Color get bgInverseActive;

  // Border colors for the light theme.
  Color get borderPrimary;
  Color get borderSecondary;
  Color get borderInput;
  Color get borderAccent;
  Color get borderSuccess;
  Color get borderWarning;
  Color get borderCritical;
  Color get borderAccentActive;

  // Text colors for the light theme.
  Color get textPrimary;
  Color get textSecondary;
  Color get textAccent;
  Color get textPrimaryOnColor;
  Color get textSecondaryOnColor;
  Color get textDisabled;
  Color get textSuccess;
  Color get textWarning;
  Color get textCritical;
  Color get textAccentActive;
}

// --- 2. Semantic Colors ---
// These classes define colors used for specific UI purposes, organized by theme mode.
// They reference the core colors defined in AppCoreColors for consistency.
final class AppSemanticColorsLight implements SemanticColors {
  final AppCoreColors appCoreColors = AppCoreColors();

  // Background colors for the light theme.
  @override
  Color get bgPrimary => appCoreColors.neutral100;

  @override
  Color get bgSecondary => appCoreColors.neutral0;

  @override
  Color get bgTertiary => appCoreColors.neutral150;

  @override
  Color get bgAccent => appCoreColors.blue600;

  @override
  Color get bgAccentSubtle => appCoreColors.blue100;

  @override
  Color get bgDisabled => appCoreColors.neutral300;

  @override
  Color get bgSuccess => appCoreColors.green600;

  @override
  Color get bgCritical => appCoreColors.red600;

  @override
  Color get bgWarning => appCoreColors.yellow300;

  @override
  Color get bgSuccessSubtle => appCoreColors.green100;

  @override
  Color get bgWarningSubtle => appCoreColors.yellow100;

  @override
  Color get bgCriticalSubtle => appCoreColors.red100;

  @override
  Color get bgPrimaryActive => appCoreColors.neutral150;

  @override
  Color get bgSecondaryActive => appCoreColors.neutral100;

  @override
  Color get bgAccentActive => appCoreColors.blue700;

  @override
  Color get bgOverlay => Color(0x80141415);

  @override
  Color get bgInverse => appCoreColors.neutral950;

  @override
  Color get bgInverseActive => appCoreColors.neutral800;

  // Border colors for the light theme.
  @override
  Color get borderPrimary => appCoreColors.neutral300;

  @override
  Color get borderSecondary => appCoreColors.neutral200;

  @override
  Color get borderInput => appCoreColors.neutral500;

  @override
  Color get borderAccent => appCoreColors.blue600;

  @override
  Color get borderSuccess => appCoreColors.green400;

  @override
  Color get borderWarning => appCoreColors.yellow300;

  @override
  Color get borderCritical => appCoreColors.red400;

  @override
  Color get borderAccentActive => appCoreColors.blue700;

  // Text colors for the light theme.
  @override
  Color get textPrimary => appCoreColors.neutral900;

  @override
  Color get textSecondary => appCoreColors.neutral600;

  @override
  Color get textAccent => appCoreColors.blue600;

  @override
  Color get textPrimaryOnColor => appCoreColors.neutral0;

  @override
  Color get textSecondaryOnColor => appCoreColors.neutral100;

  @override
  Color get textDisabled => appCoreColors.neutral400;

  @override
  Color get textSuccess => appCoreColors.green700;

  @override
  Color get textWarning => appCoreColors.yellow700;

  @override
  Color get textCritical => appCoreColors.red700;

  @override
  Color get textAccentActive => appCoreColors.blue700;
}

final class AppSemanticColorsDark implements SemanticColors {
  final AppCoreColors appCoreColors = AppCoreColors();

  // Background colors for the dark theme.
  @override
  Color get bgPrimary => appCoreColors.neutral1000;

  @override
  Color get bgSecondary => appCoreColors.neutral950;

  @override
  Color get bgTertiary => appCoreColors.neutral900;

  @override
  Color get bgAccent => appCoreColors.blue600;

  @override
  Color get bgAccentSubtle => appCoreColors.blue950;

  @override
  Color get bgDisabled => appCoreColors.neutral800;

  @override
  Color get bgSuccess => appCoreColors.green600;

  @override
  Color get bgCritical => appCoreColors.red700;

  @override
  Color get bgWarning => appCoreColors.yellow600;

  @override
  Color get bgSuccessSubtle => appCoreColors.green900;

  @override
  Color get bgWarningSubtle => appCoreColors.yellow900;

  @override
  Color get bgCriticalSubtle => appCoreColors.red900;

  @override
  Color get bgPrimaryActive => appCoreColors.neutral950;

  @override
  Color get bgSecondaryActive => appCoreColors.neutral900;

  @override
  Color get bgAccentActive => appCoreColors.blue700;

  @override
  Color get bgOverlay => Color(0x80141415);

  @override
  Color get bgInverse => appCoreColors.neutral0;

  @override
  Color get bgInverseActive => appCoreColors.neutral200;

  // Border colors for the dark theme.

  @override
  Color get borderPrimary => appCoreColors.neutral700;

  @override
  Color get borderSecondary => appCoreColors.neutral800;

  @override
  Color get borderInput => appCoreColors.neutral600;

  @override
  Color get borderAccent => appCoreColors.blue600;

  @override
  Color get borderSuccess => appCoreColors.green600;

  @override
  Color get borderWarning => appCoreColors.yellow600;

  @override
  Color get borderCritical => appCoreColors.red600;

  @override
  Color get borderAccentActive => appCoreColors.blue700;

  // Text colors for the dark theme.

  @override
  Color get textPrimary => appCoreColors.neutral0;

  @override
  Color get textSecondary => appCoreColors.neutral500;

  @override
  Color get textAccent => appCoreColors.blue500;

  @override
  Color get textPrimaryOnColor => appCoreColors.neutral0;

  @override
  Color get textSecondaryOnColor => appCoreColors.neutral100;

  @override
  Color get textDisabled => appCoreColors.neutral700;

  @override
  Color get textSuccess => appCoreColors.green400;

  @override
  Color get textWarning => appCoreColors.yellow300;

  @override
  Color get textCritical => appCoreColors.red500;

  @override
  Color get textAccentActive => appCoreColors.blue400;
}

// --- 3. Typography ---
// This class defines typography constants, including font sizes, weights,
// letter spacing, line heights, and pre-defined text styles for various elements.
final class AppTypography {
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
  static const double fontSizeXs = 0.75 * 16; // 12px
  static const double fontSizeSm = 0.875 * 16; // 14px
  static const double fontSizeMd = 1.0 * 16; // 16px
  static const double fontSizeLg = 1.125 * 16; // 18px
  static const double fontSizeXl = 1.25 * 16; // 20px
  static const double fontSize2xl = 1.375 * 16; // 22px
  static const double fontSize3xl = 1.5 * 16; // 24px
  static const double fontSize4xl = 1.625 * 16; // 26px
  static const double fontSize5xl = 1.75 * 16; // 28px
  static const double fontSize6xl = 2.0 * 16; // 32px
  static const double fontSize7xl = 2.5 * 16; // 40px
  static const double fontSize8xl = 3.0 * 16; // 48px
  static const double fontSize9xl = 3.5 * 16; // 56px

  // Font Weights corresponding to the design tokens.
  static const FontWeight fontWeightNormal = FontWeight.w400;
  static const FontWeight fontWeightMedium = FontWeight.w500;
  static const FontWeight fontWeightBold =
      FontWeight.w600; // Typically Flutter's w600 is "semi-bold"

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
  final TextStyle display = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize5xl,
    letterSpacing: letterSpacingSm,
    height: 36 / (1.75 * 16), // 36px line-height
  );

  final TextStyle heading = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize3xl,
    letterSpacing: letterSpacingMd,
    height: 32 / (1.5 * 16), // 32px line-height
  );

  final TextStyle subHeading = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightMedium,
    fontSize: fontSizeMd,
    letterSpacing: letterSpacingMd,
    height: 22 / (1.0 * 16), // 22px line-height
  );

  final TextStyle body = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightNormal,
    fontSize: fontSizeMd,
    letterSpacing: letterSpacingMd,
    height: 22 / (1.0 * 16), // 22px line-height
  );

  final TextStyle subBody = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightNormal,
    fontSize: fontSizeSm,
    letterSpacing: letterSpacingMd,
    height: 20 / (0.875 * 16), // 20px line-height
  );

  final TextStyle caption = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightNormal,
    fontSize: fontSizeXs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.75 * 16), // 18px line-height
  );

  final TextStyle captionMedium = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightMedium,
    fontSize: fontSizeXs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.75 * 16), // 18px line-height
  );

  // Additional heading styles.
  final TextStyle heading2xl = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize7xl,
    letterSpacing: letterSpacingSm,
    height: 48 / (2.5 * 16),
  );

  final TextStyle headingXl = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize6xl,
    letterSpacing: letterSpacingSm,
    height: 40 / (2.0 * 16),
  );

  final TextStyle headingLg = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize5xl,
    letterSpacing: letterSpacingSm,
    height: 36 / (1.75 * 16),
  );

  final TextStyle headingMd = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize4xl,
    letterSpacing: letterSpacingMd,
    height: 34 / (1.625 * 16),
  );

  final TextStyle headingSm = TextStyle(
    fontFamily: fontFamilyHeading,
    fontWeight: fontWeightBold,
    fontSize: fontSize2xl,
    letterSpacing: letterSpacingMd,
    height: 30 / (1.375 * 16),
  );

  // Additional body styles.
  final TextStyle bodyLg = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightNormal,
    fontSize: fontSizeLg,
    letterSpacing: letterSpacingMd,
    height: 24 / (1.125 * 16),
  );

  final TextStyle bodyLgMedium = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightMedium,
    fontSize: fontSizeLg,
    letterSpacing: letterSpacingMd,
    height: 24 / (1.125 * 16),
  );

  final TextStyle bodyLgBold = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightBold,
    fontSize: fontSizeLg,
    letterSpacing: letterSpacingMd,
    height: 24 / (1.125 * 16),
  );

  final TextStyle bodyMd = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightNormal,
    fontSize: fontSizeMd,
    letterSpacing: letterSpacingMd,
    height: 22 / (1.0 * 16),
  );

  final TextStyle bodyMdMedium = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightMedium,
    fontSize: fontSizeMd,
    letterSpacing: letterSpacingMd,
    height: 22 / (1.0 * 16),
  );

  final TextStyle bodyMdBold = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: FontWeight.bold,
    fontSize: fontSizeMd,
    letterSpacing: letterSpacingMd,
    height: 22 / (1.0 * 16),
  );

  final TextStyle bodySm = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: fontWeightNormal,
    fontSize: fontSizeSm,
    letterSpacing: letterSpacingMd,
    height: 20 / (0.875 * 16),
  );

  final TextStyle bodySmMedium = TextStyle(
    fontFamily: fontFamilyBody,
    // fontWeight: FontWeight.medium,
    // TODO: change it
    fontWeight: FontWeight.normal,
    fontSize: fontSizeSm,
    letterSpacing: letterSpacingMd,
    height: 20 / (0.875 * 16),
  );

  final TextStyle bodySmBold = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: FontWeight.bold,
    fontSize: fontSizeSm,
    letterSpacing: letterSpacingMd,
    height: 20 / (0.875 * 16),
  );

  final TextStyle bodyXs = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: FontWeight.normal,
    fontSize: fontSizeXs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.75 * 16),
  );

  final TextStyle bodyXsMedium = TextStyle(
    fontFamily: fontFamilyBody,
    // fontWeight: FontWeight.medium,
    // TODO: change it
    fontWeight: FontWeight.normal,
    fontSize: fontSizeXs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.75 * 16),
  );

  final TextStyle bodyXsBold = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: FontWeight.bold,
    fontSize: fontSizeXs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.75 * 16),
  );

  final TextStyle body2xs = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: FontWeight.normal,
    fontSize: fontSize2xs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.6875 * 16),
  );

  final TextStyle body2xsMedium = TextStyle(
    fontFamily: fontFamilyBody,
    // fontWeight: FontWeight.medium,
    // TODO: change it
    fontWeight: FontWeight.normal,
    fontSize: fontSize2xs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.6875 * 16),
  );

  final TextStyle body2xsBold = TextStyle(
    fontFamily: fontFamilyBody,
    fontWeight: FontWeight.bold,
    fontSize: fontSize2xs,
    letterSpacing: letterSpacingMd,
    height: 18 / (0.6875 * 16),
  );
}

// --- 4. Spacing ---
// This class defines common spacing values, converted to logical pixels.
final class AppSpacing {
  const AppSpacing._(); // Private constructor to prevent instantiation.

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
final class AppBorderRadius {
  const AppBorderRadius._(); // Private constructor to prevent instantiation.

  static const BorderRadius none = BorderRadius.zero;
  static const BorderRadius xs = BorderRadius.all(Radius.circular(3.0));
  static const BorderRadius sm = BorderRadius.all(Radius.circular(6.0));
  static const BorderRadius md = BorderRadius.all(Radius.circular(12.0));
  static const BorderRadius lg = BorderRadius.all(Radius.circular(20.0));
  static const BorderRadius full = BorderRadius.all(Radius.circular(9999.0));
}

// --- 6. Border Width ---
// This class defines common border width values.
final class AppBorderWidth {
  const AppBorderWidth._(); // Private constructor to prevent instantiation.

  static const double none = 0.0;
  static const double sm = 0.5;
  static const double md = 1.0;
  static const double lg = 2.0;
  static const double xl = 3.0;
}

// --- 7. Box Shadows ---
// This class defines box shadow configurations for light and dark themes.
final class AppBoxShadows {
  const AppBoxShadows._(); // Private constructor to prevent instantiation.

  // Light theme shadows.
  static List<BoxShadow> lightNone = []; // An empty list represents no shadow.
  static List<BoxShadow> lightSm = const [
    BoxShadow(
      offset: Offset(0, 1),
      blurRadius: 2,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.1),
    ),
  ];
  static List<BoxShadow> lightMd = const [
    BoxShadow(
      offset: Offset(0, 2),
      blurRadius: 4,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.16),
    ),
  ];
  static List<BoxShadow> lightLg = const [
    BoxShadow(
      offset: Offset(0, 4),
      blurRadius: 8,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.16),
    ),
  ];
  static List<BoxShadow> lightLgReverse = const [
    BoxShadow(
      offset: Offset(0, -4),
      blurRadius: 8,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.15),
    ),
  ];
  static List<BoxShadow> lightModal = const [
    BoxShadow(
      offset: Offset(0, 4),
      blurRadius: 8,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.16),
    ),
  ];
  static List<BoxShadow> lightPopover = const [
    BoxShadow(
      offset: Offset(0, 2),
      blurRadius: 4,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.16),
    ),
  ];
  static List<BoxShadow> lightBottomSheet = const [
    BoxShadow(
      offset: Offset(0, -4),
      blurRadius: 8,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.16),
    ),
  ];
  static List<BoxShadow> lightAccentMd = const [
    BoxShadow(
      offset: Offset(0, 0),
      blurRadius: 8,
      spreadRadius: 0,
      color: Color.fromRGBO(62, 95, 255, 0),
    ),
  ];
  static List<BoxShadow> lightBevel = const [
    BoxShadow(
      offset: Offset(0, 0.5),
      blurRadius: 0,
      spreadRadius: 0,
      color: Color.fromRGBO(255, 255, 255, 0.16),
    ),
  ];

  // Dark theme shadows.
  static List<BoxShadow> darkSm = const [
    BoxShadow(
      offset: Offset(0, 1),
      blurRadius: 2,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.7),
    ),
  ];
  static List<BoxShadow> darkMd = const [
    BoxShadow(
      offset: Offset(0, 2),
      blurRadius: 4,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.7),
    ),
  ];
  static List<BoxShadow> darkLg = const [
    BoxShadow(
      offset: Offset(0, 4),
      blurRadius: 8,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.7),
    ),
  ];
  static List<BoxShadow> darkLgReverse = const [
    BoxShadow(
      offset: Offset(0, -4),
      blurRadius: 8,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.7),
    ),
  ];
  static List<BoxShadow> darkModal = const [
    BoxShadow(
      offset: Offset(0, 4),
      blurRadius: 8,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.7),
    ),
  ];
  static List<BoxShadow> darkPopover = const [
    BoxShadow(
      offset: Offset(0, 2),
      blurRadius: 4,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.7),
    ),
  ];
  static List<BoxShadow> darkBottomSheet = const [
    BoxShadow(
      offset: Offset(0, -4),
      blurRadius: 8,
      spreadRadius: 0,
      color: Color.fromRGBO(0, 0, 0, 0.7),
    ),
  ];
  static List<BoxShadow> darkAccentMd = const [
    BoxShadow(
      offset: Offset(0, 0),
      blurRadius: 8,
      spreadRadius: 0,
      color: Color.fromRGBO(62, 95, 255, 0.4),
    ),
  ];
  static List<BoxShadow> darkBevel = const [
    BoxShadow(
      offset: Offset(0, 0.5),
      blurRadius: 0,
      spreadRadius: 0,
      color: Color.fromRGBO(255, 255, 255, 0.16),
    ),
  ];
}

// --- 8. Opacity ---
// This class defines common opacity values (0.0 to 1.0).
final class AppOpacity {
  const AppOpacity._(); // Private constructor to prevent instantiation.

  static const double o0 = 0.0;
  static const double o25 = 0.25;
  static const double o50 = 0.5;
  static const double o100 = 1.0;
}

// --- 9. Transitions ---
// This class defines transition durations and timing functions (curves) for animations.
final class AppTransitions {
  const AppTransitions._(); // Private constructor to prevent instantiation.

  // Durations for animations.
  static const Duration durationDefault = Duration(milliseconds: 250);
  static const Duration durationSlow = Duration(milliseconds: 400);
  static const Duration durationMedium = Duration(milliseconds: 250);
  static const Duration durationFast = Duration(milliseconds: 150);

  // Timing Functions (Curves) - these are approximations for the cubic-bezier values
  // provided in your design tokens, using Flutter's built-in `Curves`.
  static const Curve timingFunctionDefault =
      Curves.easeInOut; // Corresponds to cubic-bezier(0.4, 0, 0.2, 1)
  static const Curve timingFunctionIn =
      Curves.easeIn; // Corresponds to cubic-bezier(0.4, 0, 1, 1)
  static const Curve timingFunctionOut =
      Curves.easeOut; // Corresponds to cubic-bezier(0, 0, 0.2, 1)
  static const Curve timingFunctionInOut =
      Curves.easeInOut; // Corresponds to cubic-bezier(0.4, 0, 0.2, 1)
}

// --- Main AppTheme Class ---
// This class aggregates all design token categories into a single, central theme
// file, making it easy to access all design constants from one place.
final class AppDesign {
  AppDesign(this.mode)
    : semanticColors = mode == ThemeMode.light
          ? AppSemanticColorsLight()
          : AppSemanticColorsDark();

  final ThemeMode mode;

  // Access to semantic colors for light/dark theme.
  final SemanticColors semanticColors;

  // Access to typography definitions (font sizes, weights, styles).
  final AppTypography typography = AppTypography();

  // Access to core color palette.
  final AppCoreColors colors = AppCoreColors();

  // Access to spacing values.
  static const AppSpacing spacing = AppSpacing._();

  // Access to border radius values.
  static const AppBorderRadius borderRadius = AppBorderRadius._();

  // Access to border width values.
  static const AppBorderWidth borderWidth = AppBorderWidth._();

  // Access to box shadow configurations.
  static const AppBoxShadows boxShadows = AppBoxShadows._();

  // Access to opacity values.
  static const AppOpacity opacity = AppOpacity._();

  // Access to transition parameters.
  static const AppTransitions transitions = AppTransitions._();
}
