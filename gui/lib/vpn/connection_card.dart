import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/theme/connection_card_theme.dart';
import 'package:nordvpn/vpn/connection_card_buttons.dart';
import 'package:nordvpn/vpn/connection_card_icon.dart';
import 'package:nordvpn/vpn/connection_card_label.dart';
import 'package:nordvpn/vpn/connection_card_server_info.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/round_container.dart';

final class ConnectionCard extends StatelessWidget {
  const ConnectionCard({super.key});

  @override
  Widget build(BuildContext context) {
    final connectionCardTheme = context.connectionCardTheme;

    return RoundContainer(
      radius: connectionCardTheme.borderRadius,
      padding: connectionCardTheme.mapPadding,
      margin: connectionCardTheme.margin,
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
    final connectionCardTheme = context.connectionCardTheme;

    return ConstrainedBox(
      constraints: BoxConstraints(minWidth: connectionCardTheme.minWidth),
      child: Container(
        padding: connectionCardTheme.connectionCardPadding,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          key: ValueKey(vpnStatus.status),
          children: [
            Row(
              spacing: connectionCardTheme.smallSpacing,
              children: [
                ConnectionCardIcon(status: vpnStatus),
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
        ),
      ),
    );
  }
}
