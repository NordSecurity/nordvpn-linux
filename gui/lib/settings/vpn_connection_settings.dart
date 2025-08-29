import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/settings/autoconnect_settings.dart';
import 'package:nordvpn/settings/settings_wrapper_widget.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/autoconnect_panel_theme.dart';
import 'package:nordvpn/widgets/advanced_list_tile.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';
import 'package:nordvpn/widgets/radio_button.dart';

enum _VpnConnectionItems { autoConnect, killSwitch, protocol }

final class VpnConnectionSettings extends ConsumerWidget {
  final ImagesManager imagesManager;

  VpnConnectionSettings({super.key, ImagesManager? imagesManager})
    : imagesManager = imagesManager ?? sl();

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
    return SettingsWrapperWidget(
      itemsCount: _VpnConnectionItems.values.length,
      itemBuilder: (context, index) {
        switch (_VpnConnectionItems.values[index]) {
          case _VpnConnectionItems.autoConnect:
            return _buildAutoConnect(context, ref, settings);
          case _VpnConnectionItems.killSwitch:
            return SettingsWrapperWidget.buildListItem(
              context,
              title: t.ui.killSwitch,
              subtitle: t.ui.killSwitchDescription,
              trailing: OnOffSwitch(
                value: settings.killSwitch,
                onChanged: (value) => ref
                    .read(vpnSettingsControllerProvider.notifier)
                    .setKillSwitch(value),
              ),
            );
          case _VpnConnectionItems.protocol:
            return _buildProtocolsList(context, ref, settings);
        }
      },
    );
  }

  Widget _buildAutoConnect(
    BuildContext context,
    WidgetRef ref,
    ApplicationSettings settings,
  ) {
    final panelTheme = context.autoconnectPanelTheme;
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        SettingsWrapperWidget.buildListItem(
          context,
          title: t.ui.autoConnect,
          subtitle: t.ui.autoConnectDescription,
          trailing: OnOffSwitch(
            value: settings.autoConnect,
            onChanged: (value) => ref
                .read(vpnSettingsControllerProvider.notifier)
                .setAutoConnect(value, null),
          ),
        ),
        SettingsWrapperWidget.buildListItem(
          context,
          title: "${t.ui.autoConnectTo}:",
          titleStyle: panelTheme.primaryFont,
          center: _buildCenter(context, ref, settings),
          trailingLocation: TrailingLocation.center,
          trailing: DynamicThemeImage("right_arrow.svg"),
          onTap: () => context.navigateToRoute(AppRoute.settingsAutoconnect),
          enabled: settings.autoConnect,
        ),
      ],
    );
  }

  Widget _buildCenter(
    BuildContext context,
    WidgetRef ref,
    ApplicationSettings settings,
  ) {
    final vpnStatusProvider = ref.watch(vpnStatusControllerProvider);
    final appTheme = context.appTheme;

    return vpnStatusProvider.when(
      error: (error, _) => CustomErrorWidget(message: "$error"),
      loading: () => LoadingIndicator(),
      data: (vpnStatus) {
        return Expanded(
          child: Padding(
            padding: EdgeInsets.only(left: appTheme.outerPadding),
            child: AutoconnectSelectionStatus(
              vpnStatus: vpnStatus,
              savedLocation: settings.autoConnectLocation,
            ),
          ),
        );
      },
    );
  }

  Widget _buildProtocolsList(
    BuildContext context,
    WidgetRef ref,
    ApplicationSettings settings,
  ) {
    final appTheme = context.appTheme;
    List<({String label, VpnProtocol value})> protocols = [
      (label: t.ui.nordLynx, value: VpnProtocol.nordlynx),
      (label: t.ui.nordWhisper, value: VpnProtocol.nordWhisper),
      (label: t.ui.openVpnTcp, value: VpnProtocol.openVpnTcp),
      (label: t.ui.openVpnUdp, value: VpnProtocol.openVpnUdp),
    ];

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        SettingsWrapperWidget.buildListItem(context, title: t.ui.vpnProtocol),
        Padding(
          padding: EdgeInsets.symmetric(horizontal: appTheme.outerPadding),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              for (final item in protocols)
                RadioButton(
                  value: item.value,
                  groupValue: settings.protocol,
                  onChanged: (value) => _setProtocol(ref, value),
                  label: item.label,
                ),
            ],
          ),
        ),
      ],
    );
  }

  void _setProtocol(WidgetRef ref, VpnProtocol value) async {
    await ref
        .read(vpnSettingsControllerProvider.notifier)
        .setVpnProtocol(value);
  }
}
