import 'package:nordvpn/data/models/allow_list.dart';
import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/server_group_extension.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/data/providers/consent_status_provider.dart';
import 'package:nordvpn/pb/daemon/config/analytics_consent.pbenum.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:freezed_annotation/freezed_annotation.dart';

part 'app_settings.freezed.dart';

// Class to store application settings
@freezed
abstract class ApplicationSettings with _$ApplicationSettings {
  const ApplicationSettings._();

  const factory ApplicationSettings({
    required bool notifications,
    required ConsentLevel analyticsConsent,
    required bool autoConnect,
    ConnectArguments? autoConnectLocation,
    required VpnProtocol protocol,
    required bool killSwitch,
    required bool lanDiscovery,
    required bool routing,
    required bool postQuantum,
    required bool obfuscatedServers,
    required bool virtualServers,
    required bool firewall,
    required int firewallMark,
    required bool customDns,
    required List<String> customDnsServers,
    required bool threatProtection,
    required bool tray,
    required bool allowList,
    required AllowList allowListData,
  }) = _ApplicationSettings;

  factory ApplicationSettings.fromSettings(Settings settings) {
    // Compose the auto-connect information
    ConnectArguments? autoConnectData;
    if (settings.autoConnectData.enabled &&
        (settings.autoConnectData.hasCountry() ||
            settings.autoConnectData.hasCity() ||
            settings.autoConnectData.hasServerGroup())) {
      final countryCode = settings.autoConnectData.country;
      autoConnectData = ConnectArguments(
        country: countryCode.isNotEmpty ? Country.fromCode(countryCode) : null,
        city: settings.autoConnectData.city.isNotEmpty
            ? City(settings.autoConnectData.city)
            : null,
        specialtyGroup: settings.autoConnectData.hasServerGroup()
            ? settings.autoConnectData.serverGroup.toSpecialtyType()
            : null,
      );
    }

    final allowList = AllowList.fromSettings(settings.allowlist);
    return ApplicationSettings(
      notifications: settings.userSettings.notify,
      analyticsConsent: _convertToConsentLevel(settings.analyticsConsent),
      autoConnect: settings.autoConnectData.enabled,
      autoConnectLocation: autoConnectData,
      protocol: convertToVpnProtocol(settings.technology, settings.protocol),
      killSwitch: settings.killSwitch,
      lanDiscovery: settings.lanDiscovery,
      routing: settings.routing,
      postQuantum: settings.postquantumVpn,
      obfuscatedServers: settings.obfuscate,
      virtualServers: settings.virtualLocation,
      firewall: settings.firewall,
      firewallMark: settings.fwmark,
      customDns: settings.dns.isNotEmpty,
      customDnsServers: settings.dns,
      threatProtection: settings.threatProtectionLite,
      tray: settings.userSettings.tray,
      allowList: allowList.isNotEmpty,
      allowListData: allowList,
    );
  }

  bool areDipServersSupported() {
    return !obfuscatedServers &&
        ((protocol == VpnProtocol.nordlynx) ||
            (protocol == VpnProtocol.openVpnUdp) ||
            (protocol == VpnProtocol.openVpnTcp));
  }
}

ConsentLevel _convertToConsentLevel(ConsentMode analyticsConsent) {
  return switch (analyticsConsent) {
    ConsentMode.GRANTED => ConsentLevel.acceptedAll,
    ConsentMode.DENIED => ConsentLevel.essentialOnly,
    ConsentMode.UNDEFINED => ConsentLevel.none,
    _ => throw "not supported analytics consent level: $analyticsConsent",
  };
}
