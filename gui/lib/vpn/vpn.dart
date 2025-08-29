import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/vpn/servers_list_card.dart';
import 'package:nordvpn/vpn/vpn_status_card.dart';
import 'package:nordvpn/widgets/round_container.dart';

// VPN screen
final class VpnWidget extends ConsumerWidget {
  const VpnWidget({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = context.appTheme;
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      spacing: theme.verticalSpaceSmall,
      children: [
        VpnStatusCard(),
        Expanded(
          child: RoundContainer(
            margin: EdgeInsets.only(
              top: 0,
              bottom: theme.margin,
              right: theme.margin,
              left: theme.margin,
            ),
            child: ServersListCard(
              onSelected: (args) async {
                await _connect(ref, args);
              },
            ),
          ),
        ),
      ],
    );
  }

  Future<void> _connect(WidgetRef ref, ConnectArguments args) async {
    final vpnController = ref.read(vpnStatusControllerProvider.notifier);
    await vpnController.connect(args);
  }
}
