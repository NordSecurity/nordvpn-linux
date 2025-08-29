import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/providers/consent_status_provider.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/providers/preferences_controller.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/popup_codes.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/settings/settings_wrapper_widget.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/custom_expansion_tile.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';
import 'package:nordvpn/widgets/radio_button.dart';
import 'package:nordvpn/widgets/rich_text_markdown_links.dart';

// Settings types displayed into the screen
enum _GeneralSettingsItems {
  appearance,
  notifications,
  analytics,
  restoreToDefaults,
}

class GeneralSettings extends ConsumerWidget {
  const GeneralSettings({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final vpnSettings = ref.watch(vpnSettingsControllerProvider);
    return vpnSettings.when(
      loading: () => const LoadingIndicator(),
      error: (error, stackTrace) => CustomErrorWidget(message: "$error"),
      data: (settings) => ref
          .watch(preferencesControllerProvider)
          .maybeWhen(
            data: (preferences) =>
                _build(context, ref, settings, preferences.appearance),
            orElse: () => _build(context, ref, settings, defaultTheme),
          ),
    );
  }

  Widget _build(
    BuildContext context,
    WidgetRef ref,
    ApplicationSettings settings,
    ThemeMode mode,
  ) {
    final appTheme = context.appTheme;

    return SettingsWrapperWidget(
      itemsCount: _GeneralSettingsItems.values.length,
      itemBuilder: (context, index) {
        switch (_GeneralSettingsItems.values[index]) {
          case _GeneralSettingsItems.appearance:
            return SettingsWrapperWidget.buildListItem(
              context,
              title: t.ui.appearance,
              trailing: _buildAppearanceTrailing(context, ref, mode),
            );
          case _GeneralSettingsItems.notifications:
            return SettingsWrapperWidget.buildListItem(
              context,
              title: t.ui.showNotifications,
              trailing: OnOffSwitch(
                value: settings.notifications,
                onChanged: (value) => ref
                    .read(vpnSettingsControllerProvider.notifier)
                    .setNotifications(value),
              ),
            );
          case _GeneralSettingsItems.analytics:
            return CustomExpansionTile(
              title: Text(t.ui.privacyPreferences, style: appTheme.body),
              subtitle: Text(
                t.ui.privacyPreferencesDescription,
                style: appTheme.caption,
              ),
              contentPadding: EdgeInsets.zero,
              children: [
                SettingsWrapperWidget.buildListItem(
                  context,
                  title: t.ui.essentialRequired,
                  subtitleWidget: RichTextMarkdownLinks(
                    text: t.ui.requiredAnalyticsDescription(
                      termsUrl: termsOfServiceUrl,
                    ),
                  ),
                  trailing: OnOffSwitch(value: true, onChanged: null),
                ),
                SettingsWrapperWidget.buildListItem(
                  context,
                  title: t.ui.analytics,
                  subtitle: t.ui.analyticsDescription,
                  trailing: OnOffSwitch(
                    value:
                        settings.analyticsConsent == ConsentLevel.acceptedAll,
                    onChanged: (value) => ref
                        .read(vpnSettingsControllerProvider.notifier)
                        .setAnalytics(value),
                  ),
                ),
              ],
            );
          case _GeneralSettingsItems.restoreToDefaults:
            return SettingsWrapperWidget.buildListItem(
              context,
              title: t.ui.resetToDefaults,
              trailing: ElevatedButton(
                child: Text(t.ui.reset),
                onPressed: () => _resetToDefaults(ref),
              ),
            );
        }
      },
    );
  }
}

Widget _buildAppearanceTrailing(
  BuildContext context,
  WidgetRef ref,
  ThemeMode mode,
) {
  final appTheme = context.appTheme;

  return Row(
    mainAxisSize: MainAxisSize.min,
    spacing: appTheme.verticalSpaceSmall,
    children: [
      RadioButton(
        value: ThemeMode.system,
        groupValue: mode,
        onChanged: (value) => _setAppearance(ref, value),
        label: t.ui.system,
        labelStyle: appTheme.body,
      ),
      RadioButton(
        value: ThemeMode.light,
        groupValue: mode,
        onChanged: (value) => _setAppearance(ref, value),
        label: t.ui.light,
        labelStyle: appTheme.body,
      ),
      RadioButton(
        value: ThemeMode.dark,
        groupValue: mode,
        onChanged: (value) => _setAppearance(ref, value),
        label: t.ui.dark,
        labelStyle: appTheme.body,
      ),
    ],
  );
}

void _setAppearance(WidgetRef ref, ThemeMode value) async {
  await ref.read(preferencesControllerProvider.notifier).setAppearance(value);
}

void _resetToDefaults(WidgetRef ref) async {
  final vpnStatus = ref.read(vpnStatusControllerProvider).value;
  ref
      .read(popupsProvider.notifier)
      .show(
        (vpnStatus != null && vpnStatus.isConnected())
            ? PopupCodes.resetSettingsAndDisconnect
            : PopupCodes.resetSettings,
      );
}
