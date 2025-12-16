import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/servers_list_theme.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';
import 'package:nordvpn/widgets/padded_circle_avatar.dart';

final class ServerItemImage extends ConsumerWidget {
  final Widget image;
  final bool Function(VpnStatus)? shouldHighlight;

  const ServerItemImage({super.key, required this.image, this.shouldHighlight});

  @override
  // Build the icon for a server. The icon reacts to VPN status changes
  // TODO: check performance when all servers are added to the list
  Widget build(BuildContext context, WidgetRef ref) {
    final appTheme = context.appTheme;
    final serversListThemeData = context.serversListTheme;

    // when the border is missing use transparent color to ensure that the
    // flag has always the same size
    var borderColor = Colors.transparent;

    // listen on the VPN status changes
    final asyncStatus = ref.watch(vpnStatusControllerProvider);
    if (asyncStatus.hasValue) {
      final status = asyncStatus.value!;
      if (shouldHighlight?.call(status) ?? false) {
        if (status.isConnecting()) {
          // while connecting
          return Padding(
            padding: EdgeInsets.all(appTheme.flagsBorderSize),
            child: LoadingIndicator(size: serversListThemeData.loaderSize),
          );
        } else if (status.isConnected()) {
          borderColor = appTheme.successColor;
        }
      }
    }
    return PaddedCircleAvatar(
      size: serversListThemeData.flagSize,
      borderColor: borderColor,
      borderSize: appTheme.flagsBorderSize,
      child: image,
    );
  }
}
