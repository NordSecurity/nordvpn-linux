import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/popup_codes.dart';
import 'package:nordvpn/settings/settings_wrapper_widget.dart';
import 'package:nordvpn/widgets/accessible_item.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';

class RealTimeProtectionSettings extends ConsumerWidget {
  const RealTimeProtectionSettings({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return ref
        .watch(vpnSettingsControllerProvider)
        .when(
          loading: () => const LoadingIndicator(),
          error: (error, stackTrace) => CustomErrorWidget(message: "$error"),
          data: (settings) => _build(context, ref, settings),
        );
  }

  Widget _build(
    BuildContext context,
    WidgetRef ref,
    ApplicationSettings settings,
  ) {
    Future<bool> shouldChange(bool toValue) async {
      // when user tries to enable it, but Custom DNS is set, we need to disable
      // Custom DNS first - ask the user and don't allow to switch TP here (it
      // will be done in popup)
      if (toValue && settings.customDnsServers.isNotEmpty) {
        ref
            .read(popupsProvider.notifier)
            .show(PopupCodes.realTimeDisableCustomDNS);
        return false;
      }

      // allow to switch only when there is no conflict with Custom DNS
      return true;
    }

    Future<void> onChanged(bool value) async {
      await ref
          .read(vpnSettingsControllerProvider.notifier)
          .setRealTimeProtection(value);
    }

    return SettingsWrapperWidget(
      itemsCount: 1,
      itemBuilder: (context, index) {
        return AccessibleItem(
          toggled: settings.realTimeProtection,
          label:
              '${t.ui.realTimeProtection}. ${t.ui.realTimeProtectionDescription}',
          onActivate: () => unawaited(
            _activate(settings.realTimeProtection, shouldChange, onChanged),
          ),
          child: SettingsWrapperWidget.buildListItem(
            context,
            title: t.ui.realTimeProtection,
            subtitle: t.ui.realTimeProtectionDescription,
            trailing: OnOffSwitch(
              value: settings.realTimeProtection,
              shouldChange: shouldChange,
              onChanged: onChanged,
            ),
          ),
        );
      },
    );
  }

  Future<void> _activate(
    bool value,
    Future<bool> Function(bool toValue) shouldChange,
    Future<void> Function(bool) onChanged,
  ) async {
    final toValue = !value;
    if (!(await shouldChange(toValue))) return;
    await onChanged(toValue);
  }
}
