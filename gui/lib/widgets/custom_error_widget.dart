import 'package:flutter/material.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/loading_button.dart';
import 'package:nordvpn/widgets/rich_text_markdown_links.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

// Custom errors widget
class CustomErrorWidget extends StatelessWidget {
  final String message;
  final String buttonText;
  final FutureOr<void> Function()? onPressed;

  const CustomErrorWidget({
    super.key,
    required this.message,
    this.buttonText = "",
    this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final appTheme = theme.extension<AppTheme>()!;

    return Padding(
      padding: EdgeInsets.all(appTheme.padding),
      child: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          spacing: appTheme.verticalSpaceMedium,
          children: [
            Text(message, style: appTheme.bodyStrong),
            if (buttonText.isNotEmpty && (onPressed != null))
              LoadingOutlinedButton(
                onPressed: onPressed,
                child: Text(buttonText),
              ),
            RichTextMarkdownLinks(
              text: t.ui.forTroubleshooting(supportUrl: supportCenterUrl),
            ),
          ],
        ),
      ),
    );
  }
}
