import 'package:freezed_annotation/freezed_annotation.dart';
import 'package:nordvpn/i18n/string_translation_extension.dart';

part 'city.freezed.dart';

@freezed
abstract class City with _$City {
  const City._();

  @Assert('name.isNotEmpty', 'City name cannot be empty')
  factory City(String name) = _City;

  String get localizedName => translationKey.tr(name);

  String get sanitizedName => name.toLowerCase().replaceAll(" ", "_");

  @visibleForTesting
  String get translationKey => "cities.$name";

  @override
  String toString() => name;
}
