import 'package:flutter/material.dart';
import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/recent_connections.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/pb/daemon/config/group.pb.dart';
import 'package:nordvpn/pb/daemon/server_selection_rule.pb.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/servers_list_theme.dart';
import 'package:nordvpn/vpn/server_item_image.dart';
import 'package:nordvpn/widgets/custom_list_tile.dart';

/// The two title lines shown on a recent connection item.
typedef _TitleParts = ({String primary, String? secondary});

/// Factory for building list items for recent connections
final class RecentConnectionsItemFactory {
  final ImagesManager imagesManager;

  RecentConnectionsItemFactory({required this.imagesManager});

  /// Build a list item for a recent connection
  Widget forRecentConnection({
    required BuildContext context,
    required RecentConnection model,
    required void Function(ConnectArguments) onTap,
  }) {
    final appTheme = context.appTheme;
    final serversListTheme = context.serversListTheme;

    final isSpecialtyServer = _isSpecialtyServer(model);

    // Pre-compute connect arguments to avoid recalculation on each tap
    final connectArgs = _buildConnectArgs(model, isSpecialtyServer);
    final titleParts = _buildTitleParts(model, isSpecialtyServer);

    return MergeSemantics(
      child: CustomListTile(
        minTileHeight: serversListTheme.listItemHeight,
        contentPadding: EdgeInsets.only(left: 0),
        leading: ServerItemImage(image: _buildImage(model, isSpecialtyServer)),
        title: Semantics(
          label: semanticsLabelFor(model),
          button: true,
          excludeSemantics: true,
          child: _buildTitle(appTheme, titleParts),
        ),
        onTap: () => onTap(connectArgs),
      ),
    );
  }

  bool _isSpecialtyServer(RecentConnection model) =>
      model.group != ServerGroup.UNDEFINED &&
      model.group != ServerGroup.STANDARD_VPN_SERVERS;

  String semanticsLabelFor(RecentConnection model) {
    final parts = _buildTitleParts(model, _isSpecialtyServer(model));
    final details = parts.secondary == null
        ? parts.primary
        : "${parts.primary}, ${parts.secondary}";
    return "${t.ui.recentConnections}: $details";
  }

  Widget _buildImage(RecentConnection model, bool isSpecialtyServer) {
    final isCountry = model.countryCode.isNotEmpty && model.country.isNotEmpty;

    // early return for specialty server without country
    if (isSpecialtyServer && !isCountry) {
      final serverType = toServerType(model.group);
      return serverType != null
          ? imagesManager.forSpecialtyServer(serverType)
          : const Icon(Icons.history);
    }

    // handle country-based images (works for both specialty and standard servers)
    if (isCountry) {
      return imagesManager.forCountry(Country.fromCode(model.countryCode));
    }

    // fallback: try to get specialty server image or default icon
    final serverType = toServerType(model.group);
    return serverType != null
        ? imagesManager.forSpecialtyServer(serverType)
        : const Icon(Icons.history);
  }

  _TitleParts _buildTitleParts(RecentConnection model, bool isSpecialtyServer) {
    if (isSpecialtyServer) {
      if (model.country.isNotEmpty) {
        final country = Country.fromCode(model.countryCode);
        final city = model.city;
        final location = city.isEmpty ? t.ui.fastest : City(city).localizedName;
        final subtitle = "${country.localizedName} - $location";

        return (
          primary: model.specialtyServer,
          secondary: _maybeAddVirtualLabel(subtitle, model.isVirtual),
        );
      }

      return (primary: model.specialtyServer, secondary: t.ui.fastest);
    }

    final country = Country.fromCode(model.countryCode);

    final isCity =
        model.city.isNotEmpty &&
        model.connectionType == ServerSelectionRule.CITY;

    if (isCity) {
      return (
        primary: country.localizedName,
        secondary: _maybeAddVirtualLabel(
          City(model.city).localizedName,
          model.isVirtual,
        ),
      );
    }

    final isSpecificServer =
        model.specificServerName.isNotEmpty &&
        model.connectionType == ServerSelectionRule.SPECIFIC_SERVER;

    if (isSpecificServer) {
      final serverId = model.serverId;
      return (
        primary: country.localizedName,
        secondary: serverId != null
            ? _maybeAddVirtualLabel(serverId, model.isVirtual)
            : null,
      );
    }

    return (primary: country.localizedName, secondary: t.ui.fastest);
  }

  Widget _buildTitle(AppTheme appTheme, _TitleParts parts) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Text(parts.primary, style: appTheme.body),
        if (parts.secondary != null)
          Text(parts.secondary!, style: appTheme.caption),
      ],
    );
  }

  String _maybeAddVirtualLabel(String text, bool isVirtual) {
    return isVirtual ? "$text - ${t.ui.virtual}" : text;
  }

  ConnectArguments _buildConnectArgs(
    RecentConnection model,
    bool isSpecialtyServer,
  ) {
    if (model.connectionType == ServerSelectionRule.SPECIFIC_SERVER &&
        model.specificServer.isNotEmpty) {
      return ConnectArguments(
        server: ServerInfo(
          id: 0,
          hostname: "${model.specificServer}.nordvpn.com",
          isVirtual: model.isVirtual,
        ),
      );
    } else {
      Country? country;
      if (model.countryCode.isNotEmpty) {
        country = Country.fromCode(model.countryCode);
      }
      return ConnectArguments(
        country: country,
        city: model.city.isNotEmpty ? City(model.city) : null,
        specialtyGroup: isSpecialtyServer ? toServerType(model.group) : null,
      );
    }
  }
}
