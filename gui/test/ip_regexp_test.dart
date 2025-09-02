import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/internal/pair.dart';
import 'package:nordvpn/settings/allow_list/allow_list_helpers.dart';

void main() {
  test('RegExp works for custom DNS', () async {
    final ips = <Pair<String, bool>>[
      Pair("1.1.1.1", true),
      Pair("192.168.0.1", true),
      Pair("100.100.100.100", true),
      Pair("10.1.100.1", true),
      Pair("0.0.0.0", true),
      // failures
      Pair("256.0.0.0", false),
      Pair("0.256.0.0", false),
      Pair("0.0.256.0", false),
      Pair("0.0.0.256", false),
      Pair("01.0.0.0", false),
      Pair("0.01.0.0", false),
      Pair("0.0.01.0", false),
      Pair("2606:4700:4700::1123", false),
    ];
    for (final pair in ips) {
      expect(
        ipv4Regex.hasMatch(pair.first),
        pair.second,
        reason: "fails for ${pair.first}",
      );
    }
  });

  test('RegExp is working for allowlist', () async {
    final ips = <Pair<String, bool>>[
      Pair("1.1.1.1/32", true),
      Pair("192.168.0.1/24", true),
      Pair("100.100.100.100/15", true),
      Pair("100.100.100.100/8", true),
      Pair("10.1.100.1/4", true),
      Pair("0.0.0.0/0", true),

      // failures
      Pair("256.0.0.0/32", false),
      Pair("0.256.0.0/32", false),
      Pair("0.0.256.0/32", false),
      Pair("0.0.0.256/32", false),
      Pair("01.0.0.0/32", false),
      Pair("0.01.0.0/32", false),
      Pair("0.0.01.0/32", false),
      Pair("1.1.1.1/33", false),
      Pair("1.1.1.1/01", false),
      Pair("2606:4700:4700::1123/63", false),
    ];
    for (final pair in ips) {
      expect(
        subnetFormatPattern.hasMatch(pair.first),
        pair.second,
        reason: "fails for ${pair.first}",
      );

      if (pair.second) {
        expect(subnetFormatPattern.firstMatch(pair.first)!.groupCount, 5);
      }

      expect(
        isSubnetValid(pair.first),
        pair.second,
        reason: "fails for ${pair.first}",
      );
    }
  });
}
