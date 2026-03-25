import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/vpn_status_card_theme.dart';
import 'package:nordvpn/vpn/connection_card_buttons.dart';
import 'package:nordvpn/vpn/connection_card_icon.dart';
import 'package:nordvpn/vpn/connection_card_label.dart';
import 'package:nordvpn/vpn/connection_card_server_info.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/round_container.dart';

final class ConnectionCard extends StatelessWidget {
  final ImagesManager imagesManager;

  ConnectionCard({super.key, ImagesManager? imagesManager})
    : imagesManager = imagesManager ?? sl();

  @override
  Widget build(BuildContext context) {
    final statusCardTheme = context.vpnStatusCardTheme;
    final appTheme = context.appTheme;

    return RoundContainer(
      minHeight: statusCardTheme.height,
      radius: 20,
      padding: statusCardTheme.connectionCardPadding,
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
    final connectionCardTheme = context.vpnStatusCardTheme;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      key: ValueKey(vpnStatus.status),
      children: [
        Row(
          children: [
            ConnectionCardIcon(status: vpnStatus),
            SizedBox(width: connectionCardTheme.smallSpacing),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                ConnectionCardServerInfo(vpnStatus: vpnStatus),
                ConnectionCardLabel(vpnStatus: vpnStatus),
              ],
            ),
          ],
        ),
        SizedBox(height: connectionCardTheme.mediumSpacing),
        ConnectionCardButtons(vpnStatus: vpnStatus),
      ],
    );
  }
}
