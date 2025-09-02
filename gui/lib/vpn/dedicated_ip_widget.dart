import 'package:flutter/material.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

// Widget to be displayed for dedicated IP service
class DedicatedIpWidget extends StatelessWidget {
  final String title;
  final String subtitle;
  final String buttonTitle;
  final String icon;
  final void Function(BuildContext) onPressed;

  const DedicatedIpWidget({
    super.key,
    required this.title,
    required this.subtitle,
    required this.buttonTitle,
    required this.icon,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final appTheme = theme.extension<AppTheme>()!;

    return Padding(
      padding: const EdgeInsets.all(20),
      child: Row(
        children: [
          Flexible(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(title, style: appTheme.title),
                Padding(
                  padding: const EdgeInsets.only(top: 16, bottom: 24),
                  child: Text(subtitle, style: appTheme.body),
                ),
                ElevatedButton(
                  child: Text(buttonTitle),
                  onPressed: () => onPressed(context),
                ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.only(left: 40.0),
            child: DynamicThemeImage(icon),
          ),
        ],
      ),
    );
  }
}
