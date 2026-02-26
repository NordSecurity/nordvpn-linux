///
/// Generated file. Do not edit.
///
// coverage:ignore-file
// ignore_for_file: type=lint, unused_import
// dart format off

part of 'strings.g.dart';

// Path: <root>
typedef TranslationsEn = Translations; // ignore: unused_element
class Translations with BaseTranslations<AppLocale, Translations> {
	/// Returns the current translations of the given [context].
	///
	/// Usage:
	/// final t = Translations.of(context);
	static Translations of(BuildContext context) => InheritedLocaleData.of<AppLocale, Translations>(context).translations;

	/// You can call this constructor and build your own translation instance of this locale.
	/// Constructing via the enum [AppLocale.build] is preferred.
	Translations({Map<String, Node>? overrides, PluralResolver? cardinalResolver, PluralResolver? ordinalResolver, TranslationMetadata<AppLocale, Translations>? meta})
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
	@override final TranslationMetadata<AppLocale, Translations> $meta;

	/// Access flat map
	dynamic operator[](String key) => $meta.getTranslation(key);

	late final Translations _root = this; // ignore: unused_field

	Translations $copyWith({TranslationMetadata<AppLocale, Translations>? meta}) => Translations(meta: meta ?? this.$meta);

	// Translations
	late final TranslationsCitiesEn cities = TranslationsCitiesEn._(_root);
	late final TranslationsCountriesEn countries = TranslationsCountriesEn._(_root);
	late final TranslationsDaemonEn daemon = TranslationsDaemonEn._(_root);
	late final TranslationsUiEn ui = TranslationsUiEn._(_root);
}

// Path: cities
class TranslationsCitiesEn {
	TranslationsCitiesEn._(this._root);

	final Translations _root; // ignore: unused_field

	// Translations

	/// en: 'Tirana'
	String get tirana => 'Tirana';

	/// en: 'Algiers'
	String get algiers => 'Algiers';

	/// en: 'Addis Ababa'
	String get addis_ababa => 'Addis Ababa';

	/// en: 'Andorra la Vella'
	String get andorra_la_vella => 'Andorra la Vella';

	/// en: 'Buenos Aires'
	String get buenos_aires => 'Buenos Aires';

	/// en: 'Yerevan'
	String get yerevan => 'Yerevan';

	/// en: 'Adelaide'
	String get adelaide => 'Adelaide';

	/// en: 'Brisbane'
	String get brisbane => 'Brisbane';

	/// en: 'Melbourne'
	String get melbourne => 'Melbourne';

	/// en: 'Perth'
	String get perth => 'Perth';

	/// en: 'Sydney'
	String get sydney => 'Sydney';

	/// en: 'Vienna'
	String get vienna => 'Vienna';

	/// en: 'Baku'
	String get baku => 'Baku';

	/// en: 'Nassau'
	String get nassau => 'Nassau';

	/// en: 'Dhaka'
	String get dhaka => 'Dhaka';

	/// en: 'Brussels'
	String get brussels => 'Brussels';

	/// en: 'Belmopan'
	String get belmopan => 'Belmopan';

	/// en: 'Hamilton'
	String get hamilton => 'Hamilton';

	/// en: 'Thimphu'
	String get thimphu => 'Thimphu';

	/// en: 'La Paz'
	String get la_paz => 'La Paz';

	/// en: 'Novi Travnik'
	String get novi_travnik => 'Novi Travnik';

	/// en: 'Sao Paulo'
	String get sao_paulo => 'Sao Paulo';

	/// en: 'Bandar Seri Begawan'
	String get bandar_seri_begawan => 'Bandar Seri Begawan';

	/// en: 'Sofia'
	String get sofia => 'Sofia';

	/// en: 'Phnom Penh'
	String get phnom_penh => 'Phnom Penh';

	/// en: 'Montreal'
	String get montreal => 'Montreal';

	/// en: 'Toronto'
	String get toronto => 'Toronto';

	/// en: 'Vancouver'
	String get vancouver => 'Vancouver';

	/// en: 'George Town'
	String get george_town => 'George Town';

	/// en: 'Santiago'
	String get santiago => 'Santiago';

	/// en: 'Bogota'
	String get bogota => 'Bogota';

	/// en: 'San Jose'
	String get san_jose => 'San Jose';

	/// en: 'Zagreb'
	String get zagreb => 'Zagreb';

	/// en: 'Nicosia'
	String get nicosia => 'Nicosia';

	/// en: 'Prague'
	String get prague => 'Prague';

	/// en: 'Copenhagen'
	String get copenhagen => 'Copenhagen';

	/// en: 'Santo Domingo'
	String get santo_domingo => 'Santo Domingo';

	/// en: 'Quito'
	String get quito => 'Quito';

	/// en: 'Cairo'
	String get cairo => 'Cairo';

	/// en: 'San Salvador'
	String get san_salvador => 'San Salvador';

	/// en: 'Tallinn'
	String get tallinn => 'Tallinn';

	/// en: 'Helsinki'
	String get helsinki => 'Helsinki';

	/// en: 'Marseille'
	String get marseille => 'Marseille';

	/// en: 'Paris'
	String get paris => 'Paris';

	/// en: 'Tbilisi'
	String get tbilisi => 'Tbilisi';

	/// en: 'Berlin'
	String get berlin => 'Berlin';

	/// en: 'Frankfurt'
	String get frankfurt => 'Frankfurt';

	/// en: 'Hamburg'
	String get hamburg => 'Hamburg';

	/// en: 'Accra'
	String get accra => 'Accra';

	/// en: 'Athens'
	String get athens => 'Athens';

	/// en: 'Nuuk'
	String get nuuk => 'Nuuk';

	/// en: 'Hagatna'
	String get hagatna => 'Hagatna';

	/// en: 'Guatemala City'
	String get guatemala_city => 'Guatemala City';

	/// en: 'Tegucigalpa'
	String get tegucigalpa => 'Tegucigalpa';

	/// en: 'Hong Kong'
	String get hong_kong => 'Hong Kong';

	/// en: 'Budapest'
	String get budapest => 'Budapest';

	/// en: 'Reykjavik'
	String get reykjavik => 'Reykjavik';

	/// en: 'Mumbai'
	String get mumbai => 'Mumbai';

	/// en: 'Jakarta'
	String get jakarta => 'Jakarta';

	/// en: 'Dublin'
	String get dublin => 'Dublin';

	/// en: 'Douglas'
	String get douglas => 'Douglas';

	/// en: 'Tel Aviv'
	String get tel_aviv => 'Tel Aviv';

	/// en: 'Milan'
	String get milan => 'Milan';

	/// en: 'Palermo'
	String get palermo => 'Palermo';

	/// en: 'Rome'
	String get rome => 'Rome';

	/// en: 'Kingston'
	String get kingston => 'Kingston';

	/// en: 'Osaka'
	String get osaka => 'Osaka';

	/// en: 'Tokyo'
	String get tokyo => 'Tokyo';

	/// en: 'Saint Helier'
	String get saint_helier => 'Saint Helier';

	/// en: 'Astana'
	String get astana => 'Astana';

	/// en: 'Nairobi'
	String get nairobi => 'Nairobi';

	/// en: 'Vientiane'
	String get vientiane => 'Vientiane';

	/// en: 'Riga'
	String get riga => 'Riga';

	/// en: 'Beirut'
	String get beirut => 'Beirut';

	/// en: 'Vaduz'
	String get vaduz => 'Vaduz';

	/// en: 'Vilnius'
	String get vilnius => 'Vilnius';

	/// en: 'Luxembourg'
	String get luxembourg => 'Luxembourg';

	/// en: 'Kuala Lumpur'
	String get kuala_lumpur => 'Kuala Lumpur';

	/// en: 'Valletta'
	String get valletta => 'Valletta';

	/// en: 'Mexico'
	String get mexico => 'Mexico';

	/// en: 'Chisinau'
	String get chisinau => 'Chisinau';

	/// en: 'Monte Carlo'
	String get monte_carlo => 'Monte Carlo';

	/// en: 'Ulaanbaatar'
	String get ulaanbaatar => 'Ulaanbaatar';

	/// en: 'Podgorica'
	String get podgorica => 'Podgorica';

	/// en: 'Rabat'
	String get rabat => 'Rabat';

	/// en: 'Naypyidaw'
	String get naypyidaw => 'Naypyidaw';

	/// en: 'Kathmandu'
	String get kathmandu => 'Kathmandu';

	/// en: 'Amsterdam'
	String get amsterdam => 'Amsterdam';

	/// en: 'Auckland'
	String get auckland => 'Auckland';

	/// en: 'Lagos'
	String get lagos => 'Lagos';

	/// en: 'Skopje'
	String get skopje => 'Skopje';

	/// en: 'Oslo'
	String get oslo => 'Oslo';

	/// en: 'Karachi'
	String get karachi => 'Karachi';

	/// en: 'Panama City'
	String get panama_city => 'Panama City';

	/// en: 'Port Moresby'
	String get port_moresby => 'Port Moresby';

	/// en: 'Asuncion'
	String get asuncion => 'Asuncion';

	/// en: 'Lima'
	String get lima => 'Lima';

	/// en: 'Manila'
	String get manila => 'Manila';

	/// en: 'Warsaw'
	String get warsaw => 'Warsaw';

	/// en: 'Lisbon'
	String get lisbon => 'Lisbon';

	/// en: 'San Juan'
	String get san_juan => 'San Juan';

	/// en: 'Bucharest'
	String get bucharest => 'Bucharest';

	/// en: 'Belgrade'
	String get belgrade => 'Belgrade';

	/// en: 'Singapore'
	String get singapore => 'Singapore';

	/// en: 'Bratislava'
	String get bratislava => 'Bratislava';

	/// en: 'Ljubljana'
	String get ljubljana => 'Ljubljana';

	/// en: 'Johannesburg'
	String get johannesburg => 'Johannesburg';

	/// en: 'Seoul'
	String get seoul => 'Seoul';

	/// en: 'Barcelona'
	String get barcelona => 'Barcelona';

	/// en: 'Madrid'
	String get madrid => 'Madrid';

	/// en: 'Colombo'
	String get colombo => 'Colombo';

	/// en: 'Stockholm'
	String get stockholm => 'Stockholm';

	/// en: 'Zurich'
	String get zurich => 'Zurich';

	/// en: 'Taipei'
	String get taipei => 'Taipei';

	/// en: 'Bangkok'
	String get bangkok => 'Bangkok';

	/// en: 'Port of Spain'
	String get port_of_spain => 'Port of Spain';

	/// en: 'Istanbul'
	String get istanbul => 'Istanbul';

	/// en: 'Kyiv'
	String get kyiv => 'Kyiv';

	/// en: 'Dubai'
	String get dubai => 'Dubai';

	/// en: 'Edinburgh'
	String get edinburgh => 'Edinburgh';

	/// en: 'Glasgow'
	String get glasgow => 'Glasgow';

	/// en: 'London'
	String get london => 'London';

	/// en: 'Manchester'
	String get manchester => 'Manchester';

	/// en: 'Atlanta'
	String get atlanta => 'Atlanta';

	/// en: 'Buffalo'
	String get buffalo => 'Buffalo';

	/// en: 'Charlotte'
	String get charlotte => 'Charlotte';

	/// en: 'Chicago'
	String get chicago => 'Chicago';

	/// en: 'Dallas'
	String get dallas => 'Dallas';

	/// en: 'Denver'
	String get denver => 'Denver';

	/// en: 'Detroit'
	String get detroit => 'Detroit';

	/// en: 'Kansas City'
	String get kansas_city => 'Kansas City';

	/// en: 'Los Angeles'
	String get los_angeles => 'Los Angeles';

	/// en: 'Manassas'
	String get manassas => 'Manassas';

	/// en: 'Miami'
	String get miami => 'Miami';

	/// en: 'New York'
	String get new_york => 'New York';

	/// en: 'Phoenix'
	String get phoenix => 'Phoenix';

	/// en: 'Pittsburgh'
	String get pittsburgh => 'Pittsburgh';

	/// en: 'Saint Louis'
	String get saint_louis => 'Saint Louis';

	/// en: 'Salt Lake City'
	String get salt_lake_city => 'Salt Lake City';

	/// en: 'San Francisco'
	String get san_francisco => 'San Francisco';

	/// en: 'Seattle'
	String get seattle => 'Seattle';

	/// en: 'Montevideo'
	String get montevideo => 'Montevideo';

	/// en: 'Tashkent'
	String get tashkent => 'Tashkent';

	/// en: 'Caracas'
	String get caracas => 'Caracas';

	/// en: 'Hanoi'
	String get hanoi => 'Hanoi';

	/// en: 'Ho Chi Minh City'
	String get ho_chi_minh_city => 'Ho Chi Minh City';

	/// en: 'Houston'
	String get houston => 'Houston';

	/// en: 'McAllen'
	String get mcallen => 'McAllen';

	/// en: 'Luanda'
	String get luanda => 'Luanda';

	/// en: 'Manama'
	String get manama => 'Manama';

	/// en: 'Amman'
	String get amman => 'Amman';

	/// en: 'Kuwait City'
	String get kuwait_city => 'Kuwait City';

