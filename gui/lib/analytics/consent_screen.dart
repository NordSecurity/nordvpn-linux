import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/analytics/customize_consent.dart';
import 'package:nordvpn/analytics/main_consent_dialog.dart';
import 'package:nordvpn/data/providers/consent_status_provider.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/theme/consent_screen_theme.dart';

final _navigatorKey = GlobalKey<NavigatorState>();
const _customizePath = "/customize";

// ConsentScreen - represents the entire consent screen in which also the
// navigation to customize is made
final class ConsentScreen extends ConsumerStatefulWidget {
  const ConsentScreen({super.key});

  @override
  ConsumerState<ConsentScreen> createState() => _ConsentScreenState();
}

final class _ConsentScreenState extends ConsumerState<ConsentScreen> {
  bool _allowNonEssentials = true;

  @override
  Widget build(BuildContext context) {
    final consentTheme = context.consentScreenTheme;
    return Container(
      color: consentTheme.overlayColor,
      child: Center(
        child: ScalerResponsiveBox(
          maxWidth: consentTheme.width,
          maxHeight: consentTheme.height,
          child: Navigator(
            key: _navigatorKey,
            onGenerateRoute: (settings) {
              WidgetBuilder builder;
              switch (settings.name) {
                case _customizePath:
                  builder = (context) => CustomizeConsent(
                    onBack: () {
                      if (_navigatorKey.currentState?.canPop() == true) {
                        _navigatorKey.currentState?.pop();
                      }
                    },
                    onConfirm: _submitCustomizedLevel,
                    onNonEssentialsToggle: (allowNonEssentials) =>
                        _allowNonEssentials = allowNonEssentials,
                    allowNonEssentials: _allowNonEssentials,
                  );
                  break;

                default:
                  builder = (context) => MainConsentDialog(
                    onAccept: _submitCustomizedLevel,
                    onAcceptNonEssentials: () =>
                        _setConsentLevel(ConsentLevel.essentialOnly),
                    onCustomize: () =>
                        _navigatorKey.currentState?.pushNamed(_customizePath),
                  );
              }
              return MaterialPageRoute(builder: builder, settings: settings);
            },
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
