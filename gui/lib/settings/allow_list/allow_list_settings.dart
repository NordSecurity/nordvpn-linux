import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/allow_list.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/popup_codes.dart';
import 'package:nordvpn/settings/allow_list/add_to_allow_list_card.dart';
import 'package:nordvpn/settings/allow_list/allow_list_content_display.dart';
import 'package:nordvpn/settings/settings_wrapper_widget.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/advanced_list_tile.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';

class AllowListSettings extends ConsumerStatefulWidget {
  const AllowListSettings({super.key});

  @override
  ConsumerState<AllowListSettings> createState() => _AllowListSettingsState();
}

class _AllowListSettingsState extends ConsumerState<AllowListSettings> {
  @override
  void initState() {
    super.initState();
    final settings = ref.read(vpnSettingsControllerProvider).valueOrNull;
    if (settings == null || settings.allowListData.isNotEmpty) return;
    // Allow List servers are not set and user just opened this page - switch
    // Allow List setting to false so the OnOffSwitch is off.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(vpnSettingsControllerProvider.notifier).setAllowList(false);
    });
  }

  @override
  Widget build(BuildContext context) {
    return ref
        .watch(vpnSettingsControllerProvider)
        .when(
          loading: () => const LoadingIndicator(),
          error: (error, stackTrace) => CustomErrorWidget(message: "$error"),
          data: (settings) => _build(context, settings),
        );
  }

  Widget _build(BuildContext context, ApplicationSettings settings) {
    final appTheme = context.appTheme;

    // NOTE: If the toggle was changed in UI, then it stays on even when user
    // clears Allow List by using daemon
    final isAllowListEnabled =
        settings.allowList || settings.allowListData.isNotEmpty;

    return SingleChildSettingsWrapperWidget(
      showDivider: false,
      stickyHeader: Padding(
        padding: EdgeInsets.only(
          left: appTheme.borderRadiusLarge,
          right: appTheme.borderRadiusLarge,
          bottom: appTheme.borderRadiusLarge,
        ),
        child: Column(
          spacing: appTheme.verticalSpaceSmall,
          children: [
            SettingsWrapperWidget.buildListItem(
              context,
              title: t.ui.useAllowList,
              subtitle: t.ui.useAllowListDescription,
              trailingLocation: TrailingLocation.center,
              trailing: OnOffSwitch(
                value: isAllowListEnabled,
                shouldChange: (toValue) => _canChange(settings, toValue),
                onChanged: (value) => _toggleAllowList(settings, value),
              ),
            ),
            AddToAllowListCard(
              key: ValueKey(isAllowListEnabled),
              enabled: isAllowListEnabled,
              allowList: settings.allowListData,
              onSubmitted: _addToAllowList,
            ),
          ],
        ),
      ),
      child: settings.allowListData.isNotEmpty
          ? Padding(
              padding: EdgeInsets.only(
                left: appTheme.verticalSpaceMedium,
                right: appTheme.verticalSpaceMedium,
              ),
              child: AllowListContentDisplay(
                allowList: settings.allowListData,
                onDeleted: _deleteFromAllowList,
              ),
            )
          : SizedBox.shrink(),
    );
  }

  Future<bool> _canChange(ApplicationSettings settings, bool toValue) async {
    // when user tries to disable it and Allow List is not empty, show
    // popup with warning and don't allow to switch to off here (it will
    // be done in popup)
    if (!toValue && settings.allowListData.isNotEmpty) {
      ref.read(popupsProvider.notifier).show(PopupCodes.turnOffAllowList);
      return false;
    }

    // allow to switch only when Allow List is empty
    return true;
  }

  Future<void> _toggleAllowList(
    ApplicationSettings settings,
    bool value,
  ) async {
    ref.read(vpnSettingsControllerProvider.notifier).setAllowList(value);
  }

  Future<bool> _addToAllowList({PortInterval? port, Subnet? subnet}) async {
    final res = await ref
        .read(vpnSettingsControllerProvider.notifier)
        .addToAllowList(port: port, subnet: subnet);
    return res == DaemonStatusCode.success;
  }

  Future<void> _deleteFromAllowList({
    PortInterval? port,
    Subnet? subnet,
  }) async {
    await ref
        .read(vpnSettingsControllerProvider.notifier)
        .removeFromAllowList(port: port, subnet: subnet);
  }
}
