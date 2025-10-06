import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/i18n/country_names_service.dart';
import 'package:nordvpn/pb/daemon/connect.pb.dart';
import 'package:nordvpn/service_locator.dart';

void main() {
  setUpAll(() {
    final service = CountryNamesService();
    service.register(code: "DE", name: "Germany");
    sl.registerSingleton(service);
  });
  group('Construct ConnectRequest from ConnectArguments Tests', () {
    test('Convert from server', () async {
      final server = ServerInfo(
        id: 9999,
        hostname: "de1234.nordvpn.com",
        isVirtual: true,
      );

      expect(
        ConnectArguments(server: server).toConnectRequest(),
        ConnectRequest(serverTag: "de1234"),
      );
    });

    test('Convert from country code', () async {
      expect(
        ConnectArguments(country: Country.fromCode("DE")).toConnectRequest(),
        ConnectRequest(serverTag: "de"),
      );
    });

    test('Convert from country code and city name', () async {
      expect(
        ConnectArguments(
          country: Country.fromCode("DE"),
          city: City("Berlin"),
        ).toConnectRequest(),
        ConnectRequest(serverTag: "de berlin"),
      );
    });

    test('Convert for specialty group', () async {
      expect(
        ConnectArguments(
          specialtyGroup: ServerType.doubleVpn,
        ).toConnectRequest(),
        ConnectRequest(serverGroup: "Double_vpn"),
      );
    });

    test('Convert for specialty group and country', () async {
      expect(
        ConnectArguments(
          country: Country.fromCode("DE"),
          city: City("Berlin"),
          specialtyGroup: ServerType.p2p,
        ).toConnectRequest(),
        ConnectRequest(serverGroup: "p2p", serverTag: "de berlin"),
      );
    });
  });
}
