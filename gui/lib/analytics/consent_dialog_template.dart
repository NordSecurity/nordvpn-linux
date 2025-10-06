import 'package:flutter/material.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/consent_screen_theme.dart';
import 'package:nordvpn/widgets/round_container.dart';

// ConsentDialogTemplate is the template dialog used for main consent screen
// and for the customized, just updating the fields
final class ConsentDialogTemplate extends StatelessWidget {
  final Widget windowIcon;
  final String windowTitle;
  final String title;
  final Widget content;
  final List<Widget> buttons;

  const ConsentDialogTemplate({
    super.key,
    required this.windowIcon,
    required this.windowTitle,
    required this.title,
    required this.content,
    required this.buttons,
  });

  @override
  Widget build(BuildContext context) {
    final consentTheme = context.consentScreenTheme;

    return RoundContainer(
      margin: EdgeInsets.zero,
      padding: EdgeInsets.symmetric(horizontal: consentTheme.padding / 2),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildTitle(context),
          Expanded(
            child: Padding(
              padding: EdgeInsets.only(
                left: consentTheme.padding / 2,
                right: consentTheme.padding / 2,
                bottom: consentTheme.padding,
              ),
              child: Column(
                spacing: context.appTheme.verticalSpaceMedium,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(title, style: consentTheme.titleTextStyle),
                  content,
                  _buildButtons(context),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTitle(BuildContext context) {
    // add the 2px padding so the give some space for hover effect
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: SizedBox(
        height: context.consentScreenTheme.titleBarWidth,
        child: Row(
          spacing: 2,
          children: [
            windowIcon,
            Text(
              windowTitle,
              style: context.consentScreenTheme.titleBarTextStyle,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildButtons(BuildContext context) {
    return Row(
      spacing: context.appTheme.horizontalSpaceSmall,
      children: buttons,
    );
  }
}
