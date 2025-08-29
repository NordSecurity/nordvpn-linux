import 'dart:async';

import 'package:fixnum/fixnum.dart';
import 'package:nordvpn/data/mocks/daemon/cancelable_delayed.dart';
import 'package:nordvpn/data/mocks/daemon/connect_arguments_extension.dart';
import 'package:nordvpn/data/mocks/daemon/mock_servers_list.dart';
import 'package:nordvpn/data/models/allow_list.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/pb/daemon/common.pb.dart';
import 'package:nordvpn/pb/daemon/config/analytics_consent.pbenum.dart';
import 'package:nordvpn/pb/daemon/config/protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';
import 'package:nordvpn/pb/daemon/connect.pb.dart';
import 'package:nordvpn/pb/daemon/ping.pb.dart';
import 'package:nordvpn/pb/daemon/set.pb.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:nordvpn/pb/daemon/state.pb.dart';

// Store information about the application settings for the mocked daemon
final class MockApplicationSettings extends CancelableDelayed {
  final StreamController<AppState> stream;
  final MockServersList serversList;
  MockApplicationSettings(this.stream, this.serversList) {
    setDefaults();
  }

  static var delayDuration = Duration(milliseconds: 500);

  var _settings = SettingsResponse();
  String? error;
  int? errorCode;
  SetErrorCode? errorLanDiscovery;
  SetErrorCode? errorSetProtocol;
  SetErrorCode? errorTpLite;
  SetErrorCode? errorDns;

  SettingsResponse get settings => _settings;
  void replaceSettings(Settings value){
    setSettings(
      killSwitch: value.hasKillSwitch() ? value.killSwitch : null,
      protocol: value.hasProtocol() ? value.protocol : null,
      technology: value.hasTechnology() ? value.technology : null,
      obfuscate: value.hasObfuscate() ? value.obfuscate : null,
    );
  }

  Future<Payload> setSettings({
    ConsentMode? analyticsConsent,
    bool? firewall,
    int? fwmark,
    Technology? technology,
    Protocol? protocol,
    bool? virtualLocation,
    bool? killSwitch,
    bool? obfuscate,
    bool? lanDiscovery,
    bool? postquantumVpn,
    bool? routing,
    bool? threatProtectionLite,
    bool? notify,
    bool? tray,
    Allowlist? allowList,
    List<String>? dns,
    AutoconnectData? autoConnectData,
  }) async {
    await delayed(delayDuration);
    if (error != null) {
      throw error!;
    }

    if (errorCode != null) {
      return Payload(type: Int64(errorCode!));
    }

    final val = _settings.data;
    final s = Settings(
      analyticsConsent: analyticsConsent ?? val.analyticsConsent,
      firewall: firewall ?? val.firewall,
      fwmark: fwmark ?? val.fwmark,
      technology: technology ?? val.technology,
      protocol: protocol ?? val.protocol,
      virtualLocation: virtualLocation ?? val.virtualLocation,
      killSwitch: killSwitch ?? val.killSwitch,
      obfuscate: obfuscate ?? val.obfuscate,
      lanDiscovery: lanDiscovery ?? val.lanDiscovery,
      postquantumVpn: postquantumVpn ?? val.postquantumVpn,
      routing: routing ?? val.routing,
      threatProtectionLite: threatProtectionLite ?? val.threatProtectionLite,
      allowlist: allowList ?? val.allowlist,
      dns: dns ?? val.dns,
      autoConnectData: autoConnectData ?? val.autoConnectData,
      userSettings: UserSpecificSettings(
        notify: notify ?? val.userSettings.notify,
        tray: tray ?? val.userSettings.tray,
      ),
    );
    _settings = SettingsResponse(data: s);

    stream.add(AppState(settingsChange: s));
    return Payload(type: Int64(DaemonStatusCode.success));
  }

  PingResponse pingResponse = PingResponse(
    type: Int64(DaemonStatusCode.success),
    major: Int64(1),
    minor: Int64(2),
    patch: Int64(3),
    metadata: "NordVPN fake daemon",
  );

