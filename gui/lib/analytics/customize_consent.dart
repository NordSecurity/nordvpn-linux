import 'package:flutter/material.dart';
import 'package:nordvpn/analytics/consent_dialog_template.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/consent_screen_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/enabled_widget.dart';
import 'package:nordvpn/widgets/loading_button.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';
import 'package:nordvpn/widgets/rich_text_markdown_links.dart';

// CustomizeConsent is the dialog displayed into the consent screen at the app
// startup in which the user can disable non essential analytics
final class CustomizeConsent extends StatefulWidget {
  final VoidCallback onBack;
  final Future<void> Function() onConfirm;
  final bool allowNonEssentials;
  final void Function(bool) onNonEssentialsToggle;

  const CustomizeConsent({
    super.key,
    required this.onBack,
    required this.onConfirm,
    required this.allowNonEssentials,
    required this.onNonEssentialsToggle,
  });

  @override
  State<CustomizeConsent> createState() => _CustomizeConsentState();
}

class _CustomizeConsentState extends State<CustomizeConsent> {
  bool _isEnabled = true;

  @override
  Widget build(BuildContext context) {
    final consentTheme = context.consentScreenTheme;
    final appTheme = context.appTheme;

    return ConsentDialogTemplate(
      windowIcon: EnabledWidget(
        enabled: _isEnabled,
        disabledOpacity: appTheme.disabledOpacity,
        child: IconButton(
          onPressed: widget.onBack,
          icon: DynamicThemeImage("back_arrow.svg"),
        ),
      ),
      windowTitle: t.ui.back,
      title: t.ui.privacyPolicy,
      content: Expanded(
        child: SingleChildScrollView(
          child: Column(
            spacing: appTheme.verticalSpaceMedium,
            children: [
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text(
                  t.ui.essentialRequired,
                  style: consentTheme.listItemTitle,
                ),
                subtitle: RichTextMarkdownLinks(
                  text: t.ui.requiredAnalyticsDescription(
                    termsUrl: termsOfServiceUrl,
                  ),
                  style: consentTheme.listItemSubtitle,
                ),
                trailing: OnOffSwitch(
                  value: true,
                  onChanged: null,
                  shouldChange: (toValue) async => false,
                ),
              ),
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text(t.ui.analytics, style: consentTheme.listItemTitle),
                subtitle: Text(
                  t.ui.analyticsDescription,
                  style: consentTheme.listItemSubtitle,
                ),
                trailing: EnabledWidget(
                  enabled: _isEnabled,
                  disabledOpacity: appTheme.disabledOpacity,
                  child: OnOffSwitch(
                    value: widget.allowNonEssentials,
                    onChanged: _isEnabled
                        ? (value) async => widget.onNonEssentialsToggle(value)
                        : null,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
      buttons: [
        LoadingElevatedButton(
          onPressed: () async => await _submit(),
          displayModeOnLoading: DisplayModeOnLoading.both,
          child: Text(t.ui.confirmPreferences),
        ),
      ],
    );
  }

  Future<void> _submit() async {
    setState(() {
      _isEnabled = false;
    });

    await widget.onConfirm();

    setState(() {
      _isEnabled = true;
    });
  }
}
