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
  final RecentConnection model;
  final void Function(ConnectArguments) onTap;
  final bool enabled;
  final ImagesManager imagesManager;

  const RecentServerListItem({
    super.key,
    required this.model,
    required this.onTap,
    this.enabled = true,
    required this.imagesManager,
  });

  @override
  Widget build(BuildContext context) {
    final isSpecialtyServer =
        model.group != ServerGroup.UNDEFINED &&
        model.group != ServerGroup.STANDARD_VPN_SERVERS;

    final image = _buildItemImage(isSpecialtyServer);
    final title = _buildItemTitle(context, isSpecialtyServer);
    final connectArgs = _buildItemConnectArgs(isSpecialtyServer);

    return CustomListTile(
      minTileHeight: context.serversListTheme.listItemHeight,
      contentPadding: EdgeInsets.only(left: 0),
      leading: ServerItemImage(image: image),
      title: title,
      onTap: enabled ? () => onTap(connectArgs) : null,
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

    Text maybeAddVirtualLabel(Text text) {
      return model.isVirtual
          ? Text("${text.data} - ${t.ui.virtual}", style: text.style)
          : text;
    }

    final isCity =
        model.city.isNotEmpty &&
        model.connectionType == ServerSelectionRule.CITY;

    if (isCity) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(model.country, style: appTheme.body),
          maybeAddVirtualLabel(Text(model.city, style: appTheme.caption)),
        ],
      );
    }

    final isSpecificServer =
        model.specificServerName.isNotEmpty &&
        model.connectionType == ServerSelectionRule.SPECIFIC_SERVER;

    if (isSpecificServer) {
      // match only server id
      final serverIdMatch = RegExp(
        r'^[A-Za-z\s-]+(#\d+)$',
      ).firstMatch(model.specificServerName);

      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(model.country, style: appTheme.body),
          if (serverIdMatch != null)
            maybeAddVirtualLabel(
              Text(serverIdMatch[1]!, style: appTheme.caption),
            ),
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