	/// en: 'Maputo'
	String get maputo => 'Maputo';

	/// en: 'Dakar'
	String get dakar => 'Dakar';

	/// en: 'Tunis'
	String get tunis => 'Tunis';

	/// en: 'Boston'
	String get boston => 'Boston';

	/// en: 'Strasbourg'
	String get strasbourg => 'Strasbourg';

	/// en: 'Omaha'
	String get omaha => 'Omaha';

	/// en: 'Moroni'
	String get moroni => 'Moroni';

	/// en: 'Baghdad'
	String get baghdad => 'Baghdad';

	/// en: 'Tripoli'
	String get tripoli => 'Tripoli';

	/// en: 'Doha'
	String get doha => 'Doha';

	/// en: 'Kigali'
	String get kigali => 'Kigali';

	/// en: 'Nashville'
	String get nashville => 'Nashville';

	/// en: 'Kabul'
	String get kabul => 'Kabul';

	/// en: 'Mogadishu'
	String get mogadishu => 'Mogadishu';

	/// en: 'Nouakchott'
	String get nouakchott => 'Nouakchott';

	/// en: 'Ashburn'
	String get ashburn => 'Ashburn';

	/// en: 'Port Louis'
	String get port_louis => 'Port Louis';

	/// en: 'Dushanbe'
	String get dushanbe => 'Dushanbe';

	/// en: 'Lewiston'
	String get lewiston => 'Lewiston';
}

// Path: countries
class TranslationsCountriesEn {
	TranslationsCountriesEn._(this._root);

	final Translations _root; // ignore: unused_field

	// Translations

	/// en: 'Albania'
	String get AL => 'Albania';

	/// en: 'Algeria'
	String get DZ => 'Algeria';

	/// en: 'Andorra'
	String get AD => 'Andorra';

	/// en: 'Angola'
	String get AO => 'Angola';

	/// en: 'Argentina'
	String get AR => 'Argentina';

	/// en: 'Armenia'
	String get AM => 'Armenia';

	/// en: 'Australia'
	String get AU => 'Australia';

	/// en: 'Austria'
	String get AT => 'Austria';

	/// en: 'Azerbaijan'
	String get AZ => 'Azerbaijan';

	/// en: 'Bahamas'
	String get BS => 'Bahamas';

	/// en: 'Bahrain'
	String get BH => 'Bahrain';

	/// en: 'Bangladesh'
	String get BD => 'Bangladesh';

	/// en: 'Belgium'
	String get BE => 'Belgium';

	/// en: 'Belize'
	String get BZ => 'Belize';

	/// en: 'Bermuda'
	String get BM => 'Bermuda';

	/// en: 'Bhutan'
	String get BT => 'Bhutan';

	/// en: 'Bolivia'
	String get BO => 'Bolivia';

	/// en: 'Bosnia and Herzegovina'
	String get BA => 'Bosnia and Herzegovina';

	/// en: 'Brazil'
	String get BR => 'Brazil';

	/// en: 'Brunei Darussalam'
	String get BN => 'Brunei Darussalam';

	/// en: 'Bulgaria'
	String get BG => 'Bulgaria';

	/// en: 'Cambodia'
	String get KH => 'Cambodia';

	/// en: 'Canada'
	String get CA => 'Canada';

	/// en: 'Cayman Islands'
	String get KY => 'Cayman Islands';

	/// en: 'Chile'
	String get CL => 'Chile';

	/// en: 'Colombia'
	String get CO => 'Colombia';

	/// en: 'Costa Rica'
	String get CR => 'Costa Rica';

	/// en: 'Croatia'
	String get HR => 'Croatia';

	/// en: 'Cyprus'
	String get CY => 'Cyprus';

	/// en: 'Czech Republic'
	String get CZ => 'Czech Republic';

	/// en: 'Denmark'
	String get DK => 'Denmark';

	/// en: 'Dominican Republic'
	String get DO => 'Dominican Republic';

	/// en: 'Ecuador'
	String get EC => 'Ecuador';

	/// en: 'Egypt'
	String get EG => 'Egypt';

	/// en: 'El Salvador'
	String get SV => 'El Salvador';

	/// en: 'Estonia'
	String get EE => 'Estonia';

	/// en: 'Finland'
	String get FI => 'Finland';

	/// en: 'France'
	String get FR => 'France';

	/// en: 'Georgia'
	String get GE => 'Georgia';

	/// en: 'Germany'
	String get DE => 'Germany';

	/// en: 'Ghana'
	String get GH => 'Ghana';

	/// en: 'Greece'
	String get GR => 'Greece';

	/// en: 'Greenland'
	String get GL => 'Greenland';

	/// en: 'Guam'
	String get GU => 'Guam';

	/// en: 'Guatemala'
	String get GT => 'Guatemala';

	/// en: 'Honduras'
	String get HN => 'Honduras';

	/// en: 'Hong Kong'
	String get HK => 'Hong Kong';

	/// en: 'Hungary'
	String get HU => 'Hungary';

	/// en: 'Iceland'
	String get IS => 'Iceland';

	/// en: 'India'
	String get IN => 'India';

	/// en: 'Indonesia'
	String get ID => 'Indonesia';

	/// en: 'Ireland'
	String get IE => 'Ireland';

	/// en: 'Isle of Man'
	String get IM => 'Isle of Man';

	/// en: 'Israel'
	String get IL => 'Israel';

	/// en: 'Italy'
	String get IT => 'Italy';

	/// en: 'Jamaica'
	String get JM => 'Jamaica';

	/// en: 'Japan'
	String get JP => 'Japan';

	/// en: 'Jersey'
	String get JE => 'Jersey';

	/// en: 'Jordan'
	String get JO => 'Jordan';

	/// en: 'Kazakhstan'
	String get KZ => 'Kazakhstan';

	/// en: 'Kenya'
	String get KE => 'Kenya';

	/// en: 'Kuwait'
	String get KW => 'Kuwait';

	/// en: 'Lao People's Democratic Republic'
	String get LA => 'Lao People\'s Democratic Republic';

	/// en: 'Latvia'
	String get LV => 'Latvia';

	/// en: 'Lebanon'
	String get LB => 'Lebanon';

	/// en: 'Liechtenstein'
	String get LI => 'Liechtenstein';

	/// en: 'Lithuania'
	String get LT => 'Lithuania';

	/// en: 'Luxembourg'
	String get LU => 'Luxembourg';

	/// en: 'Malaysia'
	String get MY => 'Malaysia';

	/// en: 'Malta'
	String get MT => 'Malta';

	/// en: 'Mexico'
	String get MX => 'Mexico';

	/// en: 'Moldova'
	String get MD => 'Moldova';

	/// en: 'Monaco'
	String get MC => 'Monaco';

	/// en: 'Mongolia'
	String get MN => 'Mongolia';

	/// en: 'Montenegro'
	String get ME => 'Montenegro';

	/// en: 'Morocco'
	String get MA => 'Morocco';

	/// en: 'Mozambique'
	String get MZ => 'Mozambique';

	/// en: 'Myanmar'
	String get MM => 'Myanmar';

	/// en: 'Nepal'
	String get NP => 'Nepal';

	/// en: 'Netherlands'
	String get NL => 'Netherlands';

	/// en: 'New Zealand'
	String get NZ => 'New Zealand';

	/// en: 'Nigeria'
	String get NG => 'Nigeria';

	/// en: 'North Macedonia'
	String get MK => 'North Macedonia';

	/// en: 'Norway'
	String get NO => 'Norway';

	/// en: 'Pakistan'
	String get PK => 'Pakistan';

	/// en: 'Panama'
	String get PA => 'Panama';

	/// en: 'Papua New Guinea'
	String get PG => 'Papua New Guinea';

	/// en: 'Paraguay'
	String get PY => 'Paraguay';

	/// en: 'Peru'
	String get PE => 'Peru';

	/// en: 'Philippines'
	String get PH => 'Philippines';

	/// en: 'Poland'
	String get PL => 'Poland';

	/// en: 'Portugal'
	String get PT => 'Portugal';

	/// en: 'Puerto Rico'
	String get PR => 'Puerto Rico';

	/// en: 'Romania'
	String get RO => 'Romania';

	/// en: 'Serbia'
	String get RS => 'Serbia';

	/// en: 'Senegal'
	String get SN => 'Senegal';

	/// en: 'Singapore'
	String get SG => 'Singapore';

	/// en: 'Slovakia'
	String get SK => 'Slovakia';

	/// en: 'Slovenia'
	String get SI => 'Slovenia';

	/// en: 'South Africa'
	String get ZA => 'South Africa';

	/// en: 'South Korea'
	String get KR => 'South Korea';

	/// en: 'Spain'
	String get ES => 'Spain';

	/// en: 'Sri Lanka'
	String get LK => 'Sri Lanka';

	/// en: 'Sweden'
	String get SE => 'Sweden';

	/// en: 'Switzerland'
	String get CH => 'Switzerland';

	/// en: 'Taiwan'
	String get TW => 'Taiwan';

	/// en: 'Thailand'
	String get TH => 'Thailand';

	/// en: 'Trinidad and Tobago'
	String get TT => 'Trinidad and Tobago';

	/// en: 'Turkey'
	String get TR => 'Turkey';

	/// en: 'Tunisia'
	String get TN => 'Tunisia';

	/// en: 'Ukraine'
	String get UA => 'Ukraine';

	/// en: 'United Arab Emirates'
	String get AE => 'United Arab Emirates';

	/// en: 'United Kingdom'
	String get GB => 'United Kingdom';

	/// en: 'United States'
	String get US => 'United States';

	/// en: 'Uruguay'
	String get UY => 'Uruguay';

	/// en: 'Uzbekistan'
	String get UZ => 'Uzbekistan';

	/// en: 'Venezuela'
	String get VE => 'Venezuela';

	/// en: 'Vietnam'
	String get VN => 'Vietnam';
}

// Path: daemon
class TranslationsDaemonEn {
	TranslationsDaemonEn._(this._root);

	final Translations _root; // ignore: unused_field

	// Translations

	/// en: 'Unauthorized'
	String get code_3001_title => 'Unauthorized';

	/// en: 'We couldn't log you in. Make sure your credentials are correct. If you have turned on MFA, log in using the 'nordvpn login' command.'
	String get code_3001_msg => 'We couldn\'t log you in. Make sure your credentials are correct. If you have turned on MFA, log in using the \'nordvpn login\' command.';

	/// en: 'Format error'
	String get code_3003_title => 'Format error';

	/// en: 'The command is not valid.'
	String get code_3003_msg => 'The command is not valid.';

	/// en: 'Config error'
	String get code_3004_title => 'Config error';

	/// en: 'We ran into an issue with the config file. If the problem persists, please contact our customer support.'
	String get code_3004_msg => 'We ran into an issue with the config file. If the problem persists, please contact our customer support.';

	/// en: 'Empty payload'
	String get code_3005_title => 'Empty payload';

	/// en: 'Something went wrong. Please try again. If the problem persists, contact our customer support.'
	String get code_3005_msg => 'Something went wrong. Please try again. If the problem persists, contact our customer support.';

	/// en: 'You're offline'
	String get code_3007_title => 'You\'re offline';

	/// en: 'Please check your internet connection and try again.'
	String get code_3007_msg => 'Please check your internet connection and try again.';

	/// en: 'Account expired'
	String get code_3008_title => 'Account expired';

	/// en: 'Your account has expired. Renew your subscription now to continue enjoying the ultimate privacy and security with NordVPN.'
	String get code_3008_msg => 'Your account has expired. Renew your subscription now to continue enjoying the ultimate privacy and security with NordVPN.';

	/// en: 'VPN misconfigured'
	String get code_3010_title => 'VPN misconfigured';

	/// en: 'Something went wrong. Please try again. If the problem persists, contact our customer support.'
	String get code_3010_msg => 'Something went wrong. Please try again. If the problem persists, contact our customer support.';

	/// en: 'Daemon offline'
	String get code_3013_title => 'Daemon offline';

	/// en: 'We couldn't reach System Daemon.'
	String get code_3013_msg => 'We couldn\'t reach System Daemon.';

	/// en: 'Gateway error'
	String get code_3014_title => 'Gateway error';

	/// en: 'It's not you, it's us. We're having trouble reaching our servers. If the issue persists, please contact our customer support.'
	String get code_3014_msg => 'It\'s not you, it\'s us. We\'re having trouble reaching our servers. If the issue persists, please contact our customer support.';

	/// en: 'Outdated'
	String get code_3015_title => 'Outdated';

	/// en: 'A new version of NordVPN is available! Please update the app.'
	String get code_3015_msg => 'A new version of NordVPN is available! Please update the app.';

	/// en: 'Dependency error'
	String get code_3017_title => 'Dependency error';

	/// en: 'Currently in use.'
	String get code_3017_msg => 'Currently in use.';

	/// en: 'No new data error'
	String get code_3019_title => 'No new data error';

	/// en: 'You have already provided a rating for your active/previous connection.'
	String get code_3019_msg => 'You have already provided a rating for your active/previous connection.';

	/// en: 'Expired renew token'
	String get code_3021_title => 'Expired renew token';

	/// en: 'For security purposes, please log in again.'
	String get code_3021_msg => 'For security purposes, please log in again.';

