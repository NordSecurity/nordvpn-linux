import 'package:flutter/material.dart';
import 'package:flutter_svg/svg.dart';
import 'package:nordvpn/data/models/server_group_extension.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/internal/images_manager.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/theme/connection_card_theme.dart';
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
    final iconTheme = context.connectionCardTheme.iconTheme;

    if (status.isConnected()) {
      return _buildConnectedIcon(iconTheme);
    }

    if (status.isConnecting()) {
      return _buildConnectingIcon(iconTheme);
    }

    return _buildDisconnectedIcon(iconTheme);
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

    return DynamicThemeImage("flag_placeholder.svg");
  }

  Widget _addDIPIconIfNeeded(ConnectionCardIconTheme iconTheme) {
    final serverType = status.connectionParameters.group.toSpecialtyType();
    if (serverType == null || serverType != ServerType.dedicatedIP) {
      return SizedBox.shrink();
    }

    return Positioned(
      bottom: 0,
      right: 0,
      child: SizedBox(
        width: iconTheme.dipIconWidth,
        height: iconTheme.dipIconHeight,
        child: DynamicThemeImage("dip_connected_icon.svg"), // your widget here
      ),
    );
  }

  Widget _buildConnectedIcon(ConnectionCardIconTheme iconTheme) {
    assert(
      status.country != null || status.isMeshnetRouting,
      "No country and no meshnet routing",
    );

    return SizedBox(
      width: iconTheme.iconSize,
      height: iconTheme.iconSize,
      child: Stack(
        alignment: Alignment.center,
        children: [
          PaddedCircleAvatar(
            size: iconTheme.iconSize,
            borderColor: iconTheme.borderConnectedColor,
            borderSize: iconTheme.flagBorderSize,
            child: icon(),
          ),
          _addDIPIconIfNeeded(iconTheme),
        ],
      ),
    );
  }

  Widget _buildConnectingIcon(ConnectionCardIconTheme iconTheme) {
    return SizedBox(
      width: iconTheme.iconSize,
      height: iconTheme.iconSize,
      child: Stack(
        alignment: Alignment.center,
        children: [
          CircleAvatar(
            radius: (iconTheme.iconSize / 2) - (2 * iconTheme.flagBorderSize),
            backgroundColor: Colors.transparent,
            child: ClipOval(
              child: SizedBox(
                width: double.infinity,
                height: double.infinity,
                child: icon(),
              ),
            ),
          ),
          SizedBox(
            width: iconTheme.iconSize,
            height: iconTheme.iconSize,
            child: CircularProgressIndicator(
              strokeWidth: iconTheme.flagBorderSize,
              color: iconTheme.borderConnectingColor,
            ),
          ),
          _addDIPIconIfNeeded(iconTheme),
        ],
      ),
    );
  }

  Widget _buildDisconnectedIcon(ConnectionCardIconTheme iconTheme) {
    return Container(
      width: iconTheme.iconSize,
      height: iconTheme.iconSize,
      padding: iconTheme.disconnectedPadding,
      decoration: BoxDecoration(
        color: iconTheme.disconnectedBackgroundColor,
        shape: BoxShape.circle,
      ),
      child: SvgPicture.asset(iconTheme.disconnectedIcon),
    );
  }
}
