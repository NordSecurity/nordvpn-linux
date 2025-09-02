import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/allow_list.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/grpc/grpc_service.dart';
import 'package:nordvpn/pb/daemon/common.pb.dart';
import 'package:nordvpn/pb/daemon/config/protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';
import 'package:nordvpn/pb/daemon/defaults.pb.dart';
import 'package:nordvpn/pb/daemon/service.pbgrpc.dart';
import 'package:nordvpn/pb/daemon/set.pb.dart';
import 'package:fixnum/fixnum.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'vpn_settings_repository.g.dart';

// Communicates with the daemon to get and set the settings
class VpnSettingsRepository {
  final DaemonClient _client;

  VpnSettingsRepository([DaemonClient? client])
    : _client = client ?? createDaemonClient();

  Future<ApplicationSettings> fetchSettings() async {
    final response = await _client.settings(Empty());
    final settings = response.data;
    return ApplicationSettings.fromSettings(settings);
  }

  Future<int> setVpnProtocol(VpnProtocol vpnProtocol) async {
    final (technology, protocol) = toTechnologyAndProtocol(vpnProtocol);

    final technologyStatus = await _setTechnology(technology);
    if ((technologyStatus != DaemonStatusCode.success) &&
        (technologyStatus != DaemonStatusCode.nothingToDo) &&
        (technologyStatus != DaemonStatusCode.vpnIsRunning)) {
      return technologyStatus;
    }

    final protocolStatus = await _setProtocol(protocol);
    if ((protocolStatus != DaemonStatusCode.success) &&
        (protocolStatus != DaemonStatusCode.nothingToDo) &&
        (protocolStatus != DaemonStatusCode.vpnIsRunning)) {
      return protocolStatus;
    }

    if (technologyStatus == DaemonStatusCode.vpnIsRunning ||
        protocolStatus == DaemonStatusCode.vpnIsRunning) {
      return DaemonStatusCode.vpnIsRunning;
    }

    return DaemonStatusCode.success;
  }

  Future<int> _setTechnology(Technology technology) async {
    final response = await _client.setTechnology(
      SetTechnologyRequest(technology: technology),
    );
    return _checkSettingsUpdate(response);
  }

  Future<int> _setProtocol(Protocol protocol) async {
    final response = await _client.setProtocol(
      SetProtocolRequest(protocol: protocol),
    );

    if (response.hasErrorCode()) {
      final code = _codeToDaemonStatusCode(response.errorCode);
      if (code != DaemonStatusCode.success) {
        return code;
      }
    }

    if (response.hasSetProtocolStatus()) {
      switch (response.setProtocolStatus) {
        case SetProtocolStatus.INVALID_TECHNOLOGY:
          return DaemonStatusCode.invalidTechnology;
        case SetProtocolStatus.PROTOCOL_CONFIGURED:
          break;
        case SetProtocolStatus.PROTOCOL_CONFIGURED_VPN_ON:
          return DaemonStatusCode.vpnIsRunning;
      }
    }

    return DaemonStatusCode.success;
  }

  Future<int> setObfuscated(bool value) async {
    final result = await _client.setObfuscate(
      SetGenericRequest(enabled: value),
    );
    return _checkSettingsUpdate(result);
  }

  Future<int> setAnalytics(bool value) async {
    final result = await _client.setAnalytics(
      SetGenericRequest(enabled: value),
    );
    return result.type.toInt();
  }

  Future<int> setFirewall(bool value) async {
    final result = await _client.setFirewall(SetGenericRequest(enabled: value));
    return result.type.toInt();
  }

  Future<int> setNotifications(bool value) async {
    final result = await _client.setNotify(SetNotifyRequest(notify: value));
    return result.type.toInt();
  }

  Future<int> setFirewallMark(int value) async {
    final result = await _client.setFirewallMark(
      SetUint32Request(value: value),
    );
    return result.type.toInt();
  }

  Future<int> setKillSwitch(bool value) async {
    final result = await _client.setKillSwitch(
      SetKillSwitchRequest(killSwitch: value),
    );
    return result.type.toInt();
  }

  Future<int> setLocalNetwork(bool value) async {
    final response = await _client.setLANDiscovery(
      SetLANDiscoveryRequest(enabled: value),
    );

    if (response.hasErrorCode()) {
      final code = _codeToDaemonStatusCode(response.errorCode);
      if (code != DaemonStatusCode.success) {
        return code;
      }
    }

    if (response.hasSetLanDiscoveryStatus()) {
      switch (response.setLanDiscoveryStatus) {
        case SetLANDiscoveryStatus.DISCOVERY_CONFIGURED:
          break;
        case SetLANDiscoveryStatus.DISCOVERY_CONFIGURED_ALLOWLIST_RESET:
          return DaemonStatusCode.allowListModified;
      }
    }

    return DaemonStatusCode.success;
  }

