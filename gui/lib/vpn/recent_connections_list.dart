import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/providers/recent_connections_controller.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/vpn/recent_connections_item_factory.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/service_locator.dart';

final class RecentConnectionsList extends ConsumerWidget {
  final Function(ConnectArguments) onSelected;
  final RecentConnectionsItemFactory _itemFactory;

  RecentConnectionsList({
    super.key,
    required this.onSelected,
    RecentConnectionsItemFactory? itemFactory,
  }) : _itemFactory = itemFactory ?? sl();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final appTheme = context.appTheme;
    final recentConnections = ref.watch(recentConnectionsControllerProvider);

    return recentConnections.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, stack) => Center(child: Text('Error: $err')),
      data: (connections) {
        if (connections.isEmpty) {
          return const SizedBox.shrink();
        }

        return Padding(
          padding: EdgeInsets.symmetric(
            horizontal: appTheme.horizontalSpace,
            vertical: appTheme.verticalSpaceSmall,
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            spacing: appTheme.verticalSpaceSmall,
            children: [
              Text(t.ui.recentConnections, style: appTheme.caption),
              Column(
                children: connections
                    .map(
                      (connection) => _itemFactory.forRecentConnection(
                        context: context,
                        model: connection,
                        onTap: onSelected,
                      ),
                    )
                    .toList(),
              ),
            ],
          ),
        );
      },
    );
  }
}
