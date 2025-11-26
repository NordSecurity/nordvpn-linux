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

class RecentServerListItem extends StatelessWidget {
  // Matches server id from specific server name
  // e.g., "Lithuania #123" -> captures "#123"
  static final _serverIdRegex = RegExp(r'^[A-Za-z\s-]+(#\d+)$');

  final RecentConnection model;
  final void Function(ConnectArguments) onTap;
  final ImagesManager imagesManager;

  const RecentServerListItem({
    super.key,
    required this.model,
    required this.onTap,
    required this.imagesManager,
  });

  @override
  Widget build(BuildContext context) {
    final isSpecialtyServer =
        model.group != ServerGroup.UNDEFINED &&
        model.group != ServerGroup.STANDARD_VPN_SERVERS;

    // Pre-compute connect arguments to avoid recalculation on each tap
    final connectArgs = _buildItemConnectArgs(isSpecialtyServer);

    return CustomListTile(
      minTileHeight: context.serversListTheme.listItemHeight,
      contentPadding: EdgeInsets.only(left: 0),
      leading: ServerItemImage(image: _buildItemImage(isSpecialtyServer)),
      title: _buildItemTitle(context, isSpecialtyServer),
      onTap: () => onTap(connectArgs),
    );
  }

  Widget _buildItemImage(bool isSpecialtyServer) {
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
      return imagesManager.forCountry(
        Country(code: model.countryCode, name: model.country),
      );
    }

    // fallback: try to get specialty server image or default icon
    final serverType = toServerType(model.group);
    return serverType != null
        ? imagesManager.forSpecialtyServer(serverType)
        : const Icon(Icons.history);
  }

  Widget _buildItemTitle(BuildContext context, bool isSpecialtyServer) {
    final appTheme = context.appTheme;
    if (isSpecialtyServer) {
      var specialtyTitle = Text(model.specialtyServer, style: appTheme.body);
      if (model.country.isNotEmpty) {
        var subtitle = model.country;
        subtitle +=
            " - ${model.city.isEmpty ? t.ui.fastestServer : model.city}";

        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            specialtyTitle,
            Text(subtitle, style: appTheme.caption),
          ],
        );
      }

      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          specialtyTitle,
          Text(t.ui.fastestServer, style: appTheme.caption),
        ],
      );
    }

    final isCity =
        model.city.isNotEmpty &&
        model.connectionType == ServerSelectionRule.CITY;

    if (isCity) {
      final cityText = _maybeAddVirtualLabel(model.city);
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(model.country, style: appTheme.body),
          Text(cityText, style: appTheme.caption),
        ],
      );
    }

    final isSpecificServer =
        model.specificServerName.isNotEmpty &&
        model.connectionType == ServerSelectionRule.SPECIFIC_SERVER;

    if (isSpecificServer) {
      final serverId = _extractServerId();
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(model.country, style: appTheme.body),
          if (serverId != null)
            Text(_maybeAddVirtualLabel(serverId), style: appTheme.caption),
        ],
      );
    }

    String titleText = model.country;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Text(titleText, style: appTheme.body),
        Text(t.ui.fastestServer, style: appTheme.caption),
      ],
    );
  }

  /// Adds virtual label to text if the server is virtual
  String _maybeAddVirtualLabel(String text) {
    return model.isVirtual ? "$text - ${t.ui.virtual}" : text;
  }

  /// Extracts server ID from specific server name
  /// e.g., "Lithuania #123" -> "#123"
  String? _extractServerId() {
    final match = _serverIdRegex.firstMatch(model.specificServerName);
    return match?[1];
  }

  ConnectArguments _buildItemConnectArgs(bool isSpecialtyServer) {
    if (model.connectionType == ServerSelectionRule.SPECIFIC_SERVER &&
        model.specificServerName.isNotEmpty) {
      return ConnectArguments(
        server: ServerInfo(
          id: 0,
          hostname: model.specificServerName,
          isVirtual: model.isVirtual,
        ),
      );
    } else {
      Country? country;
      if (model.countryCode.isNotEmpty && model.country.isNotEmpty) {
        country = Country(code: model.countryCode, name: model.country);
      }
      return ConnectArguments(
        country: country,
        city: model.city.isNotEmpty ? City(model.city) : null,
        specialtyGroup: isSpecialtyServer ? toServerType(model.group) : null,
      );
    }
  }
}
