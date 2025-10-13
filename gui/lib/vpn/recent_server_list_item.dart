import 'package:flutter/material.dart';
import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/recent_connections.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/pb/daemon/config/group.pb.dart';
import 'package:nordvpn/pb/daemon/server_selection_rule.pb.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/servers_list_theme.dart';
import 'package:nordvpn/vpn/server_item_image.dart';
import 'package:nordvpn/widgets/custom_expansion_tile.dart';

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

    return CustomExpansionTile(
      minTileHeight: context.serversListTheme.listItemHeight,
      leading: ServerItemImage(
        image: image,
        shouldHighlight: (status) =>
            _shouldHighlightItem(status, isSpecialtyServer),
      ),
      title: title,
      onTap: enabled ? () => onTap(connectArgs) : null,
      hideExpandButton: true,
      expanded: false,
    );
  }

  bool _shouldHighlightItem(VpnStatus status, bool isSpecialtyServer) {
    if (isSpecialtyServer) {
      return status.connectionParameters.group == model.group;
    }

    if (model.connectionType == ServerSelectionRule.SPECIFIC_SERVER) {
      return status.hostname == model.specificServerName;
    }

    final countryMatches = status.country?.code == model.countryCode;
    if (model.connectionType == ServerSelectionRule.CITY) {
      return countryMatches && status.city?.name == model.city;
    }

    return countryMatches;
  }

  Widget _buildItemImage(bool isSpecialtyServer) {
    if (isSpecialtyServer) {
      final serverType = toServerType(model.group);
      if (serverType != null) {
        return imagesManager.forSpecialtyServer(serverType);
      }
      return const Icon(Icons.history);
    }

    Country? country;
    final isCountry = model.countryCode.isNotEmpty && model.country.isNotEmpty;
    if (isCountry) {
      country = Country(code: model.countryCode, name: model.country);
    }

    if (country != null) {
      return imagesManager.forCountry(country);
    }

    final serverType = toServerType(model.group);
    if (serverType != null) {
      return imagesManager.forSpecialtyServer(serverType);
    }
    return const Icon(Icons.history);
  }

  Widget _buildItemTitle(BuildContext context, bool isSpecialtyServer) {
    final appTheme = context.appTheme;
    if (isSpecialtyServer) {
      var specialtyTitle = Text(model.specialtyServer, style: appTheme.body);
      if (model.country.isNotEmpty) {
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            specialtyTitle,
            Text(model.country, style: appTheme.caption),
          ],
        );
      }

      return specialtyTitle;
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
          Text(model.city, style: appTheme.caption),
        ],
      );
    }

    String? subtitleText;
    if (model.connectionType == ServerSelectionRule.SPECIFIC_SERVER) {
      subtitleText = model.specificServerName;
    }

    String titleText = model.country;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Text(titleText, style: appTheme.body),
        if (subtitleText != null) Text(subtitleText, style: appTheme.caption),
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
          isVirtual: false,
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
