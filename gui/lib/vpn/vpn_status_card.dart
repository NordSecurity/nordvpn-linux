import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/server_group_extension.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/i18n/string_translation_extension.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/vpn_status_card_theme.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/padded_circle_avatar.dart';
import 'package:nordvpn/widgets/round_container.dart';

final class VpnStatusCard extends StatelessWidget {
  final ImagesManager imagesManager;

  VpnStatusCard({super.key, ImagesManager? imagesManager})
    : imagesManager = imagesManager ?? sl();

  @override
  Widget build(BuildContext context) {
    final statusCardTheme = context.vpnStatusCardTheme;
    final appTheme = context.appTheme;

    return RoundContainer(
      minHeight: statusCardTheme.height,
      margin: EdgeInsets.only(
        top: appTheme.margin,
        bottom: 0,
        right: appTheme.margin,
        left: appTheme.margin,
      ),
      child: Consumer(
        builder: (context, ref, _) {
          return AnimatedSwitcher(
            duration: const Duration(milliseconds: 200),
            child: ref
                .watch(vpnStatusControllerProvider)
                .when(
                  data: (status) => _build(ref, context, status),
                  error: (error, stackTrace) =>
                      CustomErrorWidget(message: "$error"),
                  loading: () => const LoadingIndicator(),
                ),
          );
        },
      ),
    );
  }

  Widget _build(WidgetRef ref, BuildContext context, VpnStatus vpnStatus) {
    final appTheme = context.appTheme;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      key: ValueKey(vpnStatus.status),
      children: [
        Row(
          children: [
            VpnStatusIcon(status: vpnStatus),
            SizedBox(width: appTheme.horizontalSpace),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                VpnStatusLabel(vpnStatus: vpnStatus),
                VpnServerInfo(vpnStatus: vpnStatus),
              ],
            ),
          ],
        ),
        SizedBox(height: appTheme.verticalSpaceSmall),
        ScalerResponsiveBox(
          maxWidth: context.vpnStatusCardTheme.maxConnectButtonWidth,
          child: IntrinsicHeight(
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              spacing: appTheme.horizontalSpaceSmall,
              children: _buildButtons(ref, appTheme, vpnStatus),
            ),
          ),
        ),
      ],
    );
  }

  List<Widget> _buildButtons(
    WidgetRef ref,
    AppTheme appTheme,
    VpnStatus status,
  ) {
    final settings = ref.watch(vpnSettingsControllerProvider).valueOrNull;
    if (status.isConnected()) {
      return [
        Expanded(
          child: OutlinedButton(
            onPressed: () =>
                ref.read(vpnStatusControllerProvider.notifier).disconnect(),
            child: Text(t.ui.disconnect),
          ),
        ),
        if (!status.isMeshnetRouting)
          OutlinedButton(
            style: OutlinedButton.styleFrom(padding: EdgeInsets.all(0)),
            onPressed: () => _reconnect(ref, status, settings),
            child: DynamicThemeImage("reconnect.svg"),
          ),
      ];
    }

    return [
      Expanded(
        child: ElevatedButton(
          onPressed: () async {
            if (status.isDisconnected()) {
              // Quick connect
              ConnectArguments? args;
              if (settings?.obfuscatedServers == true) {
                args = ConnectArguments(specialtyGroup: ServerType.obfuscated);
              }
              ref.read(vpnStatusControllerProvider.notifier).connect(args);
            } else if (status.isConnecting()) {
              ref.read(vpnStatusControllerProvider.notifier).cancelConnect();
            }
          },
          child: Text(status.isConnecting() ? t.ui.cancel : t.ui.quickConnect),
        ),
      ),
    ];
  }

  Future<void> _reconnect(
    WidgetRef ref,
    VpnStatus status,
    ApplicationSettings? settings,
  ) async {
    if (settings?.obfuscatedServers == true) {
      status.connectionParameters.group = ServerType.obfuscated.toServerGroup();
    }
    ref
        .read(vpnStatusControllerProvider.notifier)
        .reconnect(status.connectionParameters);
  }
}

