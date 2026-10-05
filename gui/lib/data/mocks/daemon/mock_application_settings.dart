import 'dart:async';

import 'package:fixnum/fixnum.dart';
import 'package:nordvpn/data/mocks/daemon/cancelable_delayed.dart';
import 'package:nordvpn/data/mocks/daemon/connect_arguments_extension.dart';
import 'package:nordvpn/data/mocks/daemon/mock_servers_list.dart';
import 'package:nordvpn/data/mocks/daemon/mock_vpn_status.dart';
import 'package:nordvpn/data/models/allow_list.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/pb/daemon/common.pb.dart';
import 'package:nordvpn/pb/daemon/config/analytics_consent.pbenum.dart';
import 'package:nordvpn/pb/daemon/config/vpn_protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/connect.pb.dart';
import 'package:nordvpn/pb/daemon/ping.pb.dart';
import 'package:nordvpn/pb/daemon/set.pb.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:nordvpn/pb/daemon/state.pb.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';

// Store information about the application settings for the mocked daemon
final class MockApplicationSettings extends CancelableDelayed {
  final StreamController<AppState> stream;
  final MockServersList serversList;
  final MockVpnStatus vpnStatus;
  MockApplicationSettings(this.stream, this.serversList, this.vpnStatus) {
    setDefaults();
  }

  static var delayDuration = Duration(milliseconds: 500);

  var _settings = SettingsResponse();
  String? error;
  int? errorCode;
  SetErrorCode? errorLanDiscovery;
  SetErrorCode? errorRealTimeProtection;
  SetErrorCode? errorDns;

  List<VPNProtocol> availableVpnProtocols = [
    VPNProtocol.VPN_PROTOCOL_NORDLYNX,
    VPNProtocol.VPN_PROTOCOL_OPENVPN_UDP,
    VPNProtocol.VPN_PROTOCOL_OPENVPN_TCP,
    VPNProtocol.VPN_PROTOCOL_NORDWHISPER,
  ];

  SettingsResponse get settings => _settings;
  Settings get currentSettings => settings.data;

  void replaceSettings(Settings value) {
    setSettings(
      killSwitch: value.hasKillSwitch() ? value.killSwitch : null,
      vpnProtocol: value.hasVpnProtocol() ? value.vpnProtocol : null,
    );
  }

  Future<Payload> setSettings({
    ConsentMode? analyticsConsent,
    bool? firewall,
    int? fwmark,
    VPNProtocol? vpnProtocol,
    bool? killSwitch,
    bool? lanDiscovery,
    bool? postquantumVpn,
    bool? routing,
    bool? realTimeProtection,
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
      vpnProtocol: vpnProtocol ?? val.vpnProtocol,
      killSwitch: killSwitch ?? val.killSwitch,
      lanDiscovery: lanDiscovery ?? val.lanDiscovery,
      postquantumVpn: postquantumVpn ?? val.postquantumVpn,
      routing: routing ?? val.routing,
      realTimeProtection: realTimeProtection ?? val.realTimeProtection,
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

    // Check if VPN is active to return appropriate status
    final isVpnActive =
        vpnStatus.status.state == ConnectionState.CONNECTED ||
        vpnStatus.status.state == ConnectionState.CONNECTING;

    return Payload(
      type: Int64(DaemonStatusCode.success),
      data: isVpnActive ? ['true'] : [],
    );
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
      vpnProtocol: VPNProtocol.VPN_PROTOCOL_NORDLYNX,
      killSwitch: false,
      lanDiscovery: false,
      notify: false,
      postquantumVpn: false,
      routing: false,
      realTimeProtection: false,
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
        final newSubnet = Subnet.fromString(
          request.setAllowlistSubnetRequest.subnet,
        );
        final hasNarrower = subnets.any(
          (s) => newSubnet.contains(Subnet.fromString(s)),
        );
        if (hasNarrower && !request.setAllowlistSubnetRequest.force) {
          return Payload(
            type: Int64(DaemonStatusCode.allowlistSubnetWiderConfirm),
          );
        }
        subnets.removeWhere((s) => newSubnet.contains(Subnet.fromString(s)));
        subnets.add(request.setAllowlistSubnetRequest.subnet);
        if (newSubnet.cidr != null && newSubnet.cidr! <= 8) {
          await setSettings(
            allowList: Allowlist(
              ports: Ports(udp: portsUdp, tcp: portsTcp),
              subnets: subnets,
            ),
          );
          return Payload(
            type: Int64(DaemonStatusCode.allowlistSubnetTooWideWarn),
          );
        }
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

  Future<Payload> setVpnProtocol(VPNProtocol vpnProtocol) async {
    await delayed(delayDuration);
    if (error != null) {
      throw error!;
    }

    if (!availableVpnProtocols.contains(vpnProtocol)) {
      return Payload(type: Int64(DaemonStatusCode.featureHidden));
    }

    if (currentSettings.vpnProtocol == vpnProtocol) {
      return Payload(type: Int64(DaemonStatusCode.nothingToDo));
    }

    final res = await setSettings(vpnProtocol: vpnProtocol);
    if (res.type.toInt() != DaemonStatusCode.success) {
      return res;
    }

    final isVpnActive =
        vpnStatus.status.state == ConnectionState.CONNECTED ||
        vpnStatus.status.state == ConnectionState.CONNECTING;

    return Payload(
      type: Int64(
        isVpnActive
            ? DaemonStatusCode.successReconnectRequired
            : DaemonStatusCode.success,
      ),
    );
  }

  Future<SetRealTimeProtectionResponse> setRealTimeProtection(
    SetRealTimeProtectionRequest request,
  ) async {
    await delayed(delayDuration);
    if (error != null) {
      throw error!;
    }

    if (errorRealTimeProtection != null) {
      return SetRealTimeProtectionResponse(errorCode: errorRealTimeProtection!);
    }

    bool replaceDns =
        _settings.data.dns.isNotEmpty && request.realTimeProtection;
    final res = await setSettings(
      realTimeProtection: request.realTimeProtection,
      dns: [],
    );

    if (res.type.toInt() != DaemonStatusCode.success) {
      return SetRealTimeProtectionResponse(errorCode: SetErrorCode.FAILURE);
    }

    if (replaceDns) {
      return SetRealTimeProtectionResponse(
        setRealTimeProtectionStatus:
            SetRealTimeProtectionStatus.RTP_CONFIGURED_DNS_RESET,
      );
    }

    return SetRealTimeProtectionResponse(
      setRealTimeProtectionStatus: SetRealTimeProtectionStatus.RTP_CONFIGURED,
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

    final hasRealTimeProtection = _settings.data.realTimeProtection;

    final res = await setSettings(realTimeProtection: false, dns: request.dns);
    if (res.type.toInt() != DaemonStatusCode.success) {
      return SetDNSResponse(errorCode: SetErrorCode.FAILURE);
    }

    if (hasRealTimeProtection) {
      return SetDNSResponse(
        setDnsStatus: SetDNSStatus.DNS_CONFIGURED_RTP_RESET,
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
        countryCode: request.serverTag.isEmpty ? null : server.countryCode,
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
