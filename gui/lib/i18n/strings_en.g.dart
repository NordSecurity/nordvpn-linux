///
/// Generated file. Do not edit.
///
// coverage:ignore-file
// ignore_for_file: type=lint, unused_import

part of 'strings.g.dart';

// Path: <root>
typedef TranslationsEn = Translations; // ignore: unused_element
class Translations implements BaseTranslations<AppLocale, Translations> {
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

	/// en: 'Reconnect to VPN to apply changes'
	String get code_2002_title => 'Reconnect to VPN to apply changes';

	/// en: 'You're connected to the VPN. Please reconnect to apply the setting.'
	String get code_2002_msg => 'You\'re connected to the VPN. Please reconnect to apply the setting.';

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

	/// en: 'Terms of service'
	String get termsOfService => 'Terms of service';

	/// en: 'Privacy Policy'
	String get privacyPolicy => 'Privacy Policy';

	/// en: 'Subscription info'
	String get subscriptionInfo => 'Subscription info';

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

	/// en: 'Obfuscated'
	String get obfuscated => 'Obfuscated';

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

	/// en: 'Fastest server'
	String get fastestServer => 'Fastest server';

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

	/// en: 'Add ports, port ranges, and subnets that don’t require the VPN.'
	String get allowListDescription => 'Add ports, port ranges, and subnets that don’t require the VPN.';

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

	/// en: 'Subscription active until ${expirationDate: String}'
	String subscriptionValidationDate({required String expirationDate}) => 'Subscription active until ${expirationDate}';

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

	/// en: 'Specify ports, port ranges, or subnets to exclude from VPN protection.'
	String get useAllowListDescription => 'Specify ports, port ranges, or subnets to exclude from VPN protection.';

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
}