	/// en: 'Token renew error'
	String get code_3022_title => 'Token renew error';

	/// en: 'We couldn't load your account data. Check your internet connection and try again. If the issue persists, please contact our customer support.'
	String get code_3022_msg => 'We couldn\'t load your account data. Check your internet connection and try again. If the issue persists, please contact our customer support.';

	/// en: 'Kill switch error'
	String get code_3023_title => 'Kill switch error';

	/// en: 'Something went wrong. Please try again. If the problem persists, contact our customer support.'
	String get code_3023_msg => 'Something went wrong. Please try again. If the problem persists, contact our customer support.';

	/// en: 'Bad request'
	String get code_3024_title => 'Bad request';

	/// en: 'Username or password is not correct. Please try again.'
	String get code_3024_msg => 'Username or password is not correct. Please try again.';

	/// en: 'Internal daemon error'
	String get code_3026_title => 'Internal daemon error';

	/// en: 'Something went wrong. Please try again. If the problem persists, contact our customer support.'
	String get code_3026_msg => 'Something went wrong. Please try again. If the problem persists, contact our customer support.';

	/// en: 'Server not available'
	String get code_3032_title => 'Server not available';

	/// en: 'The specified server is not available at the moment or does not support your connection settings.'
	String get code_3032_msg => 'The specified server is not available at the moment or does not support your connection settings.';

	/// en: 'Inexistent tag'
	String get code_3033_title => 'Inexistent tag';

	/// en: 'The specified server does not exist.'
	String get code_3033_msg => 'The specified server does not exist.';

	/// en: 'Double group error'
	String get code_3034_title => 'Double group error';

	/// en: 'You cannot connect to a group and set the group option at the same time.'
	String get code_3034_msg => 'You cannot connect to a group and set the group option at the same time.';

	/// en: 'Token login error'
	String get code_3035_title => 'Token login error';

	/// en: 'Token parameter value is missing.'
	String get code_3035_msg => 'Token parameter value is missing.';

	/// en: 'Inexistent group'
	String get code_3036_title => 'Inexistent group';

	/// en: 'The specified group does not exist.'
	String get code_3036_msg => 'The specified group does not exist.';

	/// en: 'Not obfuscated auto connect server'
	String get code_3037_title => 'Not obfuscated auto connect server';

	/// en: 'Your selected server doesn't support obfuscation. Choose a different server or turn off obfuscation.'
	String get code_3037_msg => 'Your selected server doesn\'t support obfuscation. Choose a different server or turn off obfuscation.';

	/// en: 'Obfuscated auto connect server'
	String get code_3038_title => 'Obfuscated auto connect server';

	/// en: 'Turn on obfuscation to connect to obfuscated servers.'
	String get code_3038_msg => 'Turn on obfuscation to connect to obfuscated servers.';

	/// en: 'Invalid token'
	String get code_3039_title => 'Invalid token';

	/// en: 'We couldn't log you in - the access token is not valid. Please check if you've entered the token correctly. If the issue persists, contact our customer support.'
	String get code_3039_msg => 'We couldn\'t log you in - the access token is not valid. Please check if you\'ve entered the token correctly. If the issue persists, contact our customer support.';

	/// en: 'Private subnet LAN discovery'
	String get code_3040_title => 'Private subnet LAN discovery';

	/// en: 'Allowlisting a private subnet is not available while local network discovery is turned on.'
	String get code_3040_msg => 'Allowlisting a private subnet is not available while local network discovery is turned on.';

	/// en: 'Dedicated IP renew error'
	String get code_3041_title => 'Dedicated IP renew error';

	/// en: 'You don’t have a dedicated IP subscription. To get a personal IP address, continue in the browser.'
	String get code_3041_msg => 'You don’t have a dedicated IP subscription. To get a personal IP address, continue in the browser.';

	/// en: 'Dedicated IP no server'
	String get code_3042_title => 'Dedicated IP no server';

	/// en: 'This server isn't currently included in your dedicated IP subscription.'
	String get code_3042_msg => 'This server isn\'t currently included in your dedicated IP subscription.';

	/// en: 'Dedicated IP service but no server'
	String get code_3043_title => 'Dedicated IP service but no server';

	/// en: 'Please select the preferred server location for your dedicated IP in Nord Account.'
	String get code_3043_msg => 'Please select the preferred server location for your dedicated IP in Nord Account.';

	/// en: 'Allow list invalid subnet'
	String get code_3044_title => 'Allow list invalid subnet';

	/// en: 'The command is not valid.'
	String get code_3044_msg => 'The command is not valid.';

	/// en: 'Allow list port noop'
	String get code_3047_title => 'Allow list port noop';

	/// en: 'Port is already on the allowlist.'
	String get code_3047_msg => 'Port is already on the allowlist.';

	/// en: 'Here's what to know'
	String get code_3048_title => 'Here\'s what to know';

	/// en: 'The post-quantum VPN and Meshnet can't run at the same time. Please turn off one feature to use the other.'
	String get code_3048_msg => 'The post-quantum VPN and Meshnet can\'t run at the same time. Please turn off one feature to use the other.';

	/// en: 'Here's what to know'
	String get code_3049_title => 'Here\'s what to know';

	/// en: 'This setting is not compatible with post-quantum encryption. To use it, turn off post-quantum encryption first.'
	String get code_3049_msg => 'This setting is not compatible with post-quantum encryption. To use it, turn off post-quantum encryption first.';

	/// en: 'Disabled technology'
	String get code_3051_title => 'Disabled technology';

	/// en: 'Unable to connect with the current technology. Please try a different one using the command: nordvpn set technology.'
	String get code_3051_msg => 'Unable to connect with the current technology. Please try a different one using the command: nordvpn set technology.';

	/// en: 'Server access not allowed'
	String get code_3057_title => 'Server access not allowed';

	/// en: 'To connect to the selected server, turn on virtual location access using the app’s command-line interface.'
	String get code_3057_msg => 'To connect to the selected server, turn on virtual location access using the app’s command-line interface.';

	/// en: 'Restart daemon to apply setting'
	String get code_5007_title => 'Restart daemon to apply setting';

	/// en: 'Restart the daemon to apply this setting. For example, use the command `sudo systemctl restart nordvpnd` on systemd distributions.'
	String get code_5007_msg => 'Restart the daemon to apply this setting. For example, use the command `sudo systemctl restart nordvpnd` on systemd distributions.';

	/// en: 'gRPC timeout error'
	String get code_5008_title => 'gRPC timeout error';

	/// en: 'Request time out.'
	String get code_5008_msg => 'Request time out.';

	/// en: 'Missing exchange token'
	String get code_5010_title => 'Missing exchange token';

	/// en: 'The exchange token is missing. Please try logging in again. If the issue persists, contact our customer support.'
	String get code_5010_msg => 'The exchange token is missing. Please try logging in again. If the issue persists, contact our customer support.';

	/// en: 'Failed to open the browser'
	String get code_5013_title => 'Failed to open the browser';

	/// en: 'We couldn't open the browser for you to log in.'
	String get code_5013_msg => 'We couldn\'t open the browser for you to log in.';

	/// en: 'Failed to open the browser'
	String get code_5014_title => 'Failed to open the browser';

	/// en: 'We couldn't open the browser for you to create an account.'
	String get code_5014_msg => 'We couldn\'t open the browser for you to create an account.';

	/// en: 'It didn't work this time'
	String get code_5015_title => 'It didn\'t work this time';

	/// en: 'We couldn't connect you to the VPN. Please check your internet connection and try again. If the issue persists, contact our customer support.'
	String get code_5015_msg => 'We couldn\'t connect you to the VPN. Please check your internet connection and try again. If the issue persists, contact our customer support.';

	/// en: 'It didn't work this time'
	String get genericErrorTitle => 'It didn\'t work this time';

	/// en: 'Something went wrong. Please try again. If the problem persists, contact our customer support.'
	String get genericErrorMessage => 'Something went wrong. Please try again. If the problem persists, contact our customer support.';
}

// Path: ui
class TranslationsUiEn {
	TranslationsUiEn._(this._root);

	final Translations _root; // ignore: unused_field

	// Translations

	/// en: 'Search'
	String get search => 'Search';

	/// en: 'Countries'
	String get countries => 'Countries';

	/// en: 'Specialty servers'
	String get specialServers => 'Specialty servers';

	/// en: 'Cities'
	String get cities => 'Cities';

	/// en: 'No results found.'
	String get noResults => 'No results found.';

	/// en: 'Connecting to the daemon...'
	String get waitingToConnectToDaemon => 'Connecting to the daemon...';

	/// en: 'Fetching data'
	String get fetchingData => 'Fetching data';

	/// en: 'Failed to fetch data'
	String get failedToFetchData => 'Failed to fetch data';

	/// en: 'Retry'
	String get retry => 'Retry';

	/// en: 'Quick Connect'
	String get quickConnect => 'Quick Connect';

	/// en: 'Register'
	String get register => 'Register';

	/// en: 'Sign in'
	String get signIn => 'Sign in';

	/// en: 'Disconnect'
	String get disconnect => 'Disconnect';

	/// en: 'Double VPN'
	String get double_vpn => 'Double VPN';

	/// en: 'Onion Over VPN'
	String get onion_over_vpn => 'Onion Over VPN';

	/// en: 'Fatal error'
	String get fatalErrorMessage => 'Fatal error';

	/// en: 'Connected'
	String get connected => 'Connected';

	/// en: 'Not connected'
	String get notConnected => 'Not connected';

	/// en: 'Connect now or pick a country'
	String get connectOrPickCountry => 'Connect now or pick a country';

	/// en: 'General'
	String get general => 'General';

	/// en: 'Appearance, notifications and analytics settings'
	String get generalSettingsSubtitle => 'Appearance, notifications and analytics settings';

	/// en: 'Auto-connect'
	String get autoConnect => 'Auto-connect';

	/// en: 'Auto-connected'
	String get autoConnected => 'Auto-connected';

	/// en: 'Kill Switch'
	String get killSwitch => 'Kill Switch';

	/// en: 'Account'
	String get account => 'Account';

	/// en: 'Log out, subscription'
	String get accountSubtitle => 'Log out, subscription';

	/// en: 'Apps'
	String get otherApps => 'Apps';

	/// en: 'Allowlist'
	String get allowlist => 'Allowlist';

	/// en: 'DNS'
	String get dns => 'DNS';

	/// en: 'Settings'
	String get settings => 'Settings';

	/// en: 'Launch at Startup'
	String get launchAppAtStartup => 'Launch at Startup';

	/// en: 'VPN Protocol'
	String get vpnProtocol => 'VPN Protocol';

	/// en: 'Obfuscate'
	String get obfuscate => 'Obfuscate';

	/// en: 'VPN Connection Status Notifications'
	String get notificationsStatus => 'VPN Connection Status Notifications';

	/// en: 'Firewall'
	String get firewall => 'Firewall';

	/// en: 'Allow the use of the system firewall. When enabled, you can attach a firewall mark to VPN packets for custom firewall rules.'
	String get firewallDescription => 'Allow the use of the system firewall. When enabled, you can attach a firewall mark to VPN packets for custom firewall rules.';

	/// en: 'Reset all app settings to default'
	String get resetToDefaults => 'Reset all app settings to default';

	/// en: 'Firewall mark'
	String get firewallMark => 'Firewall mark';

	/// en: 'Reset'
	String get reset => 'Reset';

	/// en: 'Confirm'
	String get confirm => 'Confirm';

	/// en: 'Cancel'
	String get cancel => 'Cancel';

	/// en: 'Disable internet access if the VPN connection drops to secure your data from accidental exposure.'
	String get killSwitchDescription => 'Disable internet access if the VPN connection drops to secure your data from accidental exposure.';

	/// en: 'Threat Protection Lite'
	String get tpLite => 'Threat Protection Lite';

	/// en: 'When you're connected to VPN, DNS filtering blocks ads and malicious domains before any threats reach your device.'
	String get tpLiteDescription => 'When you\'re connected to VPN, DNS filtering blocks ads and malicious domains before any threats reach your device.';

	/// en: 'Enabling Threat Protection Lite will result in the removal of the custom DNS configuration. Continue?'
	String get tpLiteWillDisableDns => 'Enabling Threat Protection Lite will result in the removal of the custom DNS configuration. Continue?';

	/// en: 'Enabling custom DNS configuration will result in the removal of the Threat Protection Lite. Continue?'
	String get customDnsWillDisableTpLite => 'Enabling custom DNS configuration will result in the removal of the Threat Protection Lite. Continue?';

	/// en: 'Add custom DNS'
	String get addCustomDns => 'Add custom DNS';

	/// en: 'Add port'
	String get addPort => 'Add port';

	/// en: 'Add port range'
	String get addPortRange => 'Add port range';

	/// en: 'Add subnet'
	String get addSubnet => 'Add subnet';

	/// en: 'Log in to NordVPN'
	String get loginToNordVpn => 'Log in to NordVPN';

	/// en: 'New here? Sign up for Nord Account to get started'
	String get newHereMessage => 'New here? Sign up for Nord Account to get started';

	/// en: 'Terms of Service'
	String get termsOfService => 'Terms of Service';

