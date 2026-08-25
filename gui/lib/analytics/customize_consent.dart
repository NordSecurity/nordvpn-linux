import 'package:flutter/material.dart';
import 'package:nordvpn/analytics/consent_dialog_template.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/consent_screen_theme.dart';
import 'package:nordvpn/widgets/accessible_item.dart';
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
  late bool _allowNonEssentials = widget.allowNonEssentials;

  void _toggleAnalytics() {
    if (!_isEnabled) return;
    setState(() => _allowNonEssentials = !_allowNonEssentials);
    widget.onNonEssentialsToggle(_allowNonEssentials);
  }

  @override
  Widget build(BuildContext context) {
    final consentTheme = context.consentScreenTheme;
    final appTheme = context.appTheme;

    return ConsentDialogTemplate(
      windowIcon: EnabledWidget(
        enabled: _isEnabled,
        disabledOpacity: appTheme.disabledOpacity,
        child: IconButton(
          tooltip: t.ui.back,
          onPressed: widget.onBack,
          icon: Semantics(
            label: t.ui.back,
            child: DynamicThemeImage("back_arrow.svg"),
          ),
        ),
      ),
      windowTitle: t.ui.back,
      title: t.ui.privacyPolicy,
      content: Expanded(
        child: SingleChildScrollView(
          child: Column(
            spacing: appTheme.verticalSpaceMedium,
            children: [
              AccessibleItem(
                enabled: false,
                focusable: _isEnabled,
                toggled: true,
                excludeChildSemantics: false,
                onActivate: () {},
                child: ListTile(
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
                  trailing: ExcludeSemantics(
                    child: OnOffSwitch(
                      value: true,
                      onChanged: null,
                      shouldChange: (toValue) async => false,
                    ),
                  ),
                ),
              ),
              AccessibleItem(
                toggled: _allowNonEssentials,
                enabled: _isEnabled,
                label: '${t.ui.analytics}. ${t.ui.analyticsDescription}',
                onActivate: _toggleAnalytics,
                child: ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text(
                    t.ui.analytics,
                    style: consentTheme.listItemTitle,
                  ),
                  subtitle: Text(
                    t.ui.analyticsDescription,
                    style: consentTheme.listItemSubtitle,
                  ),
                  trailing: EnabledWidget(
                    enabled: _isEnabled,
                    disabledOpacity: appTheme.disabledOpacity,
                    child: OnOffSwitch(
                      key: ValueKey(_allowNonEssentials),
                      value: _allowNonEssentials,
                      onChanged: _isEnabled
                          ? (value) async => _toggleAnalytics()
                          : null,
                    ),
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
    FocusManager.instance.primaryFocus?.unfocus();
    setState(() {
      _isEnabled = false;
    });

    await widget.onConfirm();

    if (mounted) {
      setState(() {
        _isEnabled = true;
      });
    }
  }
}
