import 'dart:async';

import 'package:flutter/material.dart';
import 'package:nordvpn/analytics/consent_dialog_template.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/theme/consent_screen_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_button.dart';
import 'package:nordvpn/widgets/rich_text_markdown_links.dart';

// MainConsentDialog is the main dialog displayed on the consent screen. From
//  this the user can navigate to customize the analytics
final class MainConsentDialog extends StatefulWidget {
  final VoidCallback onCustomize;
  final Future<void> Function() onAccept;
  final Future<void> Function() onAcceptNonEssentials;

  const MainConsentDialog({
    super.key,
    required this.onCustomize,
    required this.onAccept,
    required this.onAcceptNonEssentials,
  });

  @override
  State<MainConsentDialog> createState() => _MainConsentDialogState();
}

class _MainConsentDialogState extends State<MainConsentDialog> {
  final _acceptAllBtnCtrl = LoadingButtonController();
  final _essentialsBtnCtrl = LoadingButtonController();
  final _customizeBtnCtrl = LoadingButtonController();

  @override
  void dispose() {
    _acceptAllBtnCtrl.dispose();
    _essentialsBtnCtrl.dispose();
    _customizeBtnCtrl.dispose();

    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final consentTheme = context.consentScreenTheme;

    return ConsentDialogTemplate(
      windowIcon: IconButton(
        onPressed: null,
        icon: DynamicThemeImage("nordvpn_logo.svg"),
      ),
      windowTitle: "NordVPN",
      title: t.ui.weValueYourPrivacy,
      content: Expanded(
        child: SingleChildScrollView(
          child: RichTextMarkdownLinks(
            text: t.ui.consentDescription(privacyUrl: privacyPolicyUrl),
            style: consentTheme.bodyTextStyle,
          ),
        ),
      ),
      buttons: [
        Expanded(
          flex: 8,
          child: LoadingOutlinedButton(
            controller: _customizeBtnCtrl,
            onPressed: widget.onCustomize,
            child: Text(t.ui.customize),
          ),
        ),
        Expanded(
          flex: 13, // fix to give more space for this button to fit in one line
          child: LoadingOutlinedButton(
            controller: _essentialsBtnCtrl,
            onPressed: () async => await _submit(widget.onAcceptNonEssentials, [
              _acceptAllBtnCtrl,
              _customizeBtnCtrl,
            ]),
            displayModeOnLoading: DisplayModeOnLoading.both,
            child: Text(t.ui.rejectNonEssential),
          ),
        ),
        Expanded(
          flex: 8,
          child: LoadingElevatedButton(
            controller: _acceptAllBtnCtrl,
            displayModeOnLoading: DisplayModeOnLoading.both,
            onPressed: () async => await _submit(widget.onAccept, [
              _essentialsBtnCtrl,
              _customizeBtnCtrl,
            ]),
            child: Text(t.ui.accept),
          ),
        ),
      ],
    );
  }

  Future<void> _submit(
    Future<void> Function() callback,
    List<LoadingButtonController> disableButtons,
  ) async {
    for (final ctrl in disableButtons) {
      ctrl.disable();
    }
    await callback();

    for (final ctrl in disableButtons) {
      ctrl.enable();
    }
  }
}
