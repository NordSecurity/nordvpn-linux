import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/vpn/connection_card_buttons.dart';
import 'package:nordvpn/vpn/connection_card_icon.dart';
import 'package:nordvpn/vpn/connection_card_label.dart';
import 'package:nordvpn/vpn/connection_card_server_info.dart';
import 'package:nordvpn/vpn/servers_list_card.dart';
import 'package:nordvpn/widgets/custom_error_widget.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';

final class TrayAppWidget extends ConsumerWidget {
  const TrayAppWidget({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final appTheme = context.appTheme;

    return Scaffold(
      backgroundColor: appTheme.backgroundColor,
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _TrayHeader(),
          Divider(height: 1, thickness: 1, color: appTheme.dividerColor),
          _TrayStatusSection(),
          Expanded(
            child: ServersListCard(
              onSelected: (ConnectArguments args) async {
                await ref.read(vpnStatusControllerProvider.notifier).connect(args);
              },
            ),
          ),
        ],
      ),
    );
  }
}

final class _TrayHeader extends StatelessWidget {
  const _TrayHeader();

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;

    return Padding(
      padding: EdgeInsets.symmetric(
        horizontal: appTheme.padding,
        vertical: appTheme.verticalSpaceSmall,
      ),
      child: Row(
        children: [
          _OpenAppButton(),
          const Spacer(),
          IconButton(
            icon: SizedBox(
              width: 22,
              height: 22,
              child: DynamicThemeImage("notifications_off.svg"),
            ),
            onPressed: () => context.navigateToRoute(AppRoute.settings),
            padding: const EdgeInsets.all(6),
            constraints: const BoxConstraints(),
          ),
          const SizedBox(width: 4),
          IconButton(
            icon: SizedBox(
              width: 22,
              height: 22,
              child: DynamicThemeImage("account.svg"),
            ),
            onPressed: () => context.navigateToRoute(AppRoute.settingsAccount),
            padding: const EdgeInsets.all(6),
            constraints: const BoxConstraints(),
          ),
        ],
      ),
    );
  }
}

final class _OpenAppButton extends StatelessWidget {
  const _OpenAppButton();

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;

    return OutlinedButton.icon(
      icon: SizedBox(
        width: 20,
        height: 20,
        child: DynamicThemeImage("home_on.svg"),
      ),
      label: Text(
        t.ui.openApp,
        style: appTheme.body,
      ),
      style: OutlinedButton.styleFrom(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        shape: const StadiumBorder(),
        side: BorderSide(color: appTheme.borderColor),
        foregroundColor: appTheme.backgroundColor,
      ),
      onPressed: () => context.navigateToRoute(AppRoute.vpn),
    );
  }
}

final class _TrayStatusSection extends ConsumerWidget {
  const _TrayStatusSection();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final appTheme = context.appTheme;

    return Padding(
      padding: EdgeInsets.fromLTRB(
        appTheme.padding,
        appTheme.verticalSpaceSmall,
        appTheme.padding,
        appTheme.verticalSpaceSmall,
      ),
      child: ref
          .watch(vpnStatusControllerProvider)
          .when(
            data: (vpnStatus) => _buildContent(context, ref, vpnStatus),
            loading: () => const SizedBox(height: 80, child: LoadingIndicator()),
            error: (e, _) => CustomErrorWidget(message: "$e"),
          ),
    );
  }

  Widget _buildContent(BuildContext context, WidgetRef ref, VpnStatus vpnStatus) {
    final appTheme = context.appTheme;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: appTheme.verticalSpaceSmall,
      children: [
        Row(
          spacing: appTheme.horizontalSpace,
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            ConnectionCardIcon(status: vpnStatus),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  ConnectionCardServerInfo(vpnStatus: vpnStatus),
                  ConnectionCardLabel(vpnStatus: vpnStatus),
                ],
              ),
            ),
          ],
        ),
        ConnectionCardButtons(vpnStatus: vpnStatus),
      ],
    );
  }
}
