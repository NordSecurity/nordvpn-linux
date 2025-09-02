import 'dart:async';

import 'package:flutter/material.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/error_screen_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_button.dart';

final class FullScreenError extends StatelessWidget {
  final ErrorData errorData;

  const FullScreenError({super.key, required this.errorData});

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;
    final errorScreenTheme = context.errorScreenTheme;

    return Padding(
      padding: EdgeInsets.all(appTheme.padding),
      child: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Expanded(child: SizedBox.shrink()),
            errorData.icon ?? DynamicThemeImage("something_went_wrong.svg"),
            SizedBox(height: appTheme.verticalSpaceLarge),
            Text(errorData.title, style: errorScreenTheme.titleTextStyle),
            if (errorData.subtitle != null)
              Padding(
                padding: EdgeInsets.only(top: appTheme.verticalSpaceSmall),
                child: ScalerResponsiveBox(
                  maxWidth: 262,
                  child: Text(
                    errorData.subtitle!,
                    textAlign: TextAlign.center,
                    softWrap: true,
                    style: errorScreenTheme.descriptionTextStyle,
                  ),
                ),
              ),
            if (errorData.recommendation != null)
              Padding(
                padding: EdgeInsets.only(top: appTheme.verticalSpaceMedium),
                child: errorData.recommendation!,
              ),
            if (errorData.retryCallback != null)
              Padding(
                padding: EdgeInsets.only(top: appTheme.verticalSpaceMedium),
                child: LoadingElevatedButton(
                  onPressed: () async => await errorData.retryCallback!(),
                  child: Text(t.ui.tryAgain),
                ),
              ),
            if (errorData.footer != null)
              Padding(
                padding: EdgeInsets.only(top: appTheme.verticalSpaceMedium),
                child: errorData.footer!,
              ),
            const Expanded(child: SizedBox.shrink()),
          ],
        ),
      ),
    );
  }
}

final class ErrorData {
  Widget? icon;
  String title;
  String? subtitle;
  Widget? recommendation;
  Widget? footer;
  FutureOr<void> Function()? retryCallback;

  ErrorData({
    this.icon,
    required this.title,
    this.subtitle,
    this.recommendation,
    this.footer,
    this.retryCallback,
  });
}
