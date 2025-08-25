import 'package:flutter/material.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/loading_indicator_theme.dart';

// Custom loading indicator widget with optional messages
final class LoadingIndicator extends StatelessWidget {
  final String? message;
  final String? description;
  final double? size;

  const LoadingIndicator({
    super.key,
    this.message,
    this.description,
    this.size,
  });

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;
    final loadingIndicatorTheme = context.loadingIndicatorTheme;

    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        SizedBox(
          height: size,
          width: size,
          child: CircularProgressIndicator(
            color: loadingIndicatorTheme.color,
            strokeWidth: loadingIndicatorTheme.strokeWidth,
          ),
        ),
        if (message != null || description != null) _buildMessages(appTheme),
      ],
    );
  }

  Widget _buildMessages(AppTheme appTheme) {
    return Padding(
      padding: EdgeInsets.all(appTheme.padding),
      child: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            if (message != null) SizedBox(height: appTheme.verticalSpaceSmall),
            if (message != null) Text(message!, style: appTheme.bodyStrong),
            if (description != null)
              SizedBox(height: appTheme.verticalSpaceSmall),
            if (description != null)
              Padding(
                padding: EdgeInsets.all(appTheme.padding),
                child: SelectableText(
                  description!,
                  style: appTheme.captionStrong,
                ),
              ),
          ],
        ),
      ),
    );
  }
}
