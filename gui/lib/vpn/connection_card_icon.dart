import 'package:flutter/material.dart';
import 'package:flutter_svg/svg.dart';
import 'package:nordvpn/data/models/server_group_extension.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/vpn_status_card_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/padded_circle_avatar.dart';

final class ConnectionCardIcon extends StatelessWidget {
  final ImagesManager imagesManager;
  final VpnStatus status;

  ConnectionCardIcon({
    super.key,
    required this.status,
    ImagesManager? imagesManager,
  }) : imagesManager = imagesManager ?? sl();

  @override
  Widget build(BuildContext context) {
    final statusCardTheme = context.vpnStatusCardTheme;

    if (status.isConnected()) {
      assert(status.country != null || status.isMeshnetRouting);
      final appTheme = context.appTheme;

      return PaddedCircleAvatar(
        size: statusCardTheme.iconSize,
        borderColor: appTheme.successColor,
        borderSize: appTheme.flagsBorderSize,
        child: icon(),
      );
    }
    if (status.isConnecting()) {
      return LoadingIndicator(size: statusCardTheme.iconSize);
    }

    return _buildDisconnectedIcon(context);
  }

  Widget icon() {
    if (status.isMeshnetRouting) {
      return DynamicThemeImage("linux_peer.svg");
    }

    // Prioritize showing the country flag if a country is available in the status
    if (status.country != null) {
      return imagesManager.forCountry(status.country!);
    }

    // Fallback to specialty group icon if no country is available
    final serverType = status.connectionParameters.group.toSpecialtyType();
    if (serverType != null && serverType != ServerType.standardVpn) {
      return imagesManager.forSpecialtyServer(serverType);
    }

    return imagesManager.placeholderCountryFlag;
  }

  Widget _buildDisconnectedIcon(BuildContext context) {
    final connectionCardTheme = context.vpnStatusCardTheme;

    return Container(
      width: 48,
      height: 48,
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: connectionCardTheme.iconBackgroundColor,
        shape: BoxShape.circle,
      ),
      child: SvgPicture.asset(connectionCardTheme.disconnectedIcon),
    );
  }
}
