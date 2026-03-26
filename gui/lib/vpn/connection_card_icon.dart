import 'package:flutter/material.dart';
import 'package:flutter_svg/svg.dart';
import 'package:nordvpn/data/models/server_group_extension.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/vpn_status_card_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
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
    final iconTheme = context.vpnStatusCardTheme.iconStyle;

    if (status.isConnected()) {
      assert(status.country != null || status.isMeshnetRouting);

      return PaddedCircleAvatar(
        size: iconTheme.iconSize,
        borderColor: iconTheme.borderConnectedColor,
        borderSize: iconTheme.flagBorderSize,
        child: icon(),
      );
    }
    if (status.isConnecting()) {
      return _buildConnectingIcon(context, iconTheme);
    }

    return _buildDisconnectedIcon(context, iconTheme);
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

  Widget _buildConnectingIcon(
    BuildContext context,
    ConnectionCardIconThemeStyle iconTheme,
  ) {
    return SizedBox(
      width: iconTheme.iconSize,
      height: iconTheme.iconSize,
      child: Stack(
        alignment: Alignment.center,
        children: [
          // Avatar
          CircleAvatar(
            radius: (iconTheme.iconSize / 2) - (2 * iconTheme.flagBorderSize),
            child: ClipOval(
              child: SizedBox(
                width: double.infinity,
                height: double.infinity,
                child: icon(),
              ),
            ),
          ),

          // Animated border
          SizedBox(
            width: iconTheme.iconSize,
            height: iconTheme.iconSize,
            child: CircularProgressIndicator(
              strokeWidth: iconTheme.strokeWidth,
              color: iconTheme.borderConnectingColor,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDisconnectedIcon(
    BuildContext context,
    ConnectionCardIconThemeStyle iconTheme,
  ) {
    return Container(
      width: iconTheme.iconSize,
      height: iconTheme.iconSize,
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: iconTheme.disconnectedBackgroundColor,
        shape: BoxShape.circle,
      ),
      child: SvgPicture.asset(iconTheme.disconnectedIcon),
    );
  }
}
