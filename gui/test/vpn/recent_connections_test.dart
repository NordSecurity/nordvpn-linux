import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/recent_connections.dart';
import 'package:nordvpn/i18n/country_names_service.dart';
import 'package:nordvpn/pb/daemon/config/group.pb.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';
import 'package:nordvpn/pb/daemon/server_selection_rule.pbenum.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/vpn/recent_connections_item_factory.dart';
import 'package:nordvpn/pb/daemon/recent_connections.pb.dart';

void main() {
  setUpAll(() {
    final service = CountryNamesService();
    service.register(code: "DE", name: "Germany");
    sl.registerSingleton(service);
  });
  group('buildTitleParts output', () {
    final List<({RecentConnection model, TitleParts expected})> tests = [
      (
        model: RecentConnection.fromPb(
          RecentConnectionModel(countryCode: "DE"),
        ),
        expected: (primary: "Germany", secondary: "Fastest"),
      ),
      (
        model: RecentConnection.fromPb(
          RecentConnectionModel(
            countryCode: "DE",
            city: "Berlin",
            connectionType: ServerSelectionRule.CITY,
          ),
        ),
        expected: (primary: "Germany", secondary: "Berlin"),
      ),
      (
        model: RecentConnection.fromPb(
          RecentConnectionModel(
            countryCode: "DE",
            city: "Berlin",
            specificServer: "de1234",
            connectionType: ServerSelectionRule.SPECIFIC_SERVER,
          ),
        ),
        expected: (primary: "Germany", secondary: "Fastest"),
      ),
      (
        model: RecentConnection.fromPb(
          RecentConnectionModel(
            countryCode: "DE",
            city: "Berlin",
            specificServer: "de1234",
            specificServerName: "Berlin #1234",
            connectionType: ServerSelectionRule.SPECIFIC_SERVER,
          ),
        ),
        expected: (primary: "Germany", secondary: "#1234"),
      ),
      (
        model: RecentConnection.fromPb(
          RecentConnectionModel(
            countryCode: "DE",
            country: "Germany",
            group: ServerGroup.DOUBLE_VPN,
          ),
        ),
        expected: (primary: "Double VPN", secondary: "Germany - Fastest"),
      ),
      (
        model: RecentConnection.fromPb(
          RecentConnectionModel(
            countryCode: "DE",
            city: "Berlin",
            group: ServerGroup.ONION_OVER_VPN,
          ),
        ),
        expected: (primary: "Onion over VPN", secondary: "Germany - Berlin"),
      ),
      (
        model: RecentConnection.fromPb(
          RecentConnectionModel(
            countryCode: "DE",
            city: "Berlin",
            group: ServerGroup.OBFUSCATED,
          ),
        ),
        expected: (
          primary: "Obfuscated Servers",
          secondary: "Germany - Berlin",
        ),
      ),
      (
        model: RecentConnection.fromPb(
          RecentConnectionModel(
            countryCode: "DE",
            group: ServerGroup.OBFUSCATED,
          ),
        ),
        expected: (
          primary: "Obfuscated Servers",
          secondary: "Germany - Fastest",
        ),
      ),
      (
        model: RecentConnection.fromPb(
          RecentConnectionModel(
            countryCode: "DE",
            city: "Berlin",
            connectionTech: Technology.NORDWHISPER,
          ),
        ),
        expected: (
          primary: "Obfuscated Servers",
          secondary: "Germany - Berlin",
        ),
      ),
    ];

    for (final tc in tests) {
      test(tc.expected, () async {
        final parts = RecentConnectionsItemFactory.buildTitleParts(
          tc.model,
          RecentConnectionsItemFactory.isASpecialtyServer(tc.model),
        );
        expect(parts.primary, tc.expected.primary);
        expect(parts.secondary, tc.expected.secondary);
      });
    }
  });
}
