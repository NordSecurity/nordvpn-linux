import 'dart:async';

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
import 'package:nordvpn/theme/settings_theme.dart';
import 'package:nordvpn/widgets/accessible_item.dart';
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
            return _buildAppearance(context, ref, mode);
          case _GeneralSettingsItems.notifications:
            return _accessibleSwitchTile(
              context,
              title: t.ui.showNotifications,
              value: settings.notifications,
              onChanged: (value) => ref
                  .read(vpnSettingsControllerProvider.notifier)
                  .setNotifications(value),
            );
          case _GeneralSettingsItems.analytics:
            return CustomExpansionTile(
              title: MergeSemantics(
                child: Semantics(
                  header: true,
                  child: Text(t.ui.privacyPreferences, style: appTheme.body),
                ),
              ),
              subtitle: MergeSemantics(
                child: Text(
                  t.ui.privacyPreferencesDescription,
                  style: appTheme.caption,
                ),
              ),
              contentPadding: EdgeInsets.zero,
              semanticTitle: t.ui.privacyPreferences,
              children: [
                AccessibleItem(
                  enabled: false,
                  focusable: true,
                  toggled: true,
                  excludeChildSemantics: false,
                  onActivate: () {},
                  child: SettingsWrapperWidget.buildListItem(
                    context,
                    title: t.ui.essentialRequired,
                    subtitleWidget: RichTextMarkdownLinks(
                      text: t.ui.requiredAnalyticsDescription(
                        termsUrl: termsOfServiceUrl,
                      ),
                    ),
                    trailing: ExcludeSemantics(
                      child: OnOffSwitch(value: true, onChanged: null),
                    ),
                  ),
                ),
                _accessibleSwitchTile(
                  context,
                  title: t.ui.analytics,
                  subtitle: t.ui.analyticsDescription,
                  value: settings.analyticsConsent == ConsentLevel.acceptedAll,
                  onChanged: (value) => ref
                      .read(vpnSettingsControllerProvider.notifier)
                      .setAnalytics(value),
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

Widget _buildAppearance(BuildContext context, WidgetRef ref, ThemeMode mode) {
  final settingsTheme = context.settingsTheme;
  final appTheme = context.appTheme;

  return Padding(
    padding: settingsTheme.itemPadding,
    child: Row(
      children: [
        Expanded(
          child: ExcludeSemantics(
            child: Text(t.ui.appearance, style: settingsTheme.itemTitleStyle),
          ),
        ),
        Row(
          mainAxisSize: MainAxisSize.min,
          spacing: appTheme.verticalSpaceSmall,
          children: [
            _appearanceRadio(context, ref, ThemeMode.system, t.ui.system, mode),
            _appearanceRadio(context, ref, ThemeMode.light, t.ui.light, mode),
            _appearanceRadio(context, ref, ThemeMode.dark, t.ui.dark, mode),
          ],
        ),
      ],
    ),
  );
}

Widget _appearanceRadio(
  BuildContext context,
  WidgetRef ref,
  ThemeMode value,
  String label,
  ThemeMode groupValue,
) {
  final appTheme = context.appTheme;
  return AccessibleItem(
    inMutuallyExclusiveGroup: true,
    checked: value == groupValue,
    label: '${t.ui.appearance}, $label',
    onActivate: () => _setAppearance(ref, value),
    child: RadioButton(
      value: value,
      groupValue: groupValue,
      onChanged: (v) => _setAppearance(ref, v),
      label: label,
      labelStyle: appTheme.body,
    ),
  );
}

Widget _accessibleSwitchTile(
  BuildContext context, {
  required String title,
  String? subtitle,
  required bool value,
  required Future<void> Function(bool) onChanged,
}) {
  return AccessibleItem(
    toggled: value,
    label: subtitle == null ? title : '$title. $subtitle',
    onActivate: () => unawaited(onChanged(!value)),
    child: SettingsWrapperWidget.buildListItem(
      context,
      title: title,
      subtitle: subtitle,
      trailing: OnOffSwitch(value: value, onChanged: onChanged),
    ),
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
