import 'package:flutter/material.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/round_container.dart';

enum DialogResult { yes, no }

final class DialogFactory {
  DialogFactory._();

  // Show a simple dialog that has a title and a content widget builder
  static Future<void> _showSimpleDialog({
    required BuildContext context,
    required String title,
    Widget? icon,
    required WidgetBuilder contentBuilder,
  }) async {
    final appTheme = context.appTheme;

    return showDialog<void>(
      context: context,
      barrierDismissible: true,
      barrierColor: appTheme.overlayBackgroundColor,
      builder: (BuildContext context) {
        final screenSize = MediaQuery.of(context).size;

        return Dialog(
          backgroundColor: Colors.transparent,
          child: RoundContainer(
            // TODO: need to check with the sizes
            width: screenSize.width * 0.6,
            maxHeight: screenSize.height * 0.8,
            radius: appTheme.borderRadiusLarge,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              mainAxisSize: MainAxisSize.min,
              children: [
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  leading: icon,
                  title: Text(title, style: appTheme.bodyStrong),
                  trailing: IconButton(
                    icon: DynamicThemeImage("close.svg"),
                    onPressed: () {
                      Navigator.of(context).pop();
                    },
                  ),
                ),
                Flexible(child: contentBuilder(context)),
              ],
            ),
          ),
        );
      },
    );
  }

  // Show a popup with title, child widget and a button
  static Future<void> showPopover({
    required BuildContext context,
    Widget? icon,
    required String title,
    required Widget child,
    String buttonTitle = "",
    bool showDivider = false,
    VoidCallback? onButtonClicked,
    bool stretchButton = false,
  }) async {
    final appTheme = context.appTheme;

    return _showSimpleDialog(
      context: context,
      icon: icon,
      title: title,
      contentBuilder: (context) {
        return Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: stretchButton
              ? CrossAxisAlignment.stretch
              : CrossAxisAlignment.center,
          children: [
            if (showDivider) const Divider(),
            Flexible(
              child: Padding(
                padding: EdgeInsets.symmetric(vertical: appTheme.padding),
                child: child,
              ),
            ),
            if (buttonTitle.isNotEmpty)
              ElevatedButton(
                child: Text(buttonTitle),
                onPressed: () {
                  if (onButtonClicked != null) onButtonClicked();
                  Navigator.of(context).pop();
                },
              ),
          ],
        );
      },
    );
  }

  // Convenient function to close the current dialog
  static void close(BuildContext context) {
    Navigator.of(context).pop();
  }
}