  Future<Payload> setDefaults() async {
    return await setSettings(
      analyticsConsent: ConsentMode.UNDEFINED,
      firewall: true,
      fwmark: 0xAB12,
      technology: Technology.NORDLYNX,
      protocol: Protocol.UDP,
      virtualLocation: true,
      killSwitch: false,
      lanDiscovery: false,
      notify: false,
      obfuscate: false,
      postquantumVpn: false,
      routing: false,
      threatProtectionLite: false,
      tray: true,
      allowList: Allowlist(),
      autoConnectData: AutoconnectData(),
      dns: [],
    );
  }

  Future<SetLANDiscoveryResponse> setLanDiscovery(bool enabled) async {
    await delayed(delayDuration);
    if (error != null) {
      throw error!;
    }

    if (errorLanDiscovery != null) {
      return SetLANDiscoveryResponse(errorCode: errorLanDiscovery!);
    }

    Allowlist allowlist = _settings.data.allowlist;
    var hasLan = false;

    if (enabled && allowlist.subnets.isNotEmpty) {
      allowlist.subnets.removeWhere((element) {
        final ret = isIpInLAN(element);
        hasLan = hasLan || ret;
        return ret;
      });
    }

    final res = await setSettings(allowList: allowlist, lanDiscovery: enabled);
    if (hasLan) {
      return SetLANDiscoveryResponse(
        setLanDiscoveryStatus:
            SetLANDiscoveryStatus.DISCOVERY_CONFIGURED_ALLOWLIST_RESET,
      );
    }
    return res.type == Int64(DaemonStatusCode.success)
        ? SetLANDiscoveryResponse(
            setLanDiscoveryStatus: SetLANDiscoveryStatus.DISCOVERY_CONFIGURED,
          )
        : SetLANDiscoveryResponse(errorCode: SetErrorCode.FAILURE);
  }

  Future<Payload> changeAllowList(SetAllowlistRequest request, bool add) async {
    if (error != null) {
      throw error!;
    }

    if (errorCode != null) {
      return Payload(type: Int64(errorCode!));
    }

    final subnets = List<String>.from(_settings.data.allowlist.subnets);
    final portsTcp = List<Int64>.from(_settings.data.allowlist.ports.tcp);
    final portsUdp = List<Int64>.from(_settings.data.allowlist.ports.udp);

    if (request.hasSetAllowlistSubnetRequest()) {
      if (add &&
          isIpInLAN(request.setAllowlistSubnetRequest.subnet) &&
          settings.data.lanDiscovery) {
        return Payload(type: Int64(DaemonStatusCode.privateSubnetLANDiscovery));
      }
      if (subnets.contains(request.setAllowlistSubnetRequest.subnet) == add) {
        return Payload(type: Int64(DaemonStatusCode.allowlistSubnetNoop));
      }
      if (add) {
        subnets.add(request.setAllowlistSubnetRequest.subnet);
      } else {
        subnets.remove(request.setAllowlistSubnetRequest.subnet);
      }
    }

    if (request.hasSetAllowlistPortsRequest()) {
      final ports = request.setAllowlistPortsRequest;
      bool changed = false;
      for (
        Int64 port = ports.portRange.startPort;
        port <= ports.portRange.endPort;
        port += 1
      ) {
        if ((port < 1) || (port > 65535)) {
          return Payload(type: Int64(DaemonStatusCode.allowlistPortOutOfRange));
        }
        if (ports.isTcp) {
          if (portsTcp.contains(port) != add) {
            if (add) {
              portsTcp.add(port);
            } else {
              portsTcp.remove(port);
            }
            changed = true;
          }
        }
        if (ports.isUdp) {
          if (portsUdp.contains(port) != add) {
            if (add) {
              portsUdp.add(port);
            } else {
              portsUdp.remove(port);
            }
            changed = true;
          }
        }
      }

      if (!changed) {
        return Payload(type: Int64(DaemonStatusCode.allowlistSubnetNoop));
      }
    }

    return await setSettings(
      allowList: Allowlist(
        ports: Ports(udp: portsUdp, tcp: portsTcp),
        subnets: subnets,
      ),
    );
  }

