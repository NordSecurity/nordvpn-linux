import 'package:freezed_annotation/freezed_annotation.dart';
import 'package:nordvpn/i18n/country_names_service.dart';
import 'package:nordvpn/service_locator.dart';

part 'country.freezed.dart';

@freezed
abstract class Country with _$Country {
  const Country._();

  const factory Country._internal({
    required String code,
    required String name,
  }) = _Country;

  factory Country({required String code, required String name}) {
    final sanitizedCode = code.toUpperCase();
    assert(sanitizedCode.length == 2, "country code $code incorrect");
    return Country._internal(code: sanitizedCode, name: name);
  }

  factory Country.fromCode(String code) {
    return sl<CountryNamesService>().country(
      code.length == 2 ? code.toUpperCase() : code,
    );
  }

  String get localizedName => sl<CountryNamesService>().localizedName(this);

  @override
  String toString() => "$code - $name - $localizedName";
}
