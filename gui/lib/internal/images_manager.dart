import 'package:flutter/material.dart';
import 'package:flutter_svg/svg.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

const placeholderFlag = "placeholder";

final class ImagesManager {
  Widget forCountry(Country country, {double? width, double? height}) {
    assert(country.code.length == 2);
    final cc = country.code.toLowerCase();
    return SvgPicture.asset(
      _flagPath(cc),
      width: width,
      height: height,
      errorBuilder: (_, __, st) {
        logger.d("error loading '${_flagPath(cc)}': $st");
        logger.d("showing '${_flagPath(placeholderFlag)}' instead");
        return placeholderCountryFlag;
      },
      placeholderBuilder: (_) => placeholderCountryFlag,
    );
  }

  String _flagPath(String fileName) => "assets/images/flags/$fileName.svg";

  Widget get placeholderCountryFlag =>
      SvgPicture.asset(_flagPath(placeholderFlag));

  Widget forSpecialtyServer(ServerType type) {
    switch (type) {
      case ServerType.p2p:
        return DynamicThemeImage("p2p.svg");

      case ServerType.doubleVpn:
        return DynamicThemeImage("double_vpn.svg");

      case ServerType.onionOverVpn:
        return DynamicThemeImage("onion_over_vpn.svg");

      case ServerType.dedicatedIP:
        return DynamicThemeImage("dedicated_ip.svg");

      case ServerType.standardVpn:
        return DynamicThemeImage("standard_vpn.svg");

      case ServerType.obfuscated:
        return DynamicThemeImage("obfuscated_servers.svg");
    }
  }
}