	/// en: 'Privacy Policy'
	String get privacyPolicy => 'Privacy Policy';

	/// en: 'Auto - renewal terms'
	String get autoRenewalTerms => 'Auto - renewal terms';

	/// en: 'Subscription'
	String get subscription => 'Subscription';

	/// en: 'Log out'
	String get logout => 'Log out';

	/// en: '(one) {Expires in $n day on $date} (other) {Expires in $n days on $date}'
	String accountExpireIn({required num n, required Object date}) => (_root.$meta.cardinalResolver ?? PluralResolvers.cardinal('en'))(n,
		one: 'Expires in ${n} day on ${date}',
		other: 'Expires in ${n} days on ${date}',
	);

	/// en: 'Connect to'
	String get connectTo => 'Connect to';

	/// en: 'Recommended server'
	String get recommendedServer => 'Recommended server';

	/// en: 'Apps'
	String get apps => 'Apps';

	/// en: 'NordVPN'
	String get nordVpn => 'NordVPN';

	/// en: 'Use NordVPN on 6 devices at the same time at no extra cost.'
	String get useNordVpnOn6Devices => 'Use NordVPN on 6 devices at the same time at no extra cost.';

	/// en: 'Explore apps and browser extensions'
	String get exploreAppsAndExtensions => 'Explore apps and browser extensions';

	/// en: 'Scan to download mobile app'
	String get scan => 'Scan to download mobile app';

	/// en: 'More apps for all-around security'
	String get moreApps => 'More apps for all-around security';

	/// en: 'NordPass'
	String get nordPass => 'NordPass';

	/// en: 'Generate, store, and organize your passwords.'
	String get nordPassDescription => 'Generate, store, and organize your passwords.';

	/// en: 'NordLocker'
	String get nordLocker => 'NordLocker';

	/// en: 'Store your files securely in our end-to-end encrypted cloud.'
	String get nordLockerDescription => 'Store your files securely in our end-to-end encrypted cloud.';

	/// en: 'NordLayer'
	String get nordLayer => 'NordLayer';

	/// en: 'Get a powerful security solution for your business network.'
	String get nordLayerDescription => 'Get a powerful security solution for your business network.';

	/// en: 'Learn more'
	String get learnMore => 'Learn more';

	/// en: 'Email Support'
	String get emailSupport => 'Email Support';

	/// en: 'Knowledge Base'
	String get knowledgeBase => 'Knowledge Base';

	/// en: 'Routing'
	String get routing => 'Routing';

	/// en: 'Connect to VPN'
	String get connectToVpn => 'Connect to VPN';

	/// en: 'Connecting'
	String get connecting => 'Connecting';

	/// en: 'Finding server...'
	String get findingServer => 'Finding server...';

	/// en: 'No results found. Try another keyword.'
	String get noResultsFound => 'No results found. Try another keyword.';

	/// en: 'Search countries, cities, or servers'
	String get searchServersHint => 'Search countries, cities, or servers';

	/// en: '$n cities available'
	String citiesAvailable({required Object n}) => '${n} cities available';

	/// en: 'Virtual'
	String get virtual => 'Virtual';

	/// en: 'Dedicated IP'
	String get dedicatedIp => 'Dedicated IP';

	/// en: 'Double VPN'
	String get doubleVpn => 'Double VPN';

	/// en: 'Onion over VPN'
	String get onionOverVpn => 'Onion over VPN';

	/// en: 'P2P'
	String get p2p => 'P2P';

	/// en: 'Europe'
	String get europe => 'Europe';

	/// en: 'The Americas'
	String get theAmericas => 'The Americas';

	/// en: 'Asia Pacific'
	String get asiaPacific => 'Asia Pacific';

	/// en: 'Africa, the Middle East, and India'
	String get africaTheMiddleEastAndIndia => 'Africa, the Middle East, and India';

	/// en: 'Obfuscated'
	String get obfuscated => 'Obfuscated';

	/// en: 'Obfuscated Servers'
	String get obfuscatedServers => 'Obfuscated Servers';

	/// en: 'Pick a location for your IP'
	String get selectServerForDip => 'Pick a location for your IP';

	/// en: 'Select location'
	String get selectLocation => 'Select location';

	/// en: 'You have successfully purchased a dedicated IP – great! To start using it, select a location for your dedicated IP from the many options that we offer.'
	String get dipSelectLocationDescription => 'You have successfully purchased a dedicated IP – great! To start using it, select a location for your dedicated IP from the many options that we offer.';

	/// en: 'Choose a location for your dedicated IP'
	String get chooseLocationForDip => 'Choose a location for your dedicated IP';

	/// en: 'Get dedicated IP'
	String get getDip => 'Get dedicated IP';

	/// en: 'Get your personal IP'
	String get getYourDip => 'Get your personal IP';

	/// en: 'Get a personal IP address that belongs only to you. Enjoy all the benefits of VPN encryption without dealing with blocklists, identity checks, and selecting images of boats in CAPTCHAs.'
	String get getDipDescription => 'Get a personal IP address that belongs only to you. Enjoy all the benefits of VPN encryption without dealing with blocklists, identity checks, and selecting images of boats in CAPTCHAs.';

	/// en: 'Notifications'
	String get notifications => 'Notifications';

	/// en: 'Search country or city'
	String get specialtyServersSearchHint => 'Search country or city';

	/// en: 'On'
	String get on => 'On';

	/// en: 'Off'
	String get off => 'Off';

	/// en: 'Invalid format'
	String get invalidFormat => 'Invalid format';

	/// en: 'Servers'
	String get servers => 'Servers';

	/// en: 'Security and privacy'
	String get securityAndPrivacy => 'Security and privacy';

	/// en: 'Allowlist, DNS, LAN discovery, obfuscation, firewall'
	String get securityAndPrivacySubtitle => 'Allowlist, DNS, LAN discovery, obfuscation, firewall';

	/// en: 'Threat Protection'
	String get threatProtection => 'Threat Protection';

	/// en: 'Blocks harmful websites, ads, and trackers'
	String get threatProtectionSubtitle => 'Blocks harmful websites, ads, and trackers';

	/// en: 'Appearance'
	String get appearance => 'Appearance';

	/// en: 'Light'
	String get light => 'Light';

	/// en: 'Dark'
	String get dark => 'Dark';

	/// en: 'Show notifications'
	String get showNotifications => 'Show notifications';

	/// en: 'VPN connection'
	String get vpnConnection => 'VPN connection';

	/// en: 'Auto-connect, Kill Switch, protocol'
	String get vpnConnectionSubtitle => 'Auto-connect, Kill Switch, protocol';

	/// en: 'Automatically connect to the fastest available server or your chosen server location when the app starts.'
	String get autoConnectDescription => 'Automatically connect to the fastest available server or your chosen server location when the app starts.';

	/// en: 'Fastest'
	String get fastestServer => 'Fastest';

	/// en: 'Change'
	String get change => 'Change';

	/// en: 'NordLynx'
	String get nordLynx => 'NordLynx';

	/// en: 'OpenVPN (TCP)'
	String get openVpnTcp => 'OpenVPN (TCP)';

	/// en: 'OpenVPN (UDP)'
	String get openVpnUdp => 'OpenVPN (UDP)';

	/// en: 'Auto-connect to'
	String get autoConnectTo => 'Auto-connect to';

	/// en: 'Standard VPN'
	String get standardVpn => 'Standard VPN';

	/// en: 'Go back'
	String get goBack => 'Go back';

	/// en: 'Done'
	String get done => 'Done';

	/// en: 'Search for country or city'
	String get searchCountryAndCity => 'Search for country or city';

	/// en: 'Reset all custom settings to default?'
	String get resetAllCustomSettings => 'Reset all custom settings to default?';

	/// en: 'This will remove your personalized configurations across the app and restore default settings.'
	String get resetSettingsAlertDescription => 'This will remove your personalized configurations across the app and restore default settings.';

	/// en: 'This will remove your personalized configurations across the app and disconnect you from the VPN.'
	String get resetAndDisconnectDesc => 'This will remove your personalized configurations across the app and disconnect you from the VPN.';

	/// en: 'Reset settings'
	String get resetSettings => 'Reset settings';

	/// en: 'Reset and disconnect'
	String get resetAndDisconnect => 'Reset and disconnect';

	/// en: 'LAN discovery'
	String get lanDiscovery => 'LAN discovery';

	/// en: 'Make your device visible to other devices on your local network while connected to the VPN. Access printers, TVs, and other LAN devices.'
	String get lanDiscoveryDescription => 'Make your device visible to other devices on your local network while connected to the VPN. Access printers, TVs, and other LAN devices.';

	/// en: 'Custom DNS'
	String get customDns => 'Custom DNS';

	/// en: 'Set custom DNS server addresses to use.'
	String get customDnsDescription => 'Set custom DNS server addresses to use.';

	/// en: 'Use custom routing rules instead of the default VPN configuration.'
	String get routingDescription => 'Use custom routing rules instead of the default VPN configuration.';

	/// en: 'Post-quantum encryption'
	String get postQuantumVpn => 'Post-quantum encryption';

	/// en: 'Activate next-generation encryption that protects your data from threats posed by quantum computing.'
	String get postQuantumDescription => 'Activate next-generation encryption that protects your data from threats posed by quantum computing.';

	/// en: 'Avoid detection by traffic sensors in restricted networks while using a VPN. When enabled, only obfuscated servers are available.'
	String get obfuscationDescription => 'Avoid detection by traffic sensors in restricted networks while using a VPN. When enabled, only obfuscated servers are available.';

	/// en: 'Obfuscation'
	String get obfuscation => 'Obfuscation';

	/// en: 'Add'
	String get add => 'Add';

	/// en: 'Custom DNS: $n/3'
	String customDnsEntries({required Object n}) => 'Custom DNS: ${n}/3';

	/// en: 'Add up to 3 DNS servers'
	String get addUpTo3DnsServers => 'Add up to 3 DNS servers';

	/// en: 'Nothing here yet'
	String get nothingHereYet => 'Nothing here yet';

	/// en: 'To activate custom DNS, add at least one DNS server.'
	String get addCustomDnsDescription => 'To activate custom DNS, add at least one DNS server.';

	/// en: 'Blocks dangerous websites and flashy ads at the domain level. Works only when you’re connected to a VPN.'
	String get threatProtectionDescription => 'Blocks dangerous websites and flashy ads at the domain level. Works only when you’re connected to a VPN.';

	/// en: 'Custom DNS will be reset'
	String get resetCustomDns => 'Custom DNS will be reset';

	/// en: 'Turning on Threat Protection will set your custom DNS settings to default. Continue anyway?'
	String get resetCustomDnsDescription => 'Turning on Threat Protection will set your custom DNS settings to default. Continue anyway?';

	/// en: 'Continue'
	String get continueWord => 'Continue';

	/// en: 'Threat Protection will be turned off'
	String get threatProtectionWillTurnOff => 'Threat Protection will be turned off';

	/// en: 'Threat Protection works only with the default DNS. Set a custom DNS server anyway?'
	String get threatProtectionWillTurnOffDescription => 'Threat Protection works only with the default DNS. Set a custom DNS server anyway?';

	/// en: 'Set custom DNS'
	String get setCustomDns => 'Set custom DNS';

	/// en: 'Turn off custom DNS?'
	String get turnOffCustomDns => 'Turn off custom DNS?';

	/// en: 'This will remove all your previously added DNS servers.'
	String get turnOffCustomDnsDescription => 'This will remove all your previously added DNS servers.';

	/// en: 'Turn off'
	String get turnOff => 'Turn off';

	/// en: 'Active until ${expirationDate: String}'
	String subscriptionValidationDate({required String expirationDate}) => 'Active until ${expirationDate}';

	/// en: 'Log in'
	String get logIn => 'Log in';

	/// en: 'Create account'
	String get createAccount => 'Create account';

	/// en: 'What is a Nord Account?'
	String get whatIsNordAccount => 'What is a Nord Account?';

	/// en: 'For troubleshooting [go to Support Center](${supportUrl: Uri})'
	String forTroubleshooting({required Uri supportUrl}) => 'For troubleshooting [go to Support Center](${supportUrl})';

	/// en: 'Copy'
	String get copy => 'Copy';

	/// en: 'Copied to clipboard'
	String get copiedToClipboard => 'Copied to clipboard';

	/// en: 'Failed to load NordVPN service'
	String get failedToLoadService => 'Failed to load NordVPN service';

	/// en: 'Try running these commands in the terminal. Then restart your device.'
	String get tryRunningTheseCommands => 'Try running these commands in the terminal. Then restart your device.';

	/// en: 'Cybersecurity built for every day'
	String get loginTitle => 'Cybersecurity built for every day';

	/// en: 'Verifying login status...'
	String get verifyingLogin => 'Verifying login status...';

	/// en: 'Your NordVPN subscription has ended'
	String get subscriptionHasEnded => 'Your NordVPN subscription has ended';

	/// en: 'But we don’t have to say goodbye! Renew your subscription for $email to continue enjoying a safer and more private internet.'
	String pleaseRenewYourSubscription({required Object email}) => 'But we don’t have to say goodbye! Renew your subscription for ${email} to continue enjoying a safer and more private internet.';

	/// en: 'Renew subscription'
	String get renewSubscription => 'Renew subscription';