/// Flat map(s) containing all translations.
/// Only for edge cases! For simple maps, use the map function of this library.
extension on Translations {
	dynamic _flatMapFunction(String path) {
		switch (path) {
			case 'cities.tirana': return 'Tirana';
			case 'cities.algiers': return 'Algiers';
			case 'cities.addis_ababa': return 'Addis Ababa';
			case 'cities.andorra_la_vella': return 'Andorra la Vella';
			case 'cities.buenos_aires': return 'Buenos Aires';
			case 'cities.yerevan': return 'Yerevan';
			case 'cities.adelaide': return 'Adelaide';
			case 'cities.brisbane': return 'Brisbane';
			case 'cities.melbourne': return 'Melbourne';
			case 'cities.perth': return 'Perth';
			case 'cities.sydney': return 'Sydney';
			case 'cities.vienna': return 'Vienna';
			case 'cities.baku': return 'Baku';
			case 'cities.nassau': return 'Nassau';
			case 'cities.dhaka': return 'Dhaka';
			case 'cities.brussels': return 'Brussels';
			case 'cities.belmopan': return 'Belmopan';
			case 'cities.hamilton': return 'Hamilton';
			case 'cities.thimphu': return 'Thimphu';
			case 'cities.la_paz': return 'La Paz';
			case 'cities.novi_travnik': return 'Novi Travnik';
			case 'cities.sao_paulo': return 'Sao Paulo';
			case 'cities.bandar_seri_begawan': return 'Bandar Seri Begawan';
			case 'cities.sofia': return 'Sofia';
			case 'cities.phnom_penh': return 'Phnom Penh';
			case 'cities.montreal': return 'Montreal';
			case 'cities.toronto': return 'Toronto';
			case 'cities.vancouver': return 'Vancouver';
			case 'cities.george_town': return 'George Town';
			case 'cities.santiago': return 'Santiago';
			case 'cities.bogota': return 'Bogota';
			case 'cities.san_jose': return 'San Jose';
			case 'cities.zagreb': return 'Zagreb';
			case 'cities.nicosia': return 'Nicosia';
			case 'cities.prague': return 'Prague';
			case 'cities.copenhagen': return 'Copenhagen';
			case 'cities.santo_domingo': return 'Santo Domingo';
			case 'cities.quito': return 'Quito';
			case 'cities.cairo': return 'Cairo';
			case 'cities.san_salvador': return 'San Salvador';
			case 'cities.tallinn': return 'Tallinn';
			case 'cities.helsinki': return 'Helsinki';
			case 'cities.marseille': return 'Marseille';
			case 'cities.paris': return 'Paris';
			case 'cities.tbilisi': return 'Tbilisi';
			case 'cities.berlin': return 'Berlin';
			case 'cities.frankfurt': return 'Frankfurt';
			case 'cities.hamburg': return 'Hamburg';
			case 'cities.accra': return 'Accra';
			case 'cities.athens': return 'Athens';
			case 'cities.nuuk': return 'Nuuk';
			case 'cities.hagatna': return 'Hagatna';
			case 'cities.guatemala_city': return 'Guatemala City';
			case 'cities.tegucigalpa': return 'Tegucigalpa';
			case 'cities.hong_kong': return 'Hong Kong';
			case 'cities.budapest': return 'Budapest';
			case 'cities.reykjavik': return 'Reykjavik';
			case 'cities.mumbai': return 'Mumbai';
			case 'cities.jakarta': return 'Jakarta';
			case 'cities.dublin': return 'Dublin';
			case 'cities.douglas': return 'Douglas';
			case 'cities.tel_aviv': return 'Tel Aviv';
			case 'cities.milan': return 'Milan';
			case 'cities.palermo': return 'Palermo';
			case 'cities.rome': return 'Rome';
			case 'cities.kingston': return 'Kingston';
			case 'cities.osaka': return 'Osaka';
			case 'cities.tokyo': return 'Tokyo';
			case 'cities.saint_helier': return 'Saint Helier';
			case 'cities.astana': return 'Astana';
			case 'cities.nairobi': return 'Nairobi';
			case 'cities.vientiane': return 'Vientiane';
			case 'cities.riga': return 'Riga';
			case 'cities.beirut': return 'Beirut';
			case 'cities.vaduz': return 'Vaduz';
			case 'cities.vilnius': return 'Vilnius';
			case 'cities.luxembourg': return 'Luxembourg';
			case 'cities.kuala_lumpur': return 'Kuala Lumpur';
			case 'cities.valletta': return 'Valletta';
			case 'cities.mexico': return 'Mexico';
			case 'cities.chisinau': return 'Chisinau';
			case 'cities.monte_carlo': return 'Monte Carlo';
			case 'cities.ulaanbaatar': return 'Ulaanbaatar';
			case 'cities.podgorica': return 'Podgorica';
			case 'cities.rabat': return 'Rabat';
			case 'cities.naypyidaw': return 'Naypyidaw';
			case 'cities.kathmandu': return 'Kathmandu';
			case 'cities.amsterdam': return 'Amsterdam';
			case 'cities.auckland': return 'Auckland';
			case 'cities.lagos': return 'Lagos';
			case 'cities.skopje': return 'Skopje';
			case 'cities.oslo': return 'Oslo';
			case 'cities.karachi': return 'Karachi';
			case 'cities.panama_city': return 'Panama City';
			case 'cities.port_moresby': return 'Port Moresby';
			case 'cities.asuncion': return 'Asuncion';
			case 'cities.lima': return 'Lima';
			case 'cities.manila': return 'Manila';
			case 'cities.warsaw': return 'Warsaw';
			case 'cities.lisbon': return 'Lisbon';
			case 'cities.san_juan': return 'San Juan';
			case 'cities.bucharest': return 'Bucharest';
			case 'cities.belgrade': return 'Belgrade';
			case 'cities.singapore': return 'Singapore';
			case 'cities.bratislava': return 'Bratislava';
			case 'cities.ljubljana': return 'Ljubljana';
			case 'cities.johannesburg': return 'Johannesburg';
			case 'cities.seoul': return 'Seoul';
			case 'cities.barcelona': return 'Barcelona';
			case 'cities.madrid': return 'Madrid';
			case 'cities.colombo': return 'Colombo';
			case 'cities.stockholm': return 'Stockholm';
			case 'cities.zurich': return 'Zurich';
			case 'cities.taipei': return 'Taipei';
			case 'cities.bangkok': return 'Bangkok';
			case 'cities.port_of_spain': return 'Port of Spain';
			case 'cities.istanbul': return 'Istanbul';
			case 'cities.kyiv': return 'Kyiv';
			case 'cities.dubai': return 'Dubai';
			case 'cities.edinburgh': return 'Edinburgh';
			case 'cities.glasgow': return 'Glasgow';
			case 'cities.london': return 'London';
			case 'cities.manchester': return 'Manchester';
			case 'cities.atlanta': return 'Atlanta';
			case 'cities.buffalo': return 'Buffalo';
			case 'cities.charlotte': return 'Charlotte';
			case 'cities.chicago': return 'Chicago';
			case 'cities.dallas': return 'Dallas';
			case 'cities.denver': return 'Denver';
			case 'cities.detroit': return 'Detroit';
			case 'cities.kansas_city': return 'Kansas City';
			case 'cities.los_angeles': return 'Los Angeles';
			case 'cities.manassas': return 'Manassas';
			case 'cities.miami': return 'Miami';
			case 'cities.new_york': return 'New York';
			case 'cities.phoenix': return 'Phoenix';
			case 'cities.saint_louis': return 'Saint Louis';
			case 'cities.salt_lake_city': return 'Salt Lake City';
			case 'cities.san_francisco': return 'San Francisco';
			case 'cities.seattle': return 'Seattle';
			case 'cities.montevideo': return 'Montevideo';
			case 'cities.tashkent': return 'Tashkent';
			case 'cities.caracas': return 'Caracas';
			case 'cities.hanoi': return 'Hanoi';
			case 'cities.ho_chi_minh_city': return 'Ho Chi Minh City';
			case 'cities.houston': return 'Houston';
			case 'cities.mcallen': return 'McAllen';
			case 'cities.luanda': return 'Luanda';
			case 'cities.manama': return 'Manama';
			case 'cities.amman': return 'Amman';
			case 'cities.kuwait_city': return 'Kuwait City';
			case 'cities.maputo': return 'Maputo';
			case 'cities.dakar': return 'Dakar';
			case 'cities.tunis': return 'Tunis';
			case 'cities.boston': return 'Boston';
			case 'cities.strasbourg': return 'Strasbourg';
			case 'cities.omaha': return 'Omaha';
			case 'cities.moroni': return 'Moroni';
			case 'cities.baghdad': return 'Baghdad';
			case 'cities.tripoli': return 'Tripoli';
			case 'cities.doha': return 'Doha';
			case 'cities.kigali': return 'Kigali';
			case 'cities.nashville': return 'Nashville';
			case 'cities.kabul': return 'Kabul';
			case 'cities.mogadishu': return 'Mogadishu';
			case 'cities.nouakchott': return 'Nouakchott';
			case 'cities.ashburn': return 'Ashburn';
			case 'countries.AL': return 'Albania';
			case 'countries.DZ': return 'Algeria';
			case 'countries.AD': return 'Andorra';
			case 'countries.AO': return 'Angola';
			case 'countries.AR': return 'Argentina';
			case 'countries.AM': return 'Armenia';
			case 'countries.AU': return 'Australia';
			case 'countries.AT': return 'Austria';
			case 'countries.AZ': return 'Azerbaijan';
			case 'countries.BS': return 'Bahamas';
			case 'countries.BH': return 'Bahrain';
			case 'countries.BD': return 'Bangladesh';
			case 'countries.BE': return 'Belgium';
			case 'countries.BZ': return 'Belize';
			case 'countries.BM': return 'Bermuda';
			case 'countries.BT': return 'Bhutan';
			case 'countries.BO': return 'Bolivia';
			case 'countries.BA': return 'Bosnia and Herzegovina';
			case 'countries.BR': return 'Brazil';
			case 'countries.BN': return 'Brunei Darussalam';
			case 'countries.BG': return 'Bulgaria';
			case 'countries.KH': return 'Cambodia';
			case 'countries.CA': return 'Canada';
			case 'countries.KY': return 'Cayman Islands';
			case 'countries.CL': return 'Chile';
			case 'countries.CO': return 'Colombia';
			case 'countries.CR': return 'Costa Rica';
			case 'countries.HR': return 'Croatia';
			case 'countries.CY': return 'Cyprus';
			case 'countries.CZ': return 'Czech Republic';
			case 'countries.DK': return 'Denmark';
			case 'countries.DO': return 'Dominican Republic';
			case 'countries.EC': return 'Ecuador';
			case 'countries.EG': return 'Egypt';
			case 'countries.SV': return 'El Salvador';
			case 'countries.EE': return 'Estonia';
			case 'countries.FI': return 'Finland';
			case 'countries.FR': return 'France';
			case 'countries.GE': return 'Georgia';
			case 'countries.DE': return 'Germany';
			case 'countries.GH': return 'Ghana';
			case 'countries.GR': return 'Greece';
			case 'countries.GL': return 'Greenland';
			case 'countries.GU': return 'Guam';
			case 'countries.GT': return 'Guatemala';
			case 'countries.HN': return 'Honduras';
			case 'countries.HK': return 'Hong Kong';
			case 'countries.HU': return 'Hungary';
			case 'countries.IS': return 'Iceland';
			case 'countries.IN': return 'India';
			case 'countries.ID': return 'Indonesia';
			case 'countries.IE': return 'Ireland';
			case 'countries.IM': return 'Isle of Man';
			case 'countries.IL': return 'Israel';
			case 'countries.IT': return 'Italy';
			case 'countries.JM': return 'Jamaica';
			case 'countries.JP': return 'Japan';
			case 'countries.JE': return 'Jersey';
			case 'countries.JO': return 'Jordan';
			case 'countries.KZ': return 'Kazakhstan';
			case 'countries.KE': return 'Kenya';
			case 'countries.KW': return 'Kuwait';
			case 'countries.LA': return 'Lao People\'s Democratic Republic';
			case 'countries.LV': return 'Latvia';
			case 'countries.LB': return 'Lebanon';
			case 'countries.LI': return 'Liechtenstein';
			case 'countries.LT': return 'Lithuania';
			case 'countries.LU': return 'Luxembourg';
			case 'countries.MY': return 'Malaysia';
			case 'countries.MT': return 'Malta';
			case 'countries.MX': return 'Mexico';
			case 'countries.MD': return 'Moldova';
			case 'countries.MC': return 'Monaco';
			case 'countries.MN': return 'Mongolia';
			case 'countries.ME': return 'Montenegro';
			case 'countries.MA': return 'Morocco';
			case 'countries.MZ': return 'Mozambique';
			case 'countries.MM': return 'Myanmar';
			case 'countries.NP': return 'Nepal';
			case 'countries.NL': return 'Netherlands';
			case 'countries.NZ': return 'New Zealand';
			case 'countries.NG': return 'Nigeria';
			case 'countries.MK': return 'North Macedonia';
			case 'countries.NO': return 'Norway';
			case 'countries.PK': return 'Pakistan';
			case 'countries.PA': return 'Panama';
			case 'countries.PG': return 'Papua New Guinea';
			case 'countries.PY': return 'Paraguay';
			case 'countries.PE': return 'Peru';
			case 'countries.PH': return 'Philippines';
			case 'countries.PL': return 'Poland';
			case 'countries.PT': return 'Portugal';
			case 'countries.PR': return 'Puerto Rico';
			case 'countries.RO': return 'Romania';
			case 'countries.RS': return 'Serbia';
			case 'countries.SN': return 'Senegal';
			case 'countries.SG': return 'Singapore';
			case 'countries.SK': return 'Slovakia';
			case 'countries.SI': return 'Slovenia';
			case 'countries.ZA': return 'South Africa';
			case 'countries.KR': return 'South Korea';
			case 'countries.ES': return 'Spain';
			case 'countries.LK': return 'Sri Lanka';
			case 'countries.SE': return 'Sweden';
			case 'countries.CH': return 'Switzerland';
			case 'countries.TW': return 'Taiwan';
			case 'countries.TH': return 'Thailand';
			case 'countries.TT': return 'Trinidad and Tobago';
			case 'countries.TR': return 'Turkey';
			case 'countries.TN': return 'Tunisia';
			case 'countries.UA': return 'Ukraine';
			case 'countries.AE': return 'United Arab Emirates';
			case 'countries.GB': return 'United Kingdom';
			case 'countries.US': return 'United States';
			case 'countries.UY': return 'Uruguay';
			case 'countries.UZ': return 'Uzbekistan';
			case 'countries.VE': return 'Venezuela';
			case 'countries.VN': return 'Vietnam';
			case 'daemon.code_2002_title': return 'Reconnect to VPN to apply changes';
			case 'daemon.code_2002_msg': return 'You\'re connected to the VPN. Please reconnect to apply the setting.';
			case 'daemon.code_3001_title': return 'Unauthorized';
			case 'daemon.code_3001_msg': return 'We couldn\'t log you in. Make sure your credentials are correct. If you have turned on MFA, log in using the \'nordvpn login\' command.';
			case 'daemon.code_3003_title': return 'Format error';
			case 'daemon.code_3003_msg': return 'The command is not valid.';
			case 'daemon.code_3004_title': return 'Config error';
			case 'daemon.code_3004_msg': return 'We ran into an issue with the config file. If the problem persists, please contact our customer support.';
			case 'daemon.code_3005_title': return 'Empty payload';
			case 'daemon.code_3005_msg': return 'Something went wrong. Please try again. If the problem persists, contact our customer support.';
			case 'daemon.code_3007_title': return 'You\'re offline';
			case 'daemon.code_3007_msg': return 'Please check your internet connection and try again.';
			case 'daemon.code_3008_title': return 'Account expired';
			case 'daemon.code_3008_msg': return 'Your account has expired. Renew your subscription now to continue enjoying the ultimate privacy and security with NordVPN.';
			case 'daemon.code_3010_title': return 'VPN misconfigured';
			case 'daemon.code_3010_msg': return 'Something went wrong. Please try again. If the problem persists, contact our customer support.';
			case 'daemon.code_3013_title': return 'Daemon offline';
			case 'daemon.code_3013_msg': return 'We couldn\'t reach System Daemon.';
			case 'daemon.code_3014_title': return 'Gateway error';
			case 'daemon.code_3014_msg': return 'It\'s not you, it\'s us. We\'re having trouble reaching our servers. If the issue persists, please contact our customer support.';
			case 'daemon.code_3015_title': return 'Outdated';
			case 'daemon.code_3015_msg': return 'A new version of NordVPN is available! Please update the app.';
			case 'daemon.code_3017_title': return 'Dependency error';
			case 'daemon.code_3017_msg': return 'Currently in use.';
			case 'daemon.code_3019_title': return 'No new data error';
			case 'daemon.code_3019_msg': return 'You have already provided a rating for your active/previous connection.';
			case 'daemon.code_3021_title': return 'Expired renew token';
			case 'daemon.code_3021_msg': return 'For security purposes, please log in again.';
			case 'daemon.code_3022_title': return 'Token renew error';
			case 'daemon.code_3022_msg': return 'We couldn\'t load your account data. Check your internet connection and try again. If the issue persists, please contact our customer support.';
			case 'daemon.code_3023_title': return 'Kill switch error';
			case 'daemon.code_3023_msg': return 'Something went wrong. Please try again. If the problem persists, contact our customer support.';
			case 'daemon.code_3024_title': return 'Bad request';
			case 'daemon.code_3024_msg': return 'Username or password is not correct. Please try again.';
			case 'daemon.code_3026_title': return 'Internal daemon error';
			case 'daemon.code_3026_msg': return 'Something went wrong. Please try again. If the problem persists, contact our customer support.';
			case 'daemon.code_3032_title': return 'Server not available';
			case 'daemon.code_3032_msg': return 'The specified server is not available at the moment or does not support your connection settings.';
			case 'daemon.code_3033_title': return 'Inexistent tag';
			case 'daemon.code_3033_msg': return 'The specified server does not exist.';
			case 'daemon.code_3034_title': return 'Double group error';
			case 'daemon.code_3034_msg': return 'You cannot connect to a group and set the group option at the same time.';
			case 'daemon.code_3035_title': return 'Token login error';
			case 'daemon.code_3035_msg': return 'Token parameter value is missing.';
			case 'daemon.code_3036_title': return 'Inexistent group';
			case 'daemon.code_3036_msg': return 'The specified group does not exist.';
			case 'daemon.code_3037_title': return 'Not obfuscated auto connect server';
			case 'daemon.code_3037_msg': return 'Your selected server doesn\'t support obfuscation. Choose a different server or turn off obfuscation.';
			case 'daemon.code_3038_title': return 'Obfuscated auto connect server';
			case 'daemon.code_3038_msg': return 'Turn on obfuscation to connect to obfuscated servers.';
			case 'daemon.code_3039_title': return 'Invalid token';
			case 'daemon.code_3039_msg': return 'We couldn\'t log you in - the access token is not valid. Please check if you\'ve entered the token correctly. If the issue persists, contact our customer support.';
			case 'daemon.code_3040_title': return 'Private subnet LAN discovery';
			case 'daemon.code_3040_msg': return 'Allowlisting a private subnet is not available while local network discovery is turned on.';
			case 'daemon.code_3041_title': return 'Dedicated IP renew error';
			case 'daemon.code_3041_msg': return 'You don’t have a dedicated IP subscription. To get a personal IP address, continue in the browser.';
			case 'daemon.code_3042_title': return 'Dedicated IP no server';
			case 'daemon.code_3042_msg': return 'This server isn\'t currently included in your dedicated IP subscription.';
			case 'daemon.code_3043_title': return 'Dedicated IP service but no server';
			case 'daemon.code_3043_msg': return 'Please select the preferred server location for your dedicated IP in Nord Account.';
			case 'daemon.code_3044_title': return 'Allow list invalid subnet';
			case 'daemon.code_3044_msg': return 'The command is not valid.';
			case 'daemon.code_3047_title': return 'Allow list port noop';
			case 'daemon.code_3047_msg': return 'Port is already on the allowlist.';
			case 'daemon.code_3048_title': return 'Here\'s what to know';
			case 'daemon.code_3048_msg': return 'The post-quantum VPN and Meshnet can\'t run at the same time. Please turn off one feature to use the other.';
			case 'daemon.code_3049_title': return 'Here\'s what to know';
			case 'daemon.code_3049_msg': return 'This setting is not compatible with post-quantum encryption. To use it, turn off post-quantum encryption first.';
			case 'daemon.code_3051_title': return 'Disabled technology';
			case 'daemon.code_3051_msg': return 'Unable to connect with the current technology. Please try a different one using the command: nordvpn set technology.';
			case 'daemon.code_5007_title': return 'Restart daemon to apply setting';
			case 'daemon.code_5007_msg': return 'Restart the daemon to apply this setting. For example, use the command `sudo systemctl restart nordvpnd` on systemd distributions.';
			case 'daemon.code_5008_title': return 'gRPC timeout error';
			case 'daemon.code_5008_msg': return 'Request time out.';
			case 'daemon.code_5010_title': return 'Missing exchange token';
			case 'daemon.code_5010_msg': return 'The exchange token is missing. Please try logging in again. If the issue persists, contact our customer support.';
			case 'daemon.code_5013_title': return 'Failed to open the browser';
			case 'daemon.code_5013_msg': return 'We couldn\'t open the browser for you to log in.';
			case 'daemon.code_5014_title': return 'Failed to open the browser';
			case 'daemon.code_5014_msg': return 'We couldn\'t open the browser for you to create an account.';
			case 'daemon.code_5015_title': return 'It didn\'t work this time';
			case 'daemon.code_5015_msg': return 'We couldn\'t connect you to the VPN. Please check your internet connection and try again. If the issue persists, contact our customer support.';
			case 'daemon.genericErrorTitle': return 'It didn\'t work this time';
			case 'daemon.genericErrorMessage': return 'Something went wrong. Please try again. If the problem persists, contact our customer support.';
			case 'ui.search': return 'Search';
			case 'ui.countries': return 'Countries';
			case 'ui.specialServers': return 'Specialty servers';
			case 'ui.cities': return 'Cities';
			case 'ui.noResults': return 'No results found.';
			case 'ui.waitingToConnectToDaemon': return 'Connecting to the daemon...';
			case 'ui.fetchingData': return 'Fetching data';
			case 'ui.failedToFetchData': return 'Failed to fetch data';
			case 'ui.retry': return 'Retry';
			case 'ui.quickConnect': return 'Quick Connect';
			case 'ui.register': return 'Register';
			case 'ui.signIn': return 'Sign in';
			case 'ui.disconnect': return 'Disconnect';
			case 'ui.double_vpn': return 'Double VPN';
			case 'ui.onion_over_vpn': return 'Onion Over VPN';
			case 'ui.fatalErrorMessage': return 'Fatal error';
			case 'ui.connected': return 'Connected';
			case 'ui.notConnected': return 'Not connected';
			case 'ui.connectOrPickCountry': return 'Connect now or pick a country';
			case 'ui.general': return 'General';
			case 'ui.generalSettingsSubtitle': return 'Appearance, notifications and analytics settings';
			case 'ui.autoConnect': return 'Auto-connect';
			case 'ui.autoConnected': return 'Auto-connected';
			case 'ui.killSwitch': return 'Kill Switch';
			case 'ui.account': return 'Account';
			case 'ui.accountSubtitle': return 'Log out, subscription';
			case 'ui.otherApps': return 'Apps';
			case 'ui.allowlist': return 'Allowlist';
			case 'ui.dns': return 'DNS';
			case 'ui.settings': return 'Settings';
			case 'ui.launchAppAtStartup': return 'Launch at Startup';
			case 'ui.vpnProtocol': return 'VPN Protocol';
			case 'ui.obfuscate': return 'Obfuscate';
			case 'ui.notificationsStatus': return 'VPN Connection Status Notifications';
			case 'ui.firewall': return 'Firewall';
			case 'ui.firewallDescription': return 'Allow the use of the system firewall. When enabled, you can attach a firewall mark to VPN packets for custom firewall rules.';
			case 'ui.resetToDefaults': return 'Reset all app settings to default';
			case 'ui.firewallMark': return 'Firewall mark';
			case 'ui.reset': return 'Reset';
			case 'ui.confirm': return 'Confirm';
			case 'ui.cancel': return 'Cancel';
			case 'ui.killSwitchDescription': return 'Disable internet access if the VPN connection drops to secure your data from accidental exposure.';
			case 'ui.tpLite': return 'Threat Protection Lite';
			case 'ui.tpLiteDescription': return 'When you\'re connected to VPN, DNS filtering blocks ads and malicious domains before any threats reach your device.';
			case 'ui.tpLiteWillDisableDns': return 'Enabling Threat Protection Lite will result in the removal of the custom DNS configuration. Continue?';
			case 'ui.customDnsWillDisableTpLite': return 'Enabling custom DNS configuration will result in the removal of the Threat Protection Lite. Continue?';
			case 'ui.addCustomDns': return 'Add custom DNS';
			case 'ui.addPort': return 'Add port';
			case 'ui.addPortRange': return 'Add port range';
			case 'ui.addSubnet': return 'Add subnet';
			case 'ui.loginToNordVpn': return 'Log in to NordVPN';
			case 'ui.newHereMessage': return 'New here? Sign up for Nord Account to get started';
			case 'ui.termsOfService': return 'Terms of service';
			case 'ui.privacyPolicy': return 'Privacy Policy';
			case 'ui.subscriptionInfo': return 'Subscription info';
			case 'ui.logout': return 'Log out';
			case 'ui.accountExpireIn': return ({required num n, required Object date}) => (_root.$meta.cardinalResolver ?? PluralResolvers.cardinal('en'))(n,
				one: 'Expires in ${n} day on ${date}',
				other: 'Expires in ${n} days on ${date}',
			);
			case 'ui.connectTo': return 'Connect to';
			case 'ui.recommendedServer': return 'Recommended server';
			case 'ui.apps': return 'Apps';
			case 'ui.nordVpn': return 'NordVPN';
			case 'ui.useNordVpnOn6Devices': return 'Use NordVPN on 6 devices at the same time at no extra cost.';
			case 'ui.exploreAppsAndExtensions': return 'Explore apps and browser extensions';
			case 'ui.scan': return 'Scan to download mobile app';
			case 'ui.moreApps': return 'More apps for all-around security';
			case 'ui.nordPass': return 'NordPass';
			case 'ui.nordPassDescription': return 'Generate, store, and organize your passwords.';
			case 'ui.nordLocker': return 'NordLocker';
			case 'ui.nordLockerDescription': return 'Store your files securely in our end-to-end encrypted cloud.';
			case 'ui.nordLayer': return 'NordLayer';
			case 'ui.nordLayerDescription': return 'Get a powerful security solution for your business network.';
			case 'ui.learnMore': return 'Learn more';
			case 'ui.emailSupport': return 'Email Support';
			case 'ui.knowledgeBase': return 'Knowledge Base';
			case 'ui.routing': return 'Routing';
			case 'ui.connectToVpn': return 'Connect to VPN';
			case 'ui.connecting': return 'Connecting';
			case 'ui.findingServer': return 'Finding server...';
			case 'ui.noResultsFound': return 'No results found. Try another keyword.';
			case 'ui.searchServersHint': return 'Search countries, cities, or servers';
			case 'ui.citiesAvailable': return ({required Object n}) => '${n} cities available';
			case 'ui.virtual': return 'Virtual';
			case 'ui.dedicatedIp': return 'Dedicated IP';
			case 'ui.doubleVpn': return 'Double VPN';
			case 'ui.onionOverVpn': return 'Onion over VPN';
			case 'ui.p2p': return 'P2P';
			case 'ui.obfuscated': return 'Obfuscated';
			case 'ui.selectServerForDip': return 'Pick a location for your IP';
			case 'ui.selectLocation': return 'Select location';
			case 'ui.dipSelectLocationDescription': return 'You have successfully purchased a dedicated IP – great! To start using it, select a location for your dedicated IP from the many options that we offer.';
			case 'ui.chooseLocationForDip': return 'Choose a location for your dedicated IP';
			case 'ui.getDip': return 'Get dedicated IP';
			case 'ui.getYourDip': return 'Get your personal IP';
			case 'ui.getDipDescription': return 'Get a personal IP address that belongs only to you. Enjoy all the benefits of VPN encryption without dealing with blocklists, identity checks, and selecting images of boats in CAPTCHAs.';
			case 'ui.notifications': return 'Notifications';
			case 'ui.specialtyServersSearchHint': return 'Search country or city';
			case 'ui.on': return 'On';
			case 'ui.off': return 'Off';
			case 'ui.invalidFormat': return 'Invalid format';
			case 'ui.servers': return 'Servers';
			case 'ui.securityAndPrivacy': return 'Security and privacy';
			case 'ui.securityAndPrivacySubtitle': return 'Allowlist, DNS, LAN discovery, obfuscation, firewall';
			case 'ui.threatProtection': return 'Threat Protection';
			case 'ui.threatProtectionSubtitle': return 'Blocks harmful websites, ads, and trackers';
			case 'ui.appearance': return 'Appearance';
			case 'ui.light': return 'Light';
			case 'ui.dark': return 'Dark';
			case 'ui.showNotifications': return 'Show notifications';
			case 'ui.vpnConnection': return 'VPN connection';
			case 'ui.vpnConnectionSubtitle': return 'Auto-connect, Kill Switch, protocol';
			case 'ui.autoConnectDescription': return 'Automatically connect to the fastest available server or your chosen server location when the app starts.';
			case 'ui.fastestServer': return 'Fastest server';
			case 'ui.change': return 'Change';
			case 'ui.nordLynx': return 'NordLynx';
			case 'ui.openVpnTcp': return 'OpenVPN (TCP)';
			case 'ui.openVpnUdp': return 'OpenVPN (UDP)';
			case 'ui.autoConnectTo': return 'Auto-connect to';
			case 'ui.standardVpn': return 'Standard VPN';
			case 'ui.goBack': return 'Go back';
			case 'ui.done': return 'Done';
			case 'ui.searchCountryAndCity': return 'Search for country or city';
			case 'ui.resetAllCustomSettings': return 'Reset all custom settings to default?';
			case 'ui.resetSettingsAlertDescription': return 'This will remove your personalized configurations across the app and restore default settings.';
			case 'ui.resetAndDisconnectDesc': return 'This will remove your personalized configurations across the app and disconnect you from the VPN.';
			case 'ui.resetSettings': return 'Reset settings';
			case 'ui.resetAndDisconnect': return 'Reset and disconnect';
			case 'ui.allowListDescription': return 'Add ports, port ranges, and subnets that don’t require the VPN.';
			case 'ui.lanDiscovery': return 'LAN discovery';
			case 'ui.lanDiscoveryDescription': return 'Make your device visible to other devices on your local network while connected to the VPN. Access printers, TVs, and other LAN devices.';
			case 'ui.customDns': return 'Custom DNS';
			case 'ui.customDnsDescription': return 'Set custom DNS server addresses to use.';
			case 'ui.routingDescription': return 'Use custom routing rules instead of the default VPN configuration.';
			case 'ui.postQuantumVpn': return 'Post-quantum encryption';
			case 'ui.postQuantumDescription': return 'Activate next-generation encryption that protects your data from threats posed by quantum computing.';
			case 'ui.obfuscationDescription': return 'Avoid detection by traffic sensors in restricted networks while using a VPN. When enabled, only obfuscated servers are available.';
			case 'ui.obfuscation': return 'Obfuscation';
			case 'ui.add': return 'Add';
			case 'ui.customDnsEntries': return ({required Object n}) => 'Custom DNS: ${n}/3';
			case 'ui.addUpTo3DnsServers': return 'Add up to 3 DNS servers';
			case 'ui.nothingHereYet': return 'Nothing here yet';
			case 'ui.addCustomDnsDescription': return 'To activate custom DNS, add at least one DNS server.';
			case 'ui.threatProtectionDescription': return 'Blocks dangerous websites and flashy ads at the domain level. Works only when you’re connected to a VPN.';
			case 'ui.resetCustomDns': return 'Custom DNS will be reset';
			case 'ui.resetCustomDnsDescription': return 'Turning on Threat Protection will set your custom DNS settings to default. Continue anyway?';
			case 'ui.continueWord': return 'Continue';
			case 'ui.threatProtectionWillTurnOff': return 'Threat Protection will be turned off';
			case 'ui.threatProtectionWillTurnOffDescription': return 'Threat Protection works only with the default DNS. Set a custom DNS server anyway?';
			case 'ui.setCustomDns': return 'Set custom DNS';
			case 'ui.turnOffCustomDns': return 'Turn off custom DNS?';
			case 'ui.turnOffCustomDnsDescription': return 'This will remove all your previously added DNS servers.';
			case 'ui.turnOff': return 'Turn off';
			case 'ui.subscriptionValidationDate': return ({required String expirationDate}) => 'Subscription active until ${expirationDate}';
			case 'ui.logIn': return 'Log in';
			case 'ui.createAccount': return 'Create account';
			case 'ui.whatIsNordAccount': return 'What is a Nord Account?';
			case 'ui.forTroubleshooting': return ({required Uri supportUrl}) => 'For troubleshooting [go to Support Center](${supportUrl})';
			case 'ui.copy': return 'Copy';
			case 'ui.copiedToClipboard': return 'Copied to clipboard';
			case 'ui.failedToLoadService': return 'Failed to load NordVPN service';
			case 'ui.tryRunningTheseCommands': return 'Try running these commands in the terminal. Then restart your device.';
			case 'ui.loginTitle': return 'Cybersecurity built for every day';
			case 'ui.verifyingLogin': return 'Verifying login status...';
			case 'ui.subscriptionHasEnded': return 'Your NordVPN subscription has ended';
			case 'ui.pleaseRenewYourSubscription': return ({required Object email}) => 'But we don’t have to say goodbye! Renew your subscription for ${email} to continue enjoying a safer and more private internet.';
			case 'ui.renewSubscription': return 'Renew subscription';
			case 'ui.appVersionIsIncompatible': return 'Your NordVPN versions are incompatible';
			case 'ui.appVersionIsIncompatibleDescription': return 'Please install the latest versions of this graphical interface app and the NordVPN daemon.';
			case 'ui.appVersionCompatibilityRecommendation': return ({required Uri compatibilityUrl}) => 'For more options, check our [compatibility guide](${compatibilityUrl})';
			case 'ui.turnOffKillSwitchDescription': return 'Kill Switch is blocking the login. Turn it off for now to continue.';
			case 'ui.turnOffKillSwitch': return 'Turn off Kill Switch';
			case 'ui.connectNow': return 'Connect now';
			case 'ui.settingAutoconnectTo': return ({required Object target}) => 'Setting auto-connect to [${target}]...';
			case 'ui.doubleVpnDesc': return 'Encrypt your traffic twice for extra security';
			case 'ui.onionOverVpnDesc': return 'Use the Onion network with VPN protection';
			case 'ui.p2pDesc': return 'Enjoy the best download speed';
			case 'ui.save': return 'Save';
			case 'ui.close': return 'Close';
			case 'ui.to': return 'to';
			case 'ui.useCustomDns': return 'Use custom DNS';
			case 'ui.useCustomDnsDescription': return 'Add up to three DNS servers.';
			case 'ui.enterDnsAddress': return 'Enter DNS server address';
			case 'ui.duplicatedDnsServer': return 'This server is already on the list.';
			case 'ui.obfuscationSearchWarning': return 'Obfuscation is turned on, so only obfuscated server locations will show up.';
			case 'ui.obfuscationErrorNoServerFound': return 'No results found. To access all available servers, turn off obfuscation.';
			case 'ui.goToSettings': return 'Go to Settings';
			case 'ui.useAllowList': return 'Use allowlist';
			case 'ui.useAllowListDescription': return 'Specify ports, port ranges, or subnets to exclude from VPN protection.';
			case 'ui.turnOffAllowList': return 'Turn off allowlist?';
			case 'ui.turnOffAllowListDescription': return 'Disabling the allowlist will delete all your previously added ports, port ranges, and subnets.';
			case 'ui.port': return 'Port';
			case 'ui.portRange': return 'Port range';
			case 'ui.subnet': return 'Subnet';
			case 'ui.enterPort': return 'Enter port';
			case 'ui.selectProtocol': return 'Select protocol';
			case 'ui.all': return 'All';
			case 'ui.protocol': return 'Protocol';
			case 'ui.portAlreadyInList': return 'Port is already on the list';
			case 'ui.portRangeAlreadyInList': return 'The range is already on the list';
			case 'ui.enterPortRange': return 'Enter port range';
			case 'ui.enterSubnet': return 'Enter subnet';
			case 'ui.subnetAlreadyInList': return 'Subnet is already on the list';
			case 'ui.delete': return 'Delete';
			case 'ui.settingsWereNotSaved': return 'Settings weren\'t saved';
			case 'ui.couldNotSave': return 'We couldn\'t save your settings to the configuration file.';
			case 'ui.turnOffObfuscationServerTypes': return 'Turn off obfuscation for more server types';
			case 'ui.turnOffObfuscationLocations': return 'Turn off obfuscation for more locations';
			case 'ui.nordWhisper': return 'NordWhisper';
			case 'ui.system': return 'System';
			case 'ui.removePrivateSubnets': return 'We\'ll remove private subnets from allowlist';
			case 'ui.removePrivateSubnetsDescription': return 'Enabling LAN discovery will remove any private subnets from allowlist. Continue?';
			case 'ui.privateSubnetCantBeAdded': return 'Private subnet can\'t be added';
			case 'ui.privateSubnetCantBeAddedDescription': return 'Allowlisting a private subnet isn’t available while local network discovery is enabled. To add a private subnet, turn off LAN discovery.';
			case 'ui.turnOffLanDiscovery': return 'Turn off LAN discovery';
			case 'ui.startPortBiggerThanEnd': return 'Start port can’t be greater than end port';
			case 'ui.weCouldNotConnectToService': return 'We couldn’t connect to NordVPN service';
			case 'ui.needHelp': return ({required Uri supportUrl}) => 'Need help? [Visit our Support Center](${supportUrl}) ';
			case 'ui.issuePersists': return ({required Uri supportUrl}) => 'Issue persists? [Contact our customer support](${supportUrl})';
			case 'ui.meshnet': return 'Meshnet';
			case 'ui.systemdDistribution': return 'systemd distribution';
			case 'ui.nonSystemdDistro': return 'non-systemd distribution';
			case 'ui.tryRunningOneCommand': return 'The service isn’t running. To start it, use this command in the terminal.';
			case 'ui.failedToFetchConsentData': return 'Something unexpected happened while we were trying to load your analytics consent setting.';
			case 'ui.failedToFetchAccountData': return 'Something unexpected happened while we were trying to load your account data.';
			case 'ui.tryAgain': return 'Try again';
			case 'ui.weHitAnError': return 'We hit an error';
			case 'ui.weValueYourPrivacy': return 'We value your privacy';
			case 'ui.consentDescription': return ({required Uri privacyUrl}) => 'That’s why we want to be transparent about what data you agree to give us. We only collect the bare minimum of information required to offer a smooth and stable VPN experience.\nYour browsing activities remain private, regardless of your choice.\n\nBy selecting “Accept,” you allow us to collect and use limited app performance data for analytics, as explained in our [Privacy Policy](${privacyUrl}).\n\nSelect “Customize” to manage your privacy choices or learn more about each option.';
			case 'ui.accept': return 'Accept';
			case 'ui.customize': return 'Customize';
			case 'ui.rejectNonEssential': return 'Reject non-essential';
			case 'ui.privacyPreferences': return 'Privacy preferences';
			case 'ui.privacyPreferencesDescription': return 'Customize your privacy choices';
			case 'ui.essentialRequired': return 'Essential (Required)';
			case 'ui.requiredAnalyticsDescription': return ({required Uri termsUrl}) => 'Send us the essential data we need for our app to work. This also helps enforce our [Terms of Service](${termsUrl}), prevent fraud, and maintain the security of our services. ';
			case 'ui.confirmPreferences': return 'Confirm preferences';
			case 'ui.analytics': return 'Analytics';
			case 'ui.analyticsDescription': return 'Help us improve the app by sending aggregated usage data. We collect this data to keep our features relevant to your needs and to fix issues faster.';
			case 'ui.back': return 'Back';
			case 'ui.snapScreenTitle': return 'We need some permissions to connect to NordVPN service';
			case 'ui.snapScreenDescription': return 'Grant permission by running these commands in the terminal. Then refresh the screen.';
			case 'ui.refresh': return 'Refresh';
			default: return null;
		}
	}
}