  Future<int> addToAllowList({PortInterval? port, String? subnet}) async {
    final result = await _client.setAllowlist(
      SetAllowlistRequest(
        setAllowlistSubnetRequest: (subnet != null)
            ? SetAllowlistSubnetRequest(subnet: subnet)
            : null,
        setAllowlistPortsRequest: (port != null)
            ? SetAllowlistPortsRequest(
                portRange: PortRange(
                  startPort: Int64(port.start),
                  endPort: Int64(port.end),
                ),
                isTcp: port.isTcp,
                isUdp: port.isUdp,
              )
            : null,
      ),
    );

    return result.type.toInt();
  }

  Future<int> removeFromAllowList({PortInterval? port, String? subnet}) async {
    final result = await _client.unsetAllowlist(
      SetAllowlistRequest(
        setAllowlistSubnetRequest: (subnet != null)
            ? SetAllowlistSubnetRequest(subnet: subnet)
            : null,
        setAllowlistPortsRequest: (port != null)
            ? SetAllowlistPortsRequest(
                portRange: PortRange(
                  startPort: Int64(port.start),
                  endPort: Int64(port.end),
                ),
                isTcp: port.isTcp,
                isUdp: port.isUdp,
              )
            : null,
      ),
    );

    return result.type.toInt();
  }

  Future<int> disableAllowList() async {
    final result = await _client.unsetAllAllowlist(Empty());
    return result.type.toInt();
  }

  Future<int> setThreatProtection(bool value) async {
    final response = await _client.setThreatProtectionLite(
      SetThreatProtectionLiteRequest(threatProtectionLite: value),
    );

    if (response.hasErrorCode()) {
      final code = _codeToDaemonStatusCode(response.errorCode);
      if (code != DaemonStatusCode.success) {
        return code;
      }
    }

    if (response.hasSetThreatProtectionLiteStatus()) {
      switch (response.setThreatProtectionLiteStatus) {
        case SetThreatProtectionLiteStatus.TPL_CONFIGURED:
          break;
        case SetThreatProtectionLiteStatus.TPL_CONFIGURED_DNS_RESET:
          return DaemonStatusCode.dnsListModified;
      }
    }

    return DaemonStatusCode.success;
  }

  Future<int> setDns(List<String> newDnsList) async {
    final response = await _client.setDNS(SetDNSRequest(dns: newDnsList));
    if (response.hasErrorCode()) {
      final code = _codeToDaemonStatusCode(response.errorCode);
      if (code != DaemonStatusCode.success) {
        return code;
      }
    }

    if (response.hasSetDnsStatus()) {
      switch (response.setDnsStatus) {
        case SetDNSStatus.DNS_CONFIGURED:
          break;
        case SetDNSStatus.DNS_CONFIGURED_TPL_RESET:
          return DaemonStatusCode.tpLiteDisabled;

        case SetDNSStatus.INVALID_DNS_ADDRESS:
          return DaemonStatusCode.invalidDnsAddress;
        case SetDNSStatus.TOO_MANY_VALUES:
          return DaemonStatusCode.tooManyValues;
      }
    }
    return DaemonStatusCode.success;
  }

  Future<int> setAutoConnect(bool enabled, ConnectArguments? server) async {
    SetAutoconnectRequest request = SetAutoconnectRequest(enabled: enabled);
    if (server != null) {
      final countryCode = server.country?.code ?? "";
      final cityName = server.city?.sanitizedName ?? "";
      request = SetAutoconnectRequest(
        serverGroup: server.specialtyGroup?.backendName ?? "",
        serverTag: "$countryCode $cityName".trim(),
        enabled: true,
      );
    }
    final result = await _client.setAutoConnect(request);
    return result.type.toInt();
  }

  Future<int> setRouting(bool value) async {
    final result = await _client.setRouting(SetGenericRequest(enabled: value));
    return result.type.toInt();
  }

  Future<int> resetToDefaults() async {
    final result = await _client.setDefaults(
      SetDefaultsRequest(noLogout: true),
    );
    return result.type.toInt();
  }

  Future<int> setPostQuantum(bool value) async {
    final result = await _client.setPostQuantum(
      SetGenericRequest(enabled: value),
    );

    return _checkSettingsUpdate(result);
  }

  Future<int> useVirtualServers(bool value) async {
    final result = await _client.setVirtualLocation(
      SetGenericRequest(enabled: value),
    );
    return result.type.toInt();
  }

  int _checkSettingsUpdate(Payload response) {
    int status = response.type.toInt();
    if (status == DaemonStatusCode.success) {
      final vpnActive = response.data.isNotEmpty
          ? response.data.first.toLowerCase()
          : null;

      // If VPN is active we want to inform the user
      // That he needs to reconnect to VPN
      if (vpnActive == 'true') {
        status = DaemonStatusCode.vpnIsRunning;
      }
    }
    return status;
  }

  int _codeToDaemonStatusCode(SetErrorCode code) {
    switch (code) {
      case SetErrorCode.CONFIG_ERROR:
        return DaemonStatusCode.configError;
      case SetErrorCode.FAILURE:
        return DaemonStatusCode.failure;
      case SetErrorCode.ALREADY_SET:
        break;
    }

    return DaemonStatusCode.success;
  }
}

@Riverpod(keepAlive: true)
VpnSettingsRepository vpnSettings(Ref ref) {
  return VpnSettingsRepository();
}
