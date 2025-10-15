import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

// Base class providing "template" for popups.
abstract class Popup extends ConsumerWidget {
  final PopupMetadata metadata;

  const Popup({super.key, required this.metadata});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final screenSize = MediaQuery.sizeOf(context);
    final appTheme = context.appTheme;
    final theme = Theme.of(context);

    return Dialog(
      backgroundColor: Colors.transparent,
      child: Container(
        decoration: BoxDecoration(
          color: theme.colorScheme.surface,
          borderRadius: BorderRadius.circular(appTheme.borderRadiusLarge),
        ),
        padding: EdgeInsets.all(appTheme.verticalSpaceMedium),
        width: min(dynamicScale(500), screenSize.width * 0.8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          mainAxisSize: MainAxisSize.min,
          spacing: appTheme.verticalSpaceMedium,
          children: [_titleBar(context, appTheme), buildContent(context, ref)],
        ),
      ),
    );
  }

  Widget _titleBar(BuildContext context, AppTheme appTheme) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Row(
          spacing: appTheme.horizontalSpace,
          children: [if (leadingIcon != null) leadingIcon!, _title(appTheme)],
        ),
        _closeIcon(context),
      ],
    );
  }

  Widget _title(AppTheme appTheme) {
    return Text(title, style: appTheme.bodyStrong);
  }

  Widget _closeIcon(BuildContext context) {
    return IconButton(
      padding: EdgeInsets.zero,
      icon: DynamicThemeImage("close.svg"),
      onPressed: () => Navigator.of(context).pop(),
    );
  }

  void closePopup(BuildContext context) => Navigator.of(context).pop();
  String get title => metadata.title ?? t.ui.nordVpn;
  String message(WidgetRef ref) => metadata.message(ref);

  Widget? get leadingIcon => null;
  Widget buildContent(BuildContext context, WidgetRef ref);
}