	/// en: 'Your NordVPN versions are incompatible'
	String get appVersionIsIncompatible => 'Your NordVPN versions are incompatible';

	/// en: 'Please install the latest versions of this graphical interface app and the NordVPN daemon.'
	String get appVersionIsIncompatibleDescription => 'Please install the latest versions of this graphical interface app and the NordVPN daemon.';

	/// en: 'For more options, check our [compatibility guide](${compatibilityUrl: Uri})'
	String appVersionCompatibilityRecommendation({required Uri compatibilityUrl}) => 'For more options, check our [compatibility guide](${compatibilityUrl})';

	/// en: 'Kill Switch is blocking the login. Turn it off for now to continue.'
	String get turnOffKillSwitchDescription => 'Kill Switch is blocking the login. Turn it off for now to continue.';

	/// en: 'Turn off Kill Switch'
	String get turnOffKillSwitch => 'Turn off Kill Switch';

	/// en: 'Connect now'
	String get connectNow => 'Connect now';

	/// en: 'Setting auto-connect to [$target]...'
	String settingAutoconnectTo({required Object target}) => 'Setting auto-connect to [${target}]...';

	/// en: 'Encrypt your traffic twice for extra security'
	String get doubleVpnDesc => 'Encrypt your traffic twice for extra security';

	/// en: 'Use the Onion network with VPN protection'
	String get onionOverVpnDesc => 'Use the Onion network with VPN protection';

	/// en: 'Enjoy the best download speed'
	String get p2pDesc => 'Enjoy the best download speed';

	/// en: 'Save'
	String get save => 'Save';

	/// en: 'Close'
	String get close => 'Close';

	/// en: 'to'
	String get to => 'to';

	/// en: 'Use custom DNS'
	String get useCustomDns => 'Use custom DNS';

	/// en: 'Add up to three DNS servers.'
	String get useCustomDnsDescription => 'Add up to three DNS servers.';

	/// en: 'Enter DNS server address'
	String get enterDnsAddress => 'Enter DNS server address';

	/// en: 'This server is already on the list.'
	String get duplicatedDnsServer => 'This server is already on the list.';

	/// en: 'Obfuscation is turned on, so only obfuscated server locations will show up.'
	String get obfuscationSearchWarning => 'Obfuscation is turned on, so only obfuscated server locations will show up.';

	/// en: 'No results found. To access all available servers, turn off obfuscation.'
	String get obfuscationErrorNoServerFound => 'No results found. To access all available servers, turn off obfuscation.';

	/// en: 'Go to Settings'
	String get goToSettings => 'Go to Settings';

	/// en: 'Use allowlist'
	String get useAllowList => 'Use allowlist';

	/// en: 'Exclude ports, port ranges, or subnets from VPN protection.'
	String get useAllowListSettingDescription => 'Exclude ports, port ranges, or subnets from VPN protection.';

	/// en: 'Specify ports, port ranges, or subnets to exclude from VPN protection. Allowlisted ports may accept incoming connections from any external source outside your network.'
	String get useAllowListScreenDescription => 'Specify ports, port ranges, or subnets to exclude from VPN protection. Allowlisted ports may accept incoming connections from any external source outside your network.';

	/// en: 'Turn off allowlist?'
	String get turnOffAllowList => 'Turn off allowlist?';

	/// en: 'Disabling the allowlist will delete all your previously added ports, port ranges, and subnets.'
	String get turnOffAllowListDescription => 'Disabling the allowlist will delete all your previously added ports, port ranges, and subnets.';

	/// en: 'Port'
	String get port => 'Port';

	/// en: 'Port range'
	String get portRange => 'Port range';

	/// en: 'Subnet'
	String get subnet => 'Subnet';

	/// en: 'Enter port'
	String get enterPort => 'Enter port';

	/// en: 'Select protocol'
	String get selectProtocol => 'Select protocol';

	/// en: 'All'
	String get all => 'All';

	/// en: 'Protocol'
	String get protocol => 'Protocol';

	/// en: 'Port is already on the list'
	String get portAlreadyInList => 'Port is already on the list';

	/// en: 'The range is already on the list'
	String get portRangeAlreadyInList => 'The range is already on the list';

	/// en: 'Enter port range'
	String get enterPortRange => 'Enter port range';

	/// en: 'Enter subnet'
	String get enterSubnet => 'Enter subnet';

	/// en: 'Subnet is already on the list'
	String get subnetAlreadyInList => 'Subnet is already on the list';

	/// en: 'Delete'
	String get delete => 'Delete';

	/// en: 'Settings weren't saved'
	String get settingsWereNotSaved => 'Settings weren\'t saved';

	/// en: 'We couldn't save your settings to the configuration file.'
	String get couldNotSave => 'We couldn\'t save your settings to the configuration file.';

	/// en: 'Turn off obfuscation for more server types'
	String get turnOffObfuscationServerTypes => 'Turn off obfuscation for more server types';

	/// en: 'Turn off obfuscation for more locations'
	String get turnOffObfuscationLocations => 'Turn off obfuscation for more locations';

	/// en: 'NordWhisper'
	String get nordWhisper => 'NordWhisper';

	/// en: 'System'
	String get system => 'System';

	/// en: 'We'll remove private subnets from allowlist'
	String get removePrivateSubnets => 'We\'ll remove private subnets from allowlist';

	/// en: 'Enabling LAN discovery will remove any private subnets from allowlist. Continue?'
	String get removePrivateSubnetsDescription => 'Enabling LAN discovery will remove any private subnets from allowlist. Continue?';

	/// en: 'Private subnet can't be added'
	String get privateSubnetCantBeAdded => 'Private subnet can\'t be added';

	/// en: 'Allowlisting a private subnet isn’t available while local network discovery is enabled. To add a private subnet, turn off LAN discovery.'
	String get privateSubnetCantBeAddedDescription => 'Allowlisting a private subnet isn’t available while local network discovery is enabled. To add a private subnet, turn off LAN discovery.';

	/// en: 'Turn off LAN discovery'
	String get turnOffLanDiscovery => 'Turn off LAN discovery';

	/// en: 'Start port can’t be greater than end port'
	String get startPortBiggerThanEnd => 'Start port can’t be greater than end port';

	/// en: 'We couldn’t connect to NordVPN service'
	String get weCouldNotConnectToService => 'We couldn’t connect to NordVPN service';

	/// en: 'Need help? [Visit our Support Center](${supportUrl: Uri}) '
	String needHelp({required Uri supportUrl}) => 'Need help? [Visit our Support Center](${supportUrl}) ';

	/// en: 'Issue persists? [Contact our customer support](${supportUrl: Uri})'
	String issuePersists({required Uri supportUrl}) => 'Issue persists? [Contact our customer support](${supportUrl})';

	/// en: 'Meshnet'
	String get meshnet => 'Meshnet';

	/// en: 'systemd distribution'
	String get systemdDistribution => 'systemd distribution';

	/// en: 'non-systemd distribution'
	String get nonSystemdDistro => 'non-systemd distribution';

	/// en: 'The service isn’t running. To start it, use this command in the terminal.'
	String get tryRunningOneCommand => 'The service isn’t running. To start it, use this command in the terminal.';

	/// en: 'Something unexpected happened while we were trying to load your analytics consent setting.'
	String get failedToFetchConsentData => 'Something unexpected happened while we were trying to load your analytics consent setting.';

	/// en: 'Something unexpected happened while we were trying to load your account data.'
	String get failedToFetchAccountData => 'Something unexpected happened while we were trying to load your account data.';

	/// en: 'Try again'
	String get tryAgain => 'Try again';

	/// en: 'We hit an error'
	String get weHitAnError => 'We hit an error';

	/// en: 'We value your privacy'
	String get weValueYourPrivacy => 'We value your privacy';

	/// en: 'That’s why we want to be transparent about what data you agree to give us. We only collect the bare minimum of information required to offer a smooth and stable VPN experience. Your browsing activities remain private, regardless of your choice. By selecting “Accept,” you allow us to collect and use limited app performance data for analytics, as explained in our [Privacy Policy](${privacyUrl: Uri}). Select “Customize” to manage your privacy choices or learn more about each option.'
	String consentDescription({required Uri privacyUrl}) => 'That’s why we want to be transparent about what data you agree to give us. We only collect the bare minimum of information required to offer a smooth and stable VPN experience.\nYour browsing activities remain private, regardless of your choice.\n\nBy selecting “Accept,” you allow us to collect and use limited app performance data for analytics, as explained in our [Privacy Policy](${privacyUrl}).\n\nSelect “Customize” to manage your privacy choices or learn more about each option.';

	/// en: 'Accept'
	String get accept => 'Accept';

	/// en: 'Customize'
	String get customize => 'Customize';

	/// en: 'Reject non-essential'
	String get rejectNonEssential => 'Reject non-essential';

	/// en: 'Privacy preferences'
	String get privacyPreferences => 'Privacy preferences';

	/// en: 'Customize your privacy choices'
	String get privacyPreferencesDescription => 'Customize your privacy choices';

	/// en: 'Essential (Required)'
	String get essentialRequired => 'Essential (Required)';

	/// en: 'Send us the essential data we need for our app to work. This also helps enforce our [Terms of Service](${termsUrl: Uri}), prevent fraud, and maintain the security of our services. '
	String requiredAnalyticsDescription({required Uri termsUrl}) => 'Send us the essential data we need for our app to work. This also helps enforce our [Terms of Service](${termsUrl}), prevent fraud, and maintain the security of our services. ';

	/// en: 'Confirm preferences'
	String get confirmPreferences => 'Confirm preferences';

	/// en: 'Analytics'
	String get analytics => 'Analytics';

	/// en: 'Help us improve the app by sending aggregated usage data. We collect this data to keep our features relevant to your needs and to fix issues faster.'
	String get analyticsDescription => 'Help us improve the app by sending aggregated usage data. We collect this data to keep our features relevant to your needs and to fix issues faster.';

	/// en: 'Back'
	String get back => 'Back';

	/// en: 'We need some permissions to connect to NordVPN service'
	String get snapScreenTitle => 'We need some permissions to connect to NordVPN service';

	/// en: 'Grant permission by running these commands in the terminal. Then refresh the screen.'
	String get snapScreenDescription => 'Grant permission by running these commands in the terminal. Then refresh the screen.';

	/// en: 'Refresh'
	String get refresh => 'Refresh';

	/// en: 'Terms'
	String get terms => 'Terms';

	/// en: 'Learn about our legal terms.'
	String get termsSubtitle => 'Learn about our legal terms.';

	/// en: 'By continuing to use this app, you agree to our terms and how we handle your data. To read the terms and privacy policy check the links below.'
	String get termsAgreementDescription => 'By continuing to use this app, you agree to our terms and how we handle your data. To read the terms and privacy policy check the links below.';

	/// en: 'Read more'
	String get readMore => 'Read more';

	/// en: 'Using third-party DNS may limit website availability. For the best browsing experience, use our default settings.'
	String get customDnsWarning => 'Using third-party DNS may limit website availability. For the best browsing experience, use our default settings.';

	/// en: 'Account created: ${creation_date: String}'
	String accountCreatedOn({required String creation_date}) => 'Account created: ${creation_date}';

	/// en: 'Manage subscription'
	String get manageSubscription => 'Manage subscription';

	/// en: 'Change password'
	String get changePassword => 'Change password';

	/// en: 'Product Hub'
	String get productHub => 'Product Hub';

	/// en: 'Inactive'
	String get subscriptionInactive => 'Inactive';

	/// en: 'Recent connections'
	String get recentConnections => 'Recent connections';

	/// en: 'Standard VPN Servers'
	String get standardVpnServer => 'Standard VPN Servers';

	/// en: 'Reconnect now'
	String get reconnectNow => 'Reconnect now';

	/// en: 'Reconnect to change protocol'
	String get reconnectToChangeProtocol => 'Reconnect to change protocol';

	/// en: 'To apply this change, we'll reconnect you to the VPN.'
	String get reconnectToChangeProtocolDescription => 'To apply this change, we\'ll reconnect you to the VPN.';

	/// en: 'Reconnect to apply changes'
	String get reconnectToApplyChanges => 'Reconnect to apply changes';

	/// en: 'Your new settings will take effect after you reconnect to the VPN.'
	String get reconnectToApplyChangesDescription => 'Your new settings will take effect after you reconnect to the VPN.';

	/// en: 'Got it'
	String get gotIt => 'Got it';
}

