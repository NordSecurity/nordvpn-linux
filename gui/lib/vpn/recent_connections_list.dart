import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/server_group_extension.dart';
import 'package:nordvpn/data/providers/recent_connections_controller.dart';
import 'package:nordvpn/data/models/recent_connections.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/vpn/server_list_item_factory.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/service_locator.dart';

final class RecentConnectionsList extends ConsumerWidget {
  final Function(ConnectArguments) onSelected;
  final ServerListItemFactory _serverListItemFactory;
  final ImagesManager _imagesManager;
  RecentConnectionsList({
    super.key,
    required this.onSelected,
    ServerListItemFactory? itemFactory,
    ImagesManager? imagesManager,
  }) : _serverListItemFactory = itemFactory ?? sl(),
       _imagesManager = imagesManager ?? sl();

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
          padding: EdgeInsets.symmetric(vertical: appTheme.verticalSpaceSmall),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            spacing: appTheme.verticalSpaceSmall,
            children: [
              Text(t.ui.recentConnections, style: appTheme.caption),
              Column(
                children: connections
                    .map(
                      (connection) => _serverListItemFactory.forRecent(
                        recentConnection: connection,
                        onTapFunc: onSelected,
                        enabled: true,
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

  Widget addModelBasedIcon(RecentConnection conn) {
    if (conn.isSpecialtyServer) {
      final specialtyType = conn.group.toSpecialtyType();
      if (specialtyType == null) {
        return const Icon(Icons.history);
      }
      return _imagesManager.forSpecialtyServer(specialtyType);
    }

    if (conn.isCountryBased) {
      final country = Country(code: conn.countryCode, name: conn.country);
      return _imagesManager.forCountry(country);
    }

    return const Icon(Icons.history);
  }
}
