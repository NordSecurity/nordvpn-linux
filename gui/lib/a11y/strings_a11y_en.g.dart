///
/// Generated file. Do not edit.
///
// coverage:ignore-file
// ignore_for_file: type=lint, unused_import
// dart format off

part of 'strings_a11y.g.dart';

// Path: <root>
typedef TranslationsA11yEn = TranslationsA11y; // ignore: unused_element
class TranslationsA11y with BaseTranslations<AppLocale, TranslationsA11y> {
	/// Returns the current translations of the given [context].
	///
	/// Usage:
	/// final a11y = TranslationsA11y.of(context);
	static TranslationsA11y of(BuildContext context) => InheritedLocaleData.of<AppLocale, TranslationsA11y>(context).translations;

	/// You can call this constructor and build your own translation instance of this locale.
	/// Constructing via the enum [AppLocale.build] is preferred.
	TranslationsA11y({Map<String, Node>? overrides, PluralResolver? cardinalResolver, PluralResolver? ordinalResolver, TranslationMetadata<AppLocale, TranslationsA11y>? meta})
		: assert(overrides == null, 'Set "translation_overrides: true" in order to enable this feature.'),
		  $meta = meta ?? TranslationMetadata(
		    locale: AppLocale.en,
		    overrides: overrides ?? {},
		    cardinalResolver: cardinalResolver,
		    ordinalResolver: ordinalResolver,
		  ) {
		$meta.setFlatMapFunction(_flatMapFunction);
	}

	/// Metadata for the translations of <en>.
	@override final TranslationMetadata<AppLocale, TranslationsA11y> $meta;

	/// Access flat map
	dynamic operator[](String key) => $meta.getTranslation(key);

	late final TranslationsA11y _root = this; // ignore: unused_field

	TranslationsA11y $copyWith({TranslationMetadata<AppLocale, TranslationsA11y>? meta}) => TranslationsA11y(meta: meta ?? this.$meta);

	// Translations
	late final TranslationsA11yUiEn ui = TranslationsA11yUiEn._(_root);
}

// Path: ui
class TranslationsA11yUiEn {
	TranslationsA11yUiEn._(this._root);

	final TranslationsA11y _root; // ignore: unused_field

	// Translations

	/// en: 'VPN connection resumes in $minutes minutes $seconds seconds'
	String VPNResumesIn({required Object minutes, required Object seconds}) => 'VPN connection resumes in ${minutes} minutes ${seconds} seconds';

	/// en: 'VPN connection resumes in $hours hours $minutes minutes $seconds seconds'
	String VPNResumesInWithHours({required Object hours, required Object minutes, required Object seconds}) => 'VPN connection resumes in ${hours} hours ${minutes} minutes ${seconds} seconds';
}

/// The flat map containing all translations for locale <en>.
/// Only for edge cases! For simple maps, use the map function of this library.
///
/// The Dart AOT compiler has issues with very large switch statements,
/// so the map is split into smaller functions (512 entries each).
extension on TranslationsA11y {
	dynamic _flatMapFunction(String path) {
		return switch (path) {
			'ui.VPNResumesIn' => ({required Object minutes, required Object seconds}) => 'VPN connection resumes in ${minutes} minutes ${seconds} seconds',
			'ui.VPNResumesInWithHours' => ({required Object hours, required Object minutes, required Object seconds}) => 'VPN connection resumes in ${hours} hours ${minutes} minutes ${seconds} seconds',
			_ => null,
		};
	}
}
