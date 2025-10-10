import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/server_group_extension.dart';
import 'package:nordvpn/data/providers/recent_connections_controller.dart';
import 'package:nordvpn/data/models/recent_connections.dart';
import 'package:nordvpn/i18n/string_translation_extension.dart';
import 'package:nordvpn/vpn/server_list_item_factory.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/internal/images_manager.dart';

class RecentConnectionsList extends ConsumerWidget {
  final Function(ConnectArguments) onSelected;
  const RecentConnectionsList({super.key, required this.onSelected});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final appTheme = context.appTheme;
    final recentConnections = ref.watch(recentConnectionsControllerProvider);
    final serverFactory = ServerListItemFactory(imagesManager: ImagesManager());

    return recentConnections.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, stack) => Center(child: Text('Error: $err')),
      data: (connections) {
        if (connections.isEmpty) {
          return const SizedBox.shrink();
        }

        return Padding(
          padding: const EdgeInsets.symmetric(vertical: 8.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text("Recent connections".tr(), style: appTheme.caption),
              const SizedBox(height: 8.0),
              ListView.builder(
                shrinkWrap: true,
                itemCount: connections.length,
                itemBuilder: (context, index) {
                  final connection = connections[index];
                  return serverFactory.forRecent(
                    recentConnection: connection,
                    onTapFunc: onSelected,
                    enabled: true,
                  );
                },
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
      return ImagesManager().forSpecialtyServer(specialtyType);
    }

    if (conn.isCountryBased) {
      final country = Country(code: conn.countryCode, name: conn.country);
      return ImagesManager().forCountry(country);
    }

    return const Icon(Icons.history);
  }
}
