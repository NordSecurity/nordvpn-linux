import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/analytics/customize_consent.dart';
import 'package:nordvpn/analytics/main_consent_dialog.dart';
import 'package:nordvpn/data/providers/consent_status_provider.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/theme/consent_screen_theme.dart';

// ConsentScreen - represents the entire consent screen in which also the
// navigation to customize is made
final class ConsentScreen extends ConsumerStatefulWidget {
  const ConsentScreen({super.key});

  @override
  ConsumerState<ConsentScreen> createState() => _ConsentScreenState();
}

final class _ConsentScreenState extends ConsumerState<ConsentScreen> {
  bool _allowNonEssentials = true;
  bool _showCustomize = false;

  @override
  Widget build(BuildContext context) {
    final consentTheme = context.consentScreenTheme;
    return Container(
      color: consentTheme.overlayColor,
      child: Center(
        child: ScalerResponsiveBox(
          alignment: Alignment.center,
          maxWidth: consentTheme.width,
          maxHeight: consentTheme.height,
          child: _showCustomize
              ? CustomizeConsent(
                  onBack: () => setState(() => _showCustomize = false),
                  onConfirm: _submitCustomizedLevel,
                  onNonEssentialsToggle: (allowNonEssentials) =>
                      _allowNonEssentials = allowNonEssentials,
                  allowNonEssentials: _allowNonEssentials,
                )
              : MainConsentDialog(
                  onAccept: _submitCustomizedLevel,
                  onAcceptNonEssentials: () =>
                      _setConsentLevel(ConsentLevel.essentialOnly),
                  onCustomize: () => setState(() => _showCustomize = true),
                ),
        ),
      ),
    );
  }

  Future<void> _setConsentLevel(ConsentLevel level) async {
    await ref.read(consentStatusProvider.notifier).setLevel(level);
  }

  Future<void> _submitCustomizedLevel() async {
    final level = _allowNonEssentials
        ? ConsentLevel.acceptedAll
        : ConsentLevel.essentialOnly;
    await _setConsentLevel(level);
  }
}
