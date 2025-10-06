import 'dart:collection';

import 'package:fixnum/fixnum.dart';
import 'package:freezed_annotation/freezed_annotation.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/common.pb.dart';

part 'allow_list.freezed.dart';

@freezed
abstract class AllowList with _$AllowList {
  const AllowList._();

  const factory AllowList({
    required List<Subnet> subnets,
    required List<PortInterval> ports,
  }) = _AllowList;

  factory AllowList.fromSettings(Allowlist allowList) {
    return AllowList(
      subnets: _convertFromSettings(allowList.subnets),
      ports: _mergePortRanges(
        udp: allowList.ports.udp.map((e) => e.toInt()).toSet(),
        tcp: allowList.ports.tcp.map((e) => e.toInt()).toSet(),
      ),
    );
  }

  AllowList remove({PortInterval? port, Subnet? subnet}) {
    assert((port != null) || (subnet != null), " port or subnet must be valid");

    List<PortInterval>? changePortsList;
    if (port != null) {
      changePortsList = [];
      for (var p in ports) {
        if ((p.start == port.start) && (p.end == port.end)) {
          if (p.type == port.type) {
            continue;
          }
          assert(
            (p.type == PortType.both) && (port.type != PortType.both),
            "removed port range must have both protocols",
          );

          p = p.copyWith(
            type: port.type == PortType.tcp ? PortType.udp : PortType.tcp,
          );
        }

        changePortsList.add(p);
      }
    }

    return AllowList(
      ports: changePortsList ?? ports,
      subnets: (subnet != null)
          ? subnets.where((e) => e != subnet).toList()
          : subnets,
    );
  }

  AllowList add({PortInterval? port, Subnet? subnet}) {
    assert((port != null) || (subnet != null), " port or subnet must be valid");
    List<PortInterval>? changedPortsList;
    if (port != null) {
      final tcp = <int>{};
      final udp = <int>{};
      for (final p in [...ports, port]) {
        final portsList = [for (var i = p.start; i <= p.end; i += 1) i];
        switch (p.type) {
          case PortType.both:
            tcp.addAll(portsList);
            udp.addAll(portsList);
          case PortType.tcp:
            tcp.addAll(portsList);
          case PortType.udp:
            udp.addAll(portsList);
        }
      }
      changedPortsList = _mergePortRanges(udp: udp, tcp: tcp);
    }

    List<Subnet>? changedSubnets;
    if (subnet != null) {
      changedSubnets = [];
      bool found = false;
      for (final s in subnets) {
        if (!found && subnet.contains(s)) {
          changedSubnets.add(subnet);
          found = true;
        } else {
          found = found || s.contains(subnet);
          changedSubnets.add(s);
        }
      }

      if (!found) {
        changedSubnets.add(subnet);
      }
    }

    return AllowList(
      ports: changedPortsList ?? ports,
      subnets: changedSubnets ?? subnets,
    );
  }

  bool get isEmpty => subnets.isEmpty && ports.isEmpty;
  bool get isNotEmpty => !isEmpty;
  bool get hasPrivateSubnets =>
      subnets.isNotEmpty && subnets.any((s) => s.isPrivate());

  Allowlist toSettings() {
    return Allowlist(
      subnets: subnets.isEmpty ? null : subnets.map((e) => e.value),
      ports: ports.isEmpty
          ? null
          : Ports(udp: toPorts(PortType.udp), tcp: toPorts(PortType.tcp)),
    );
  }

  List<Int64>? toPorts(PortType type) {
    final result = <Int64>[];
    for (final portRange in ports) {
      if ((portRange.type == type) || (portRange.type == PortType.both)) {
        for (var v = portRange.start; v <= portRange.end; ++v) {
          result.add(Int64(v));
        }
      }
    }
    return result.isEmpty ? null : result;
  }
}

enum PortType { both, tcp, udp }

@freezed
abstract class PortInterval with _$PortInterval {
  const PortInterval._();

  const factory PortInterval({
    required int start,
    required int end,
    required PortType type,
  }) = _PortInterval;

  bool contains(PortInterval other) {
    return ((other.type == type) || (type == PortType.both)) &&
        (other.start >= start) &&
        (other.end <= end);
  }

  bool get isRange => (start < end);