  Future<SetProtocolResponse> setProtocol(SetProtocolRequest request) async {
    await delayed(delayDuration);
    if (error != null) {
      throw error!;
    }

    if (errorSetProtocol != null) {
      return SetProtocolResponse(errorCode: errorSetProtocol!);
    }

    final res = await setSettings(protocol: request.protocol);
    if (res.type.toInt() != DaemonStatusCode.success) {
      return SetProtocolResponse(errorCode: SetErrorCode.FAILURE);
    }

    return SetProtocolResponse(
      setProtocolStatus: SetProtocolStatus.PROTOCOL_CONFIGURED,
    );
  }

  Future<SetThreatProtectionLiteResponse> setThreatProtectionLite(
    SetThreatProtectionLiteRequest request,
  ) async {
    await delayed(delayDuration);
    if (error != null) {
      throw error!;
    }

    if (errorTpLite != null) {
      return SetThreatProtectionLiteResponse(errorCode: errorTpLite!);
    }

    bool replaceDns =
        _settings.data.dns.isNotEmpty && request.threatProtectionLite;
    final res = await setSettings(
      threatProtectionLite: request.threatProtectionLite,
      dns: [],
    );

    if (res.type.toInt() != DaemonStatusCode.success) {
      return SetThreatProtectionLiteResponse(errorCode: SetErrorCode.FAILURE);
    }

    if (replaceDns) {
      return SetThreatProtectionLiteResponse(
        setThreatProtectionLiteStatus:
            SetThreatProtectionLiteStatus.TPL_CONFIGURED_DNS_RESET,
      );
    }

    return SetThreatProtectionLiteResponse(
      setThreatProtectionLiteStatus:
          SetThreatProtectionLiteStatus.TPL_CONFIGURED,
    );
  }

  Future<SetDNSResponse> setDNS(SetDNSRequest request) async {
    await delayed(delayDuration);
    if (error != null) {
      throw error!;
    }

    if (errorDns != null) {
      return SetDNSResponse(errorCode: errorDns!);
    }

    if (request.dns.length > 3) {
      return SetDNSResponse(setDnsStatus: SetDNSStatus.TOO_MANY_VALUES);
    }

    final hasTpLite = _settings.data.threatProtectionLite;

    final res = await setSettings(
      threatProtectionLite: false,
      dns: request.dns,
    );
    if (res.type.toInt() != DaemonStatusCode.success) {
      return SetDNSResponse(errorCode: SetErrorCode.FAILURE);
    }

    if (hasTpLite) {
      return SetDNSResponse(
        setDnsStatus: SetDNSStatus.DNS_CONFIGURED_TPL_RESET,
      );
    }
    return SetDNSResponse(setDnsStatus: SetDNSStatus.DNS_CONFIGURED);
  }

  Future<Payload> setAutoConnect(SetAutoconnectRequest request) async {
    if (!request.enabled) {
      return await setSettings(
        autoConnectData: AutoconnectData(enabled: false),
      );
    }

    if (!request.hasServerGroup() && !request.hasServerTag()) {
      return await setSettings(autoConnectData: AutoconnectData(enabled: true));
    }

    final params = ConnectRequest(
      serverTag: request.serverTag,
      serverGroup: request.serverGroup,
    );

    final server = serversList.findServer(params);

    if (server == null) {
      throw "server not found";
    }

    return await setSettings(
      autoConnectData: AutoconnectData(
        enabled: true,
        serverGroup: request.serverGroup.isEmpty
            ? null
            : params.toServerGroup(),
        country: request.serverTag.length <= 2 ? null : server.countryCode,
        city: request.serverTag.isEmpty ? null : server.cityName,
      ),
    );
  }
}

bool isIpInLAN(String ip) {
  final subnet = Subnet.fromString(ip);

  if (subnet.ip == null) {
    return false;
  }

  final firstOctet = subnet.ip! >> 24;
  final secondOctet = (subnet.ip! >> 16) & 0xFF;

  if ((firstOctet == 10) ||
      (firstOctet == 172 && secondOctet >= 16 && secondOctet <= 31) ||
      (firstOctet == 192 && secondOctet == 168)) {
    return true;
  }

  return false;
}
