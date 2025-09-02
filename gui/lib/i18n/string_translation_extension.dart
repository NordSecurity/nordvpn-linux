import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';

extension StringTranslation on String {
  // Returns the translated version of the string.
  // When is not found in translations, returns the defaultValue parameter or the original string when defaultValue is null.
  // This should only be used for dynamic strings, because of performance issues.
  String tr([String? defaultValue]) {
    var val = t[this];
    if (val != null) {
      return val;
    }

    val = t[this];
    if (val != null) {
      return val;
    }

    final key = toLowerCase().replaceAll(" ", "_");
    val = t[key];
    if (val != null) {
      return val;
    }
    logger.d(
      "Missing translation \"$this\" for [${t.$meta.locale.languageCode}]",
    );
    return defaultValue ?? this;
  }
}

String labelForServerType(ServerType type) {
  switch (type) {
    case ServerType.p2p:
      return t.ui.p2p;

    case ServerType.doubleVpn:
      return t.ui.doubleVpn;

    case ServerType.onionOverVpn:
      return t.ui.onionOverVpn;

    case ServerType.dedicatedIP:
      return t.ui.dedicatedIp;

    case ServerType.standardVpn:
      return t.ui.standardVpn;

    case ServerType.obfuscated:
      return t.ui.obfuscated;
  }
}
