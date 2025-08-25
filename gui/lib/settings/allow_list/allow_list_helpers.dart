import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/models/allow_list.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';

bool isValidPortNumber(String value) {
  final port = int.tryParse(value);
  if (port == null) {
    return false;
  }

  return (port > 0) && (port <= maxInt16);
}

bool isSubnetValid(String subnet) {
  if (subnet.isEmpty) {
    return false;
  }

  final match = subnetFormatPattern.firstMatch(subnet);
  if ((match == null) || (match.groupCount != 5)) {
    logger.d("$subnet is not a valid subnet");
    return false;
  }

  // Extract IP parts and CIDR
  List<int> ipParts = List.generate(4, (i) => int.parse(match.group(i + 1)!));
  int cidr = int.parse(match.group(5)!);

  // Check if each IP part is in valid range (0-255)
  if (ipParts.any((part) => part < 0 || part > 255)) {
    logger.e("$subnet contains invalid IP $ipParts");
    return false;
  }

  // Validate CIDR range (0-32)
  if (cidr < 0 || cidr > 32) {
    logger.e("$subnet has invalid CIDR $cidr");
    return false;
  }

  return true;
}

String portTypeToString(PortType port) {
  switch (port) {
    case PortType.both:
      return t.ui.all;
    case PortType.tcp:
      return "TCP";
    case PortType.udp:
      return "UDP";
  }
}
