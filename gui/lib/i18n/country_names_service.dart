import 'package:flutter/foundation.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/i18n/string_translation_extension.dart';
import 'package:nordvpn/logger.dart';

// CountryNamesService will be responsible with the translation of country names
// and mapping country name(english) to country code and from country code to
// country name(in user language).
// The class will fetch the list of countries from the API and check if all have
// translations. Countries that don't have translations are stored into a map.
final class CountryNamesService {
  // map the country code and the country name to a Country object
  final Map<String, Country> _countries = {};

  Country register({required String code, required String name}) {
    if (name.isEmpty) {
      logger.i("no country name was provided");
      name = translateCountryName(code);
    }
    final country = Country(code: code, name: name);
    _countries[code] = country;
    _countries[name] = country;
    return country;
  }

  // Mapping from country code to localized country name
  String localizedName(Country country) {
    return translateCountryName(country.code, country.name);
  }

  // Construct the key used to find the country code into the string files
  @visibleForTesting
  String translateCountryName(String countryCode, [String? defaultValue]) {
    assert(
      countryCode.toUpperCase() == countryCode,
      "country code is not upper case: $countryCode",
    );
    return "countries.$countryCode".tr(defaultValue ?? countryCode);
  }

  // Map country name in english or country code to a country object
  Country country(String code) {
    final country = _countries[code];
    assert(country != null, "country not found $code");
    if (country != null) {
      return country;
    }

    return Country(code: code, name: code);
  }
}