/// The flat map containing all translations for locale <en>.
/// Only for edge cases! For simple maps, use the map function of this library.
///
/// The Dart AOT compiler has issues with very large switch statements,
/// so the map is split into smaller functions (512 entries each).
extension on Translations {
	dynamic _flatMapFunction(String path) {
		return switch (path) {
			'cities.tirana' => 'Tirana',
			'cities.algiers' => 'Algiers',
			'cities.addis_ababa' => 'Addis Ababa',
			'cities.andorra_la_vella' => 'Andorra la Vella',
			'cities.buenos_aires' => 'Buenos Aires',
			'cities.yerevan' => 'Yerevan',
			'cities.adelaide' => 'Adelaide',
			'cities.brisbane' => 'Brisbane',
			'cities.melbourne' => 'Melbourne',
			'cities.perth' => 'Perth',
			'cities.sydney' => 'Sydney',
			'cities.vienna' => 'Vienna',
			'cities.baku' => 'Baku',
			'cities.nassau' => 'Nassau',
			'cities.dhaka' => 'Dhaka',
			'cities.brussels' => 'Brussels',
			'cities.belmopan' => 'Belmopan',
			'cities.hamilton' => 'Hamilton',
			'cities.thimphu' => 'Thimphu',
			'cities.la_paz' => 'La Paz',
			'cities.novi_travnik' => 'Novi Travnik',
			'cities.sao_paulo' => 'Sao Paulo',
			'cities.bandar_seri_begawan' => 'Bandar Seri Begawan',
			'cities.sofia' => 'Sofia',
			'cities.phnom_penh' => 'Phnom Penh',
			'cities.montreal' => 'Montreal',
			'cities.toronto' => 'Toronto',
			'cities.vancouver' => 'Vancouver',
			'cities.george_town' => 'George Town',
			'cities.santiago' => 'Santiago',
			'cities.bogota' => 'Bogota',
			'cities.san_jose' => 'San Jose',
			'cities.zagreb' => 'Zagreb',
			'cities.nicosia' => 'Nicosia',
			'cities.prague' => 'Prague',
			'cities.copenhagen' => 'Copenhagen',
			'cities.santo_domingo' => 'Santo Domingo',
			'cities.quito' => 'Quito',
			'cities.cairo' => 'Cairo',
			'cities.san_salvador' => 'San Salvador',
			'cities.tallinn' => 'Tallinn',
			'cities.helsinki' => 'Helsinki',
			'cities.marseille' => 'Marseille',
			'cities.paris' => 'Paris',
			'cities.tbilisi' => 'Tbilisi',
			'cities.berlin' => 'Berlin',
			'cities.frankfurt' => 'Frankfurt',
			'cities.hamburg' => 'Hamburg',
			'cities.accra' => 'Accra',
			'cities.athens' => 'Athens',
			'cities.nuuk' => 'Nuuk',
			'cities.hagatna' => 'Hagatna',
			'cities.guatemala_city' => 'Guatemala City',
			'cities.tegucigalpa' => 'Tegucigalpa',
			'cities.hong_kong' => 'Hong Kong',
			'cities.budapest' => 'Budapest',
			'cities.reykjavik' => 'Reykjavik',
			'cities.mumbai' => 'Mumbai',
			'cities.jakarta' => 'Jakarta',
			'cities.dublin' => 'Dublin',
			'cities.douglas' => 'Douglas',
			'cities.tel_aviv' => 'Tel Aviv',
			'cities.milan' => 'Milan',
			'cities.palermo' => 'Palermo',
			'cities.rome' => 'Rome',
			'cities.kingston' => 'Kingston',
			'cities.osaka' => 'Osaka',
			'cities.tokyo' => 'Tokyo',
			'cities.saint_helier' => 'Saint Helier',
			'cities.astana' => 'Astana',
			'cities.nairobi' => 'Nairobi',
			'cities.vientiane' => 'Vientiane',
			'cities.riga' => 'Riga',
			'cities.beirut' => 'Beirut',
			'cities.vaduz' => 'Vaduz',
			'cities.vilnius' => 'Vilnius',
			'cities.luxembourg' => 'Luxembourg',
			'cities.kuala_lumpur' => 'Kuala Lumpur',
			'cities.valletta' => 'Valletta',
			'cities.mexico' => 'Mexico',
			'cities.chisinau' => 'Chisinau',
			'cities.monte_carlo' => 'Monte Carlo',
			'cities.ulaanbaatar' => 'Ulaanbaatar',
			'cities.podgorica' => 'Podgorica',
			'cities.rabat' => 'Rabat',
			'cities.naypyidaw' => 'Naypyidaw',
			'cities.kathmandu' => 'Kathmandu',
			'cities.amsterdam' => 'Amsterdam',
			'cities.auckland' => 'Auckland',
			'cities.lagos' => 'Lagos',
			'cities.skopje' => 'Skopje',
			'cities.oslo' => 'Oslo',
			'cities.karachi' => 'Karachi',
			'cities.panama_city' => 'Panama City',
			'cities.port_moresby' => 'Port Moresby',
			'cities.asuncion' => 'Asuncion',
			'cities.lima' => 'Lima',
			'cities.manila' => 'Manila',
			'cities.warsaw' => 'Warsaw',
			'cities.lisbon' => 'Lisbon',
			'cities.san_juan' => 'San Juan',
			'cities.bucharest' => 'Bucharest',
			'cities.belgrade' => 'Belgrade',
			'cities.singapore' => 'Singapore',
			'cities.bratislava' => 'Bratislava',
			'cities.ljubljana' => 'Ljubljana',
			'cities.johannesburg' => 'Johannesburg',
			'cities.seoul' => 'Seoul',
			'cities.barcelona' => 'Barcelona',
			'cities.madrid' => 'Madrid',
			'cities.colombo' => 'Colombo',
			'cities.stockholm' => 'Stockholm',
			'cities.zurich' => 'Zurich',
			'cities.taipei' => 'Taipei',
			'cities.bangkok' => 'Bangkok',
			'cities.port_of_spain' => 'Port of Spain',
			'cities.istanbul' => 'Istanbul',
			'cities.kyiv' => 'Kyiv',
			'cities.dubai' => 'Dubai',
			'cities.edinburgh' => 'Edinburgh',
			'cities.glasgow' => 'Glasgow',
			'cities.london' => 'London',
			'cities.manchester' => 'Manchester',
			'cities.atlanta' => 'Atlanta',
			'cities.buffalo' => 'Buffalo',
			'cities.charlotte' => 'Charlotte',
			'cities.chicago' => 'Chicago',
			'cities.dallas' => 'Dallas',
			'cities.denver' => 'Denver',
			'cities.detroit' => 'Detroit',
			'cities.kansas_city' => 'Kansas City',
			'cities.los_angeles' => 'Los Angeles',
			'cities.manassas' => 'Manassas',
			'cities.miami' => 'Miami',
			'cities.new_york' => 'New York',
			'cities.phoenix' => 'Phoenix',
			'cities.pittsburgh' => 'Pittsburgh',
			'cities.saint_louis' => 'Saint Louis',
			'cities.salt_lake_city' => 'Salt Lake City',
			'cities.san_francisco' => 'San Francisco',
			'cities.seattle' => 'Seattle',
			'cities.montevideo' => 'Montevideo',
			'cities.tashkent' => 'Tashkent',
			'cities.caracas' => 'Caracas',
			'cities.hanoi' => 'Hanoi',
			'cities.ho_chi_minh_city' => 'Ho Chi Minh City',
			'cities.houston' => 'Houston',
			'cities.mcallen' => 'McAllen',
			'cities.luanda' => 'Luanda',
			'cities.manama' => 'Manama',
			'cities.amman' => 'Amman',
			'cities.kuwait_city' => 'Kuwait City',
			'cities.maputo' => 'Maputo',
			'cities.dakar' => 'Dakar',
			'cities.tunis' => 'Tunis',
			'cities.boston' => 'Boston',
			'cities.strasbourg' => 'Strasbourg',
			'cities.omaha' => 'Omaha',
			'cities.moroni' => 'Moroni',
			'cities.baghdad' => 'Baghdad',
			'cities.tripoli' => 'Tripoli',
			'cities.doha' => 'Doha',
			'cities.kigali' => 'Kigali',
			'cities.nashville' => 'Nashville',
			'cities.kabul' => 'Kabul',
			'cities.mogadishu' => 'Mogadishu',
			'cities.nouakchott' => 'Nouakchott',
			'cities.ashburn' => 'Ashburn',
			'cities.port_louis' => 'Port Louis',
			'cities.dushanbe' => 'Dushanbe',
			'cities.lewiston' => 'Lewiston',
			'countries.AL' => 'Albania',
			'countries.DZ' => 'Algeria',
			'countries.AD' => 'Andorra',
			'countries.AO' => 'Angola',
			'countries.AR' => 'Argentina',
			'countries.AM' => 'Armenia',
			'countries.AU' => 'Australia',
			'countries.AT' => 'Austria',
			'countries.AZ' => 'Azerbaijan',
			'countries.BS' => 'Bahamas',
			'countries.BH' => 'Bahrain',
			'countries.BD' => 'Bangladesh',
			'countries.BE' => 'Belgium',
			'countries.BZ' => 'Belize',
			'countries.BM' => 'Bermuda',
			'countries.BT' => 'Bhutan',
			'countries.BO' => 'Bolivia',
			'countries.BA' => 'Bosnia and Herzegovina',
			'countries.BR' => 'Brazil',
			'countries.BN' => 'Brunei Darussalam',
			'countries.BG' => 'Bulgaria',
			'countries.KH' => 'Cambodia',
			'countries.CA' => 'Canada',
			'countries.KY' => 'Cayman Islands',
			'countries.CL' => 'Chile',
			'countries.CO' => 'Colombia',
			'countries.CR' => 'Costa Rica',
			'countries.HR' => 'Croatia',
			'countries.CY' => 'Cyprus',
			'countries.CZ' => 'Czech Republic',
			'countries.DK' => 'Denmark',
			'countries.DO' => 'Dominican Republic',
			'countries.EC' => 'Ecuador',
			'countries.EG' => 'Egypt',
			'countries.SV' => 'El Salvador',
			'countries.EE' => 'Estonia',
			'countries.FI' => 'Finland',
			'countries.FR' => 'France',
			'countries.GE' => 'Georgia',
			'countries.DE' => 'Germany',
			'countries.GH' => 'Ghana',
			'countries.GR' => 'Greece',
			'countries.GL' => 'Greenland',
			'countries.GU' => 'Guam',
			'countries.GT' => 'Guatemala',
			'countries.HN' => 'Honduras',
			'countries.HK' => 'Hong Kong',
			'countries.HU' => 'Hungary',
			'countries.IS' => 'Iceland',
			'countries.IN' => 'India',
			'countries.ID' => 'Indonesia',
			'countries.IE' => 'Ireland',
			'countries.IM' => 'Isle of Man',
			'countries.IL' => 'Israel',
			'countries.IT' => 'Italy',
			'countries.JM' => 'Jamaica',
			'countries.JP' => 'Japan',
			'countries.JE' => 'Jersey',
			'countries.JO' => 'Jordan',
			'countries.KZ' => 'Kazakhstan',
			'countries.KE' => 'Kenya',
			'countries.KW' => 'Kuwait',
			'countries.LA' => 'Lao People\'s Democratic Republic',
			'countries.LV' => 'Latvia',
			'countries.LB' => 'Lebanon',
			'countries.LI' => 'Liechtenstein',
			'countries.LT' => 'Lithuania',
			'countries.LU' => 'Luxembourg',
			'countries.MY' => 'Malaysia',
			'countries.MT' => 'Malta',
			'countries.MX' => 'Mexico',
			'countries.MD' => 'Moldova',
			'countries.MC' => 'Monaco',
			'countries.MN' => 'Mongolia',
			'countries.ME' => 'Montenegro',
			'countries.MA' => 'Morocco',
			'countries.MZ' => 'Mozambique',
			'countries.MM' => 'Myanmar',
			'countries.NP' => 'Nepal',
			'countries.NL' => 'Netherlands',
			'countries.NZ' => 'New Zealand',
			'countries.NG' => 'Nigeria',
			'countries.MK' => 'North Macedonia',
			'countries.NO' => 'Norway',
			'countries.PK' => 'Pakistan',
			'countries.PA' => 'Panama',
			'countries.PG' => 'Papua New Guinea',
			'countries.PY' => 'Paraguay',
			'countries.PE' => 'Peru',
			'countries.PH' => 'Philippines',
			'countries.PL' => 'Poland',
			'countries.PT' => 'Portugal',
			'countries.PR' => 'Puerto Rico',
			'countries.RO' => 'Romania',
			'countries.RS' => 'Serbia',
			'countries.SN' => 'Senegal',
			'countries.SG' => 'Singapore',
			'countries.SK' => 'Slovakia',
			'countries.SI' => 'Slovenia',
			'countries.ZA' => 'South Africa',
			'countries.KR' => 'South Korea',
			'countries.ES' => 'Spain',
			'countries.LK' => 'Sri Lanka',
			'countries.SE' => 'Sweden',
			'countries.CH' => 'Switzerland',
			'countries.TW' => 'Taiwan',
			'countries.TH' => 'Thailand',
			'countries.TT' => 'Trinidad and Tobago',
			'countries.TR' => 'Turkey',
			'countries.TN' => 'Tunisia',
			'countries.UA' => 'Ukraine',
			'countries.AE' => 'United Arab Emirates',
			'countries.GB' => 'United Kingdom',
			'countries.US' => 'United States',
			'countries.UY' => 'Uruguay',
			'countries.UZ' => 'Uzbekistan',
			'countries.VE' => 'Venezuela',
			'countries.VN' => 'Vietnam',
			'daemon.code_3001_title' => 'Unauthorized',
			'daemon.code_3001_msg' => 'We couldn\'t log you in. Make sure your credentials are correct. If you have turned on MFA, log in using the \'nordvpn login\' command.',
			'daemon.code_3003_title' => 'Format error',
			'daemon.code_3003_msg' => 'The command is not valid.',
			'daemon.code_3004_title' => 'Config error',
			'daemon.code_3004_msg' => 'We ran into an issue with the config file. If the problem persists, please contact our customer support.',
			'daemon.code_3005_title' => 'Empty payload',
			'daemon.code_3005_msg' => 'Something went wrong. Please try again. If the problem persists, contact our customer support.',
			'daemon.code_3007_title' => 'You\'re offline',
			'daemon.code_3007_msg' => 'Please check your internet connection and try again.',
			'daemon.code_3008_title' => 'Account expired',
			'daemon.code_3008_msg' => 'Your account has expired. Renew your subscription now to continue enjoying the ultimate privacy and security with NordVPN.',
			'daemon.code_3010_title' => 'VPN misconfigured',
			'daemon.code_3010_msg' => 'Something went wrong. Please try again. If the problem persists, contact our customer support.',
			'daemon.code_3013_title' => 'Daemon offline',
			'daemon.code_3013_msg' => 'We couldn\'t reach System Daemon.',
			'daemon.code_3014_title' => 'Gateway error',
			'daemon.code_3014_msg' => 'It\'s not you, it\'s us. We\'re having trouble reaching our servers. If the issue persists, please contact our customer support.',
			'daemon.code_3015_title' => 'Outdated',
			'daemon.code_3015_msg' => 'A new version of NordVPN is available! Please update the app.',
			'daemon.code_3017_title' => 'Dependency error',
			'daemon.code_3017_msg' => 'Currently in use.',
			'daemon.code_3019_title' => 'No new data error',
			'daemon.code_3019_msg' => 'You have already provided a rating for your active/previous connection.',
			'daemon.code_3021_title' => 'Expired renew token',
			'daemon.code_3021_msg' => 'For security purposes, please log in again.',
			'daemon.code_3022_title' => 'Token renew error',
			'daemon.code_3022_msg' => 'We couldn\'t load your account data. Check your internet connection and try again. If the issue persists, please contact our customer support.',
			'daemon.code_3023_title' => 'Kill switch error',
			'daemon.code_3023_msg' => 'Something went wrong. Please try again. If the problem persists, contact our customer support.',
			'daemon.code_3024_title' => 'Bad request',
			'daemon.code_3024_msg' => 'Username or password is not correct. Please try again.',
			'daemon.code_3026_title' => 'Internal daemon error',
			'daemon.code_3026_msg' => 'Something went wrong. Please try again. If the problem persists, contact our customer support.',
			'daemon.code_3032_title' => 'Server not available',
			'daemon.code_3032_msg' => 'The specified server is not available at the moment or does not support your connection settings.',
			'daemon.code_3033_title' => 'Inexistent tag',
			'daemon.code_3033_msg' => 'The specified server does not exist.',
			'daemon.code_3034_title' => 'Double group error',
			'daemon.code_3034_msg' => 'You cannot connect to a group and set the group option at the same time.',
			'daemon.code_3035_title' => 'Token login error',
			'daemon.code_3035_msg' => 'Token parameter value is missing.',
			'daemon.code_3036_title' => 'Inexistent group',
			'daemon.code_3036_msg' => 'The specified group does not exist.',
			'daemon.code_3037_title' => 'Not obfuscated auto connect server',
			'daemon.code_3037_msg' => 'Your selected server doesn\'t support obfuscation. Choose a different server or turn off obfuscation.',
			'daemon.code_3038_title' => 'Obfuscated auto connect server',
			'daemon.code_3038_msg' => 'Turn on obfuscation to connect to obfuscated servers.',
			'daemon.code_3039_title' => 'Invalid token',
			'daemon.code_3039_msg' => 'We couldn\'t log you in - the access token is not valid. Please check if you\'ve entered the token correctly. If the issue persists, contact our customer support.',
			'daemon.code_3040_title' => 'Private subnet LAN discovery',
			'daemon.code_3040_msg' => 'Allowlisting a private subnet is not available while local network discovery is turned on.',
			'daemon.code_3041_title' => 'Dedicated IP renew error',
			'daemon.code_3041_msg' => 'You don’t have a dedicated IP subscription. To get a personal IP address, continue in the browser.',
			'daemon.code_3042_title' => 'Dedicated IP no server',
			'daemon.code_3042_msg' => 'This server isn\'t currently included in your dedicated IP subscription.',
			'daemon.code_3043_title' => 'Dedicated IP service but no server',
			'daemon.code_3043_msg' => 'Please select the preferred server location for your dedicated IP in Nord Account.',
			'daemon.code_3044_title' => 'Allow list invalid subnet',
			'daemon.code_3044_msg' => 'The command is not valid.',
			'daemon.code_3047_title' => 'Allow list port noop',
			'daemon.code_3047_msg' => 'Port is already on the allowlist.',
			'daemon.code_3048_title' => 'Here\'s what to know',
			'daemon.code_3048_msg' => 'The post-quantum VPN and Meshnet can\'t run at the same time. Please turn off one feature to use the other.',
			'daemon.code_3049_title' => 'Here\'s what to know',
			'daemon.code_3049_msg' => 'This setting is not compatible with post-quantum encryption. To use it, turn off post-quantum encryption first.',
			'daemon.code_3051_title' => 'Disabled technology',
			'daemon.code_3051_msg' => 'Unable to connect with the current technology. Please try a different one using the command: nordvpn set technology.',
			'daemon.code_3057_title' => 'Server access not allowed',
			'daemon.code_3057_msg' => 'To connect to the selected server, turn on virtual location access using the app’s command-line interface.',
			'daemon.code_5007_title' => 'Restart daemon to apply setting',
			'daemon.code_5007_msg' => 'Restart the daemon to apply this setting. For example, use the command `sudo systemctl restart nordvpnd` on systemd distributions.',
			'daemon.code_5008_title' => 'gRPC timeout error',
			'daemon.code_5008_msg' => 'Request time out.',
			'daemon.code_5010_title' => 'Missing exchange token',
			'daemon.code_5010_msg' => 'The exchange token is missing. Please try logging in again. If the issue persists, contact our customer support.',
			'daemon.code_5013_title' => 'Failed to open the browser',
			'daemon.code_5013_msg' => 'We couldn\'t open the browser for you to log in.',
			'daemon.code_5014_title' => 'Failed to open the browser',
			'daemon.code_5014_msg' => 'We couldn\'t open the browser for you to create an account.',
			'daemon.code_5015_title' => 'It didn\'t work this time',
			'daemon.code_5015_msg' => 'We couldn\'t connect you to the VPN. Please check your internet connection and try again. If the issue persists, contact our customer support.',
			'daemon.genericErrorTitle' => 'It didn\'t work this time',
			'daemon.genericErrorMessage' => 'Something went wrong. Please try again. If the problem persists, contact our customer support.',
			'ui.search' => 'Search',
			'ui.countries' => 'Countries',
			'ui.specialServers' => 'Specialty servers',
			'ui.cities' => 'Cities',
			'ui.noResults' => 'No results found.',
			'ui.waitingToConnectToDaemon' => 'Connecting to the daemon...',
			'ui.fetchingData' => 'Fetching data',
			'ui.failedToFetchData' => 'Failed to fetch data',
			'ui.retry' => 'Retry',
			'ui.quickConnect' => 'Quick Connect',
			'ui.register' => 'Register',
			'ui.signIn' => 'Sign in',
			'ui.disconnect' => 'Disconnect',
			'ui.double_vpn' => 'Double VPN',
			'ui.onion_over_vpn' => 'Onion Over VPN',
			'ui.fatalErrorMessage' => 'Fatal error',
			'ui.connected' => 'Connected',
			'ui.notConnected' => 'Not connected',
			'ui.connectOrPickCountry' => 'Connect now or pick a country',
			'ui.general' => 'General',
			'ui.generalSettingsSubtitle' => 'Appearance, notifications and analytics settings',
			'ui.autoConnect' => 'Auto-connect',
			'ui.autoConnected' => 'Auto-connected',
			'ui.killSwitch' => 'Kill Switch',
			'ui.account' => 'Account',
			'ui.accountSubtitle' => 'Log out, subscription',
			'ui.otherApps' => 'Apps',
			'ui.allowlist' => 'Allowlist',
			'ui.dns' => 'DNS',
			'ui.settings' => 'Settings',
			'ui.launchAppAtStartup' => 'Launch at Startup',
			'ui.vpnProtocol' => 'VPN Protocol',
			'ui.obfuscate' => 'Obfuscate',
			'ui.notificationsStatus' => 'VPN Connection Status Notifications',
			'ui.firewall' => 'Firewall',
			'ui.firewallDescription' => 'Allow the use of the system firewall. When enabled, you can attach a firewall mark to VPN packets for custom firewall rules.',
			'ui.resetToDefaults' => 'Reset all app settings to default',
			'ui.firewallMark' => 'Firewall mark',
			'ui.reset' => 'Reset',
			'ui.confirm' => 'Confirm',
			'ui.cancel' => 'Cancel',
			'ui.killSwitchDescription' => 'Disable internet access if the VPN connection drops to secure your data from accidental exposure.',
			'ui.tpLite' => 'Threat Protection Lite',
			'ui.tpLiteDescription' => 'When you\'re connected to VPN, DNS filtering blocks ads and malicious domains before any threats reach your device.',
			'ui.tpLiteWillDisableDns' => 'Enabling Threat Protection Lite will result in the removal of the custom DNS configuration. Continue?',
			'ui.customDnsWillDisableTpLite' => 'Enabling custom DNS configuration will result in the removal of the Threat Protection Lite. Continue?',
			'ui.addCustomDns' => 'Add custom DNS',
			'ui.addPort' => 'Add port',
			'ui.addPortRange' => 'Add port range',
			'ui.addSubnet' => 'Add subnet',
			'ui.loginToNordVpn' => 'Log in to NordVPN',
			'ui.newHereMessage' => 'New here? Sign up for Nord Account to get started',
			'ui.termsOfService' => 'Terms of Service',
			'ui.privacyPolicy' => 'Privacy Policy',
			'ui.autoRenewalTerms' => 'Auto - renewal terms',
			'ui.subscription' => 'Subscription',
			'ui.logout' => 'Log out',
			'ui.accountExpireIn' => ({required num n, required Object date}) => (_root.$meta.cardinalResolver ?? PluralResolvers.cardinal('en'))(n, one: 'Expires in ${n} day on ${date}', other: 'Expires in ${n} days on ${date}', ), 
			'ui.connectTo' => 'Connect to',
			'ui.recommendedServer' => 'Recommended server',
			'ui.apps' => 'Apps',
			'ui.nordVpn' => 'NordVPN',
			'ui.useNordVpnOn6Devices' => 'Use NordVPN on 6 devices at the same time at no extra cost.',
			'ui.exploreAppsAndExtensions' => 'Explore apps and browser extensions',
			'ui.scan' => 'Scan to download mobile app',
			'ui.moreApps' => 'More apps for all-around security',
			'ui.nordPass' => 'NordPass',
			'ui.nordPassDescription' => 'Generate, store, and organize your passwords.',
			'ui.nordLocker' => 'NordLocker',
			'ui.nordLockerDescription' => 'Store your files securely in our end-to-end encrypted cloud.',
			'ui.nordLayer' => 'NordLayer',
			'ui.nordLayerDescription' => 'Get a powerful security solution for your business network.',
			'ui.learnMore' => 'Learn more',
			'ui.emailSupport' => 'Email Support',
			'ui.knowledgeBase' => 'Knowledge Base',
			'ui.routing' => 'Routing',
			'ui.connectToVpn' => 'Connect to VPN',
			'ui.connecting' => 'Connecting',
			'ui.findingServer' => 'Finding server...',
			'ui.noResultsFound' => 'No results found. Try another keyword.',
			'ui.searchServersHint' => 'Search countries, cities, or servers',
			'ui.citiesAvailable' => ({required Object n}) => '${n} cities available',
			'ui.virtual' => 'Virtual',
			'ui.dedicatedIp' => 'Dedicated IP',
			'ui.doubleVpn' => 'Double VPN',
			'ui.onionOverVpn' => 'Onion over VPN',
			'ui.p2p' => 'P2P',
			'ui.europe' => 'Europe',
			'ui.theAmericas' => 'The Americas',
			'ui.asiaPacific' => 'Asia Pacific',
			'ui.africaTheMiddleEastAndIndia' => 'Africa, the Middle East, and India',
			'ui.obfuscated' => 'Obfuscated',
			'ui.obfuscatedServers' => 'Obfuscated Servers',
			'ui.selectServerForDip' => 'Pick a location for your IP',
			'ui.selectLocation' => 'Select location',
			'ui.dipSelectLocationDescription' => 'You have successfully purchased a dedicated IP – great! To start using it, select a location for your dedicated IP from the many options that we offer.',
			'ui.chooseLocationForDip' => 'Choose a location for your dedicated IP',
			'ui.getDip' => 'Get dedicated IP',
			'ui.getYourDip' => 'Get your personal IP',
			'ui.getDipDescription' => 'Get a personal IP address that belongs only to you. Enjoy all the benefits of VPN encryption without dealing with blocklists, identity checks, and selecting images of boats in CAPTCHAs.',
			'ui.notifications' => 'Notifications',
			'ui.specialtyServersSearchHint' => 'Search country or city',
			'ui.on' => 'On',
			'ui.off' => 'Off',
			'ui.invalidFormat' => 'Invalid format',
			'ui.servers' => 'Servers',
			'ui.securityAndPrivacy' => 'Security and privacy',
			'ui.securityAndPrivacySubtitle' => 'Allowlist, DNS, LAN discovery, obfuscation, firewall',
			'ui.threatProtection' => 'Threat Protection',
			'ui.threatProtectionSubtitle' => 'Blocks harmful websites, ads, and trackers',
			'ui.appearance' => 'Appearance',
			'ui.light' => 'Light',
			'ui.dark' => 'Dark',
			'ui.showNotifications' => 'Show notifications',
			'ui.vpnConnection' => 'VPN connection',
			'ui.vpnConnectionSubtitle' => 'Auto-connect, Kill Switch, protocol',
			'ui.autoConnectDescription' => 'Automatically connect to the fastest available server or your chosen server location when the app starts.',
			'ui.fastestServer' => 'Fastest',
			'ui.change' => 'Change',
			'ui.nordLynx' => 'NordLynx',
			'ui.openVpnTcp' => 'OpenVPN (TCP)',
			'ui.openVpnUdp' => 'OpenVPN (UDP)',
			'ui.autoConnectTo' => 'Auto-connect to',
			'ui.standardVpn' => 'Standard VPN',
			'ui.goBack' => 'Go back',
			'ui.done' => 'Done',
			'ui.searchCountryAndCity' => 'Search for country or city',
			'ui.resetAllCustomSettings' => 'Reset all custom settings to default?',
			'ui.resetSettingsAlertDescription' => 'This will remove your personalized configurations across the app and restore default settings.',
			'ui.resetAndDisconnectDesc' => 'This will remove your personalized configurations across the app and disconnect you from the VPN.',
			'ui.resetSettings' => 'Reset settings',
			'ui.resetAndDisconnect' => 'Reset and disconnect',
			'ui.lanDiscovery' => 'LAN discovery',
			'ui.lanDiscoveryDescription' => 'Make your device visible to other devices on your local network while connected to the VPN. Access printers, TVs, and other LAN devices.',
			'ui.customDns' => 'Custom DNS',
			'ui.customDnsDescription' => 'Set custom DNS server addresses to use.',
			'ui.routingDescription' => 'Use custom routing rules instead of the default VPN configuration.',
			'ui.postQuantumVpn' => 'Post-quantum encryption',
			'ui.postQuantumDescription' => 'Activate next-generation encryption that protects your data from threats posed by quantum computing.',
			_ => null,
		} ?? switch (path) {
			'ui.obfuscationDescription' => 'Avoid detection by traffic sensors in restricted networks while using a VPN. When enabled, only obfuscated servers are available.',
			'ui.obfuscation' => 'Obfuscation',
			'ui.add' => 'Add',
			'ui.customDnsEntries' => ({required Object n}) => 'Custom DNS: ${n}/3',
			'ui.addUpTo3DnsServers' => 'Add up to 3 DNS servers',
			'ui.nothingHereYet' => 'Nothing here yet',
			'ui.addCustomDnsDescription' => 'To activate custom DNS, add at least one DNS server.',
			'ui.threatProtectionDescription' => 'Blocks dangerous websites and flashy ads at the domain level. Works only when you’re connected to a VPN.',
			'ui.resetCustomDns' => 'Custom DNS will be reset',
			'ui.resetCustomDnsDescription' => 'Turning on Threat Protection will set your custom DNS settings to default. Continue anyway?',
			'ui.continueWord' => 'Continue',
			'ui.threatProtectionWillTurnOff' => 'Threat Protection will be turned off',
			'ui.threatProtectionWillTurnOffDescription' => 'Threat Protection works only with the default DNS. Set a custom DNS server anyway?',
			'ui.setCustomDns' => 'Set custom DNS',
			'ui.turnOffCustomDns' => 'Turn off custom DNS?',
			'ui.turnOffCustomDnsDescription' => 'This will remove all your previously added DNS servers.',
			'ui.turnOff' => 'Turn off',
			'ui.subscriptionValidationDate' => ({required String expirationDate}) => 'Active until ${expirationDate}',
			'ui.logIn' => 'Log in',
			'ui.createAccount' => 'Create account',
			'ui.whatIsNordAccount' => 'What is a Nord Account?',
			'ui.forTroubleshooting' => ({required Uri supportUrl}) => 'For troubleshooting [go to Support Center](${supportUrl})',
			'ui.copy' => 'Copy',
			'ui.copiedToClipboard' => 'Copied to clipboard',
			'ui.failedToLoadService' => 'Failed to load NordVPN service',
			'ui.tryRunningTheseCommands' => 'Try running these commands in the terminal. Then restart your device.',
			'ui.loginTitle' => 'Cybersecurity built for every day',
			'ui.verifyingLogin' => 'Verifying login status...',
			'ui.subscriptionHasEnded' => 'Your NordVPN subscription has ended',
			'ui.pleaseRenewYourSubscription' => ({required Object email}) => 'But we don’t have to say goodbye! Renew your subscription for ${email} to continue enjoying a safer and more private internet.',
			'ui.renewSubscription' => 'Renew subscription',
			'ui.appVersionIsIncompatible' => 'Your NordVPN versions are incompatible',
			'ui.appVersionIsIncompatibleDescription' => 'Please install the latest versions of this graphical interface app and the NordVPN daemon.',
			'ui.appVersionCompatibilityRecommendation' => ({required Uri compatibilityUrl}) => 'For more options, check our [compatibility guide](${compatibilityUrl})',
			'ui.turnOffKillSwitchDescription' => 'Kill Switch is blocking the login. Turn it off for now to continue.',
			'ui.turnOffKillSwitch' => 'Turn off Kill Switch',
			'ui.connectNow' => 'Connect now',
			'ui.settingAutoconnectTo' => ({required Object target}) => 'Setting auto-connect to [${target}]...',
			'ui.doubleVpnDesc' => 'Encrypt your traffic twice for extra security',
			'ui.onionOverVpnDesc' => 'Use the Onion network with VPN protection',
			'ui.p2pDesc' => 'Enjoy the best download speed',
			'ui.save' => 'Save',
			'ui.close' => 'Close',
			'ui.to' => 'to',
			'ui.useCustomDns' => 'Use custom DNS',
			'ui.useCustomDnsDescription' => 'Add up to three DNS servers.',
			'ui.enterDnsAddress' => 'Enter DNS server address',
			'ui.duplicatedDnsServer' => 'This server is already on the list.',
			'ui.obfuscationSearchWarning' => 'Obfuscation is turned on, so only obfuscated server locations will show up.',
			'ui.obfuscationErrorNoServerFound' => 'No results found. To access all available servers, turn off obfuscation.',
			'ui.goToSettings' => 'Go to Settings',
			'ui.useAllowList' => 'Use allowlist',
			'ui.useAllowListSettingDescription' => 'Exclude ports, port ranges, or subnets from VPN protection.',
			'ui.useAllowListScreenDescription' => 'Specify ports, port ranges, or subnets to exclude from VPN protection. Allowlisted ports may accept incoming connections from any external source outside your network.',
			'ui.turnOffAllowList' => 'Turn off allowlist?',
			'ui.turnOffAllowListDescription' => 'Disabling the allowlist will delete all your previously added ports, port ranges, and subnets.',
			'ui.port' => 'Port',
			'ui.portRange' => 'Port range',
			'ui.subnet' => 'Subnet',
			'ui.enterPort' => 'Enter port',
			'ui.selectProtocol' => 'Select protocol',
			'ui.all' => 'All',
			'ui.protocol' => 'Protocol',
			'ui.portAlreadyInList' => 'Port is already on the list',
			'ui.portRangeAlreadyInList' => 'The range is already on the list',
			'ui.enterPortRange' => 'Enter port range',
			'ui.enterSubnet' => 'Enter subnet',
			'ui.subnetAlreadyInList' => 'Subnet is already on the list',
			'ui.delete' => 'Delete',
			'ui.settingsWereNotSaved' => 'Settings weren\'t saved',
			'ui.couldNotSave' => 'We couldn\'t save your settings to the configuration file.',
			'ui.turnOffObfuscationServerTypes' => 'Turn off obfuscation for more server types',
			'ui.turnOffObfuscationLocations' => 'Turn off obfuscation for more locations',
			'ui.nordWhisper' => 'NordWhisper',
			'ui.system' => 'System',
			'ui.removePrivateSubnets' => 'We\'ll remove private subnets from allowlist',
			'ui.removePrivateSubnetsDescription' => 'Enabling LAN discovery will remove any private subnets from allowlist. Continue?',
			'ui.privateSubnetCantBeAdded' => 'Private subnet can\'t be added',
			'ui.privateSubnetCantBeAddedDescription' => 'Allowlisting a private subnet isn’t available while local network discovery is enabled. To add a private subnet, turn off LAN discovery.',
			'ui.turnOffLanDiscovery' => 'Turn off LAN discovery',
			'ui.startPortBiggerThanEnd' => 'Start port can’t be greater than end port',
			'ui.weCouldNotConnectToService' => 'We couldn’t connect to NordVPN service',
			'ui.needHelp' => ({required Uri supportUrl}) => 'Need help? [Visit our Support Center](${supportUrl}) ',
			'ui.issuePersists' => ({required Uri supportUrl}) => 'Issue persists? [Contact our customer support](${supportUrl})',
			'ui.meshnet' => 'Meshnet',
			'ui.systemdDistribution' => 'systemd distribution',
			'ui.nonSystemdDistro' => 'non-systemd distribution',
			'ui.tryRunningOneCommand' => 'The service isn’t running. To start it, use this command in the terminal.',
			'ui.failedToFetchConsentData' => 'Something unexpected happened while we were trying to load your analytics consent setting.',
			'ui.failedToFetchAccountData' => 'Something unexpected happened while we were trying to load your account data.',
			'ui.tryAgain' => 'Try again',
			'ui.weHitAnError' => 'We hit an error',
			'ui.weValueYourPrivacy' => 'We value your privacy',
			'ui.consentDescription' => ({required Uri privacyUrl}) => 'That’s why we want to be transparent about what data you agree to give us. We only collect the bare minimum of information required to offer a smooth and stable VPN experience.\nYour browsing activities remain private, regardless of your choice.\n\nBy selecting “Accept,” you allow us to collect and use limited app performance data for analytics, as explained in our [Privacy Policy](${privacyUrl}).\n\nSelect “Customize” to manage your privacy choices or learn more about each option.',
			'ui.accept' => 'Accept',
			'ui.customize' => 'Customize',
			'ui.rejectNonEssential' => 'Reject non-essential',
			'ui.privacyPreferences' => 'Privacy preferences',
			'ui.privacyPreferencesDescription' => 'Customize your privacy choices',
			'ui.essentialRequired' => 'Essential (Required)',
			'ui.requiredAnalyticsDescription' => ({required Uri termsUrl}) => 'Send us the essential data we need for our app to work. This also helps enforce our [Terms of Service](${termsUrl}), prevent fraud, and maintain the security of our services. ',
			'ui.confirmPreferences' => 'Confirm preferences',
			'ui.analytics' => 'Analytics',
			'ui.analyticsDescription' => 'Help us improve the app by sending aggregated usage data. We collect this data to keep our features relevant to your needs and to fix issues faster.',
			'ui.back' => 'Back',
			'ui.snapScreenTitle' => 'We need some permissions to connect to NordVPN service',
			'ui.snapScreenDescription' => 'Grant permission by running these commands in the terminal. Then refresh the screen.',
			'ui.refresh' => 'Refresh',
			'ui.terms' => 'Terms',
			'ui.termsSubtitle' => 'Learn about our legal terms.',
			'ui.termsAgreementDescription' => 'By continuing to use this app, you agree to our terms and how we handle your data. To read the terms and privacy policy check the links below.',
			'ui.readMore' => 'Read more',
			'ui.customDnsWarning' => 'Using third-party DNS may limit website availability. For the best browsing experience, use our default settings.',
			'ui.accountCreatedOn' => ({required String creation_date}) => 'Account created: ${creation_date}',
			'ui.manageSubscription' => 'Manage subscription',
			'ui.changePassword' => 'Change password',
			'ui.productHub' => 'Product Hub',
			'ui.subscriptionInactive' => 'Inactive',
			'ui.recentConnections' => 'Recent connections',
			'ui.standardVpnServer' => 'Standard VPN Servers',
			'ui.reconnectNow' => 'Reconnect now',
			'ui.reconnectToChangeProtocol' => 'Reconnect to change protocol',
			'ui.reconnectToChangeProtocolDescription' => 'To apply this change, we\'ll reconnect you to the VPN.',
			'ui.reconnectToApplyChanges' => 'Reconnect to apply changes',
			'ui.reconnectToApplyChangesDescription' => 'Your new settings will take effect after you reconnect to the VPN.',
			'ui.gotIt' => 'Got it',
			_ => null,
		};
	}
}