final class VpnStatusIcon extends StatelessWidget {
  final ImagesManager imagesManager;
  final VpnStatus status;

  VpnStatusIcon({super.key, required this.status, ImagesManager? imagesManager})
    : imagesManager = imagesManager ?? sl();

  @override
  Widget build(BuildContext context) {
    final statusCardTheme = context.vpnStatusCardTheme;

    if (status.isConnected()) {
      assert(status.country != null || status.isMeshnetRouting);
      final appTheme = context.appTheme;

      return PaddedCircleAvatar(
        size: statusCardTheme.iconSize,
        borderColor: appTheme.successColor,
        borderSize: appTheme.flagsBorderSize,
        child: icon(),
      );
    }
    if (status.isConnecting()) {
      return LoadingIndicator(size: statusCardTheme.iconSize);
    }
    return DynamicThemeImage("vpn_not_connected.svg");
  }

  Widget icon() {
    if (status.isMeshnetRouting) {
      return DynamicThemeImage("linux_peer.svg");
    }
    if (status.country != null) {
      return imagesManager.forCountry(status.country!);
    }

    return imagesManager.placeholderCountryFlag;
  }
}

final class VpnStatusLabel extends ConsumerWidget {
  final VpnStatus vpnStatus;

  const VpnStatusLabel({super.key, required this.vpnStatus});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final appTheme = context.appTheme;
    final statusCardTheme = context.vpnStatusCardTheme;
    final settings = ref.watch(vpnSettingsControllerProvider).valueOrNull;

    return Text(
      _constructLabel(settings),
      overflow: TextOverflow.ellipsis,
      style: statusCardTheme.primaryFont.copyWith(
        color: vpnStatus.isDisconnected() || vpnStatus.isConnecting()
            ? appTheme.textErrorColor
            : appTheme.successColor,
      ),
    );
  }

  String _constructLabel(ApplicationSettings? settings) {
    var connectionStatus = t.ui.notConnected;
    if (vpnStatus.isAutoConnected()) {
      connectionStatus = t.ui.autoConnected;
    } else if (vpnStatus.isConnected()) {
      connectionStatus = vpnStatus.isMeshnetRouting
          ? t.ui.meshnet
          : t.ui.connected;
    } else if (vpnStatus.isConnecting()) {
      connectionStatus = t.ui.connecting;
    }

    final serverGroup = vpnStatus.connectionParameters.group.toSpecialtyType();
    // `standardVpn` is a regular VPN connection - no special label for it.
    if (serverGroup != null && serverGroup != ServerType.standardVpn) {
      connectionStatus += " ${t.ui.to} ${labelForServerType(serverGroup)}";
    }

    if (vpnStatus.isConnecting()) {
      connectionStatus += "...";
    }

    return connectionStatus;
  }
}

final class VpnServerInfo extends ConsumerWidget {
  final VpnStatus vpnStatus;

  const VpnServerInfo({super.key, required this.vpnStatus});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final statusCardTheme = context.vpnStatusCardTheme;
    var label = t.ui.connectToVpn;
    final settings = ref.watch(vpnSettingsControllerProvider).valueOrNull;

    if (vpnStatus.isConnected()) {
      assert(
        vpnStatus.isMeshnetRouting ||
            (vpnStatus.country != null && vpnStatus.city != null),
      );

      if (vpnStatus.isMeshnetRouting) {
        label = vpnStatus.hostname ?? vpnStatus.ip ?? "";
      } else {
        final countryName = vpnStatus.country?.localizedName ?? "";
        label = "$countryName - ${vpnStatus.city ?? ""}";
        label += vpnStatus.isVirtualLocation ? " - ${t.ui.virtual}" : "";
      }
    } else if (vpnStatus.isConnecting()) {
      label = t.ui.findingServer;
    } else if (vpnStatus.isDisconnected()) {
      if (settings?.obfuscatedServers == true) {
        label += " (${t.ui.obfuscated})".toLowerCase();
      }
    }
    return Text(
      label,
      style: statusCardTheme.secondaryFont,
      overflow: TextOverflow.ellipsis,
    );
  }
}