  bool get isTcp => (type == PortType.tcp) || (type == PortType.both);
  bool get isUdp => (type == PortType.udp) || (type == PortType.both);
}

// Combine ports received from the daemon to be displayed as port ranges
List<PortInterval> _mergePortRanges({
  required Set<int> udp,
  required Set<int> tcp,
}) {
  if (tcp.isEmpty && udp.isEmpty) {
    return [];
  }

  // Merge and sort unique ports
  final allPorts = {...tcp, ...udp}.toList()..sort();

  List<PortInterval> ranges = [];
  int start = -1;
  int end = -1;
  PortType type = PortType.both;

  for (final port in allPorts) {
    assert(port >= 1 && port <= maxInt16);
    if (port < 1 || port > maxInt16) {
      logger.e("invalid port $port");
      continue;
    }

    final currentType = _getPortType(port, tcp: tcp, udp: udp);
    if (port == end + 1 && currentType == type) {
      // Extend the current range
      end = port;
    } else {
      // Save previous range
      if (start != -1) {
        ranges.add(PortInterval(start: start, end: end, type: type));
      }
      // Start new range
      start = end = port;
      type = currentType;
    }
  }

  // Add last range
  if (start != -1) {
    ranges.add(PortInterval(start: start, end: end, type: type));
  }

  return UnmodifiableListView(ranges);
}

PortType _getPortType(
  int port, {
  required Set<int> tcp,
  required Set<int> udp,
}) {
  bool hasTcp = tcp.contains(port);
  bool hasUdp = udp.contains(port);
  if (hasTcp && hasUdp) return PortType.both;
  if (hasTcp) {
    return PortType.tcp;
  } else {
    assert(hasUdp);
    return PortType.udp;
  }
}

@freezed
abstract class Subnet with _$Subnet {
  const Subnet._();

  const factory Subnet({
    // string value for the subnet 0.0.0.0/32
    required String value,
    // int representation of the IP. It is null when value fails to be parsed
    required int? ip,
    // the number of bits used to describe the address /32
    required int? cidr,
  }) = _Subnet;

  factory Subnet.fromString(String subnet) {
    final match = subnetFormatPattern.firstMatch(subnet);
    if ((match == null) || (match.groupCount != 5)) {
      return Subnet(value: subnet, ip: null, cidr: null);
    }

    // Extract IP parts and CIDR
    List<int> ipParts = List.generate(4, (i) => int.parse(match.group(i + 1)!));
    int cidr = int.parse(match.group(5)!);

    // Check if each IP part is in valid range (0-255)
    if (ipParts.any((part) => part < 0 || part > 255)) {
      return Subnet(value: subnet, ip: null, cidr: null);
    }

    // Validate CIDR range (0-32)
    if (cidr < 0 || cidr > 32) {
      return Subnet(value: subnet, ip: null, cidr: null);
    }

    return Subnet(
      value: subnet,
      ip:
          (ipParts[0] << 24) |
          (ipParts[1] << 16) |
          (ipParts[2] << 8) |
          ipParts[3],
      cidr: cidr,
    );
  }

  // Calculate subnet mask
  int get mask => (0xFFFFFFFF << (32 - cidr!)) & 0xFFFFFFFF;

  // calculate the network address using the IP and mask
  int get networkAddress => ip! & mask;

  // Check if current subnet includes the provided parameter
  bool contains(Subnet other) {
    if ((ip == null) ||
        (other.ip == null) ||
        (cidr == null) ||
        (other.cidr == null)) {
      // if the parse failed compare by the string
      return other.value == value;
    }

    if (other.cidr! < cidr!) {
      return false;
    }
    return (other.ip! & mask) == networkAddress;
  }

  bool isPrivate() {
    if (ip == null) return false;

    int a = (ip! >> 24) & 0xFF;
    int b = (ip! >> 16) & 0xFF;

    return a == 10 ||
        (a == 172 && (b >= 16 && b <= 31)) ||
        (a == 192 && b == 168);
  }
}

// Convert from a string of subnets to a list of Subnet objects
List<Subnet> _convertFromSettings(List<String> subnets) {
  final result = <Subnet>[];
  for (final subnet in subnets) {
    try {
      result.add(Subnet.fromString(subnet));
    } catch (e) {
      logger.e("failed to convert subnet $subnet");
    }
  }
  return UnmodifiableListView(result);
}
