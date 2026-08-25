import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/i18n/country_names_service.dart';

void main() {
  late CountryNamesService service;

  setUp(() {
    service = CountryNamesService();
  });

  group('register', () {
    test('maps both the code and the name to the same country', () {
      final country = service.register(code: 'PL', name: 'Poland');

      expect(service.country('PL'), same(country));
      expect(service.country('Poland'), same(country));
      expect(country.code, 'PL');
      expect(country.name, 'Poland');
    });
  });

  group('country', () {
    test('returns the registered country by code', () {
      final registered = service.register(code: 'US', name: 'United States');

      expect(service.country('US'), same(registered));
    });

    test('returns the registered country by name', () {
      final registered = service.register(code: 'US', name: 'United States');

      expect(service.country('United States'), same(registered));
    });

    test('throws when the code or name was never registered', () {
      expect(() => service.country('XX'), throwsA(isA<AssertionError>()));
    });
  });

  group('localizedName', () {
    test('returns the translated name for a known country code', () {
      final localized = service.localizedName(
        Country(code: 'DE', name: 'ignored'),
      );

      expect(localized, 'Germany');
    });

    test(
      'falls back to the country name when no translation exists for the code',
      () {
        final localized = service.localizedName(
          Country(code: 'XX', name: 'Xetaland'),
        );

        expect(localized, 'Xetaland');
      },
    );
  });
}
