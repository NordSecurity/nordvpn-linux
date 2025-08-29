import 'dart:convert';

import 'package:collection/collection.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/servers_list.dart';
import 'package:nordvpn/i18n/country_names_service.dart';
import 'package:nordvpn/i18n/string_translation_extension.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:nordvpn/service_locator.dart';

late http.Response response;

void main() async {
  setUpAll(() async {
    // fetch all the countries from the API and register them into the service
    response = await http.get(countriesApiUrl);

    if (response.statusCode != 200) {
      throw Exception('Failed to fetch test data');
    }

    final service = CountryNamesService();
    sl.registerSingleton(service);

    for (final item in json.decode(response.body)) {
      final country = item as Map<String, dynamic>;
      final code = (country["code"] as String).toUpperCase();
      final name = (country["name"] as String);

      service.register(code: code, name: name);
    }
  });

  test(
    'All cities, country codes and names from API exist in translation files',
    () async {
      List<dynamic> jsonResponse = json.decode(response.body);
      for (final item in jsonResponse) {
        final countryItem = item as Map<String, dynamic>;
        final countryName = countryItem["name"] as String;
        final countryCode = (countryItem["code"] as String).toUpperCase();

        final country = Country.fromCode(countryCode);

        expect(country.code, countryCode);
        expect(country.name, countryName);
        // default is EN so it will work
        expect(country.localizedName, countryName);
        expect(Country.fromCode(countryCode.toLowerCase()).code, countryCode);
        expect(Country.fromCode(countryCode.toLowerCase()).name, countryName);

        final cities = countryItem["cities"] as List<dynamic>;
        for (var cityJson in cities) {
          cityJson = cityJson as Map<String, dynamic>;

          final city = City(cityJson["name"] as String);
          expect(
            city.translationKey.tr(""),
            isNotEmpty,
            reason:
                "missing translation for ${city.name} -> ${city.translationKey}",
          );
        }
      }
    },
  );

  test('Countries and cities are sorted', () async {
    Map<ServerType, List<CountryServersGroup>> groups = {
      ServerType.standardVpn: [
        CountryServersGroup(
          country: Country.fromCode("US"),
          cities: [
            CityServersGroup(cityName: "West new Coral", servers: []),
            CityServersGroup(cityName: "South", servers: []),
            CityServersGroup(cityName: "Time", servers: []),
          ],
          isVirtual: false,
        ),
        CountryServersGroup(
          country: Country.fromCode("DE"),
          cities: [
            CityServersGroup(cityName: "Z", servers: []),
            CityServersGroup(cityName: "A", servers: []),
            CityServersGroup(cityName: "S", servers: []),
          ],
          isVirtual: false,
        ),
        CountryServersGroup(
          country: Country.fromCode("LT"),
          cities: [
            CityServersGroup(cityName: "Z", servers: []),
            CityServersGroup(cityName: "A", servers: []),
            CityServersGroup(cityName: "S", servers: []),
          ],
          isVirtual: false,
        ),
      ],
    };
    ServersList serversList = ServersList(groups);
    serversList.groups.forEach((_, group) {
      expect(
        group.isSorted((a, b) => a.countryName.compareTo(b.countryName)),
        true,
      );

      for (final country in group) {
        expect(
          country.cities.isSorted(
            (a, b) => a.localizedName.compareTo(b.localizedName),
          ),
          true,
        );
      }
    });
  });

  test('Cities are sorted when are added manually into the country', () async {
    final cities = [
      CityServersGroup(cityName: "West", servers: []),
      CityServersGroup(cityName: "South", servers: []),
      CityServersGroup(cityName: "East", servers: []),
      CityServersGroup(cityName: "North", servers: []),
    ];

    Map<ServerType, List<CountryServersGroup>> groups = {
      ServerType.standardVpn: [
        CountryServersGroup(
          country: Country.fromCode("US"),
          cities: [],
          isVirtual: false,
        ),
        CountryServersGroup(
          country: Country.fromCode("DE"),
          cities: [],
          isVirtual: false,
        ),
      ],
    };

    groups.forEach((_, countries) {
      for (final country in countries) {
        country.cities.addAll(cities);
      }
    });

    ServersList serversList = ServersList(groups);
    serversList.groups.forEach((_, group) {
      expect(
        group.isSorted((a, b) => a.countryName.compareTo(b.countryName)),
        true,
        reason: "countries are not sorted",
      );

      for (final country in group) {
        expect(
          country.cities.isSorted(
            (a, b) => a.localizedName.compareTo(b.localizedName),
          ),
          true,
          reason: "Cities are not sorted ${country.cities}",
        );
      }
    });
  });
}
