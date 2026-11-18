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
	late final Translations$a11y$en a11y = Translations$a11y$en._(_root);
	late final Translations$cities$en cities = Translations$cities$en._(_root);
	late final Translations$countries$en countries = Translations$countries$en._(_root);
	late final Translations$daemon$en daemon = Translations$daemon$en._(_root);
	late final Translations$ui$en ui = Translations$ui$en._(_root);
}

// Path: a11y
class Translations$a11y$en {
	Translations$a11y$en._(this._root);

	final Translations _root; // ignore: unused_field

	// Translations

	/// en: 'VPN connection resumes in $minutes minutes $seconds seconds'
	String VPNResumesIn({required Object minutes, required Object seconds}) => 'VPN connection resumes in ${minutes} minutes ${seconds} seconds';

	/// en: 'VPN connection resumes in $hours hours $minutes minutes $seconds seconds'
	String VPNResumesInWithHours({required Object hours, required Object minutes, required Object seconds}) => 'VPN connection resumes in ${hours} hours ${minutes} minutes ${seconds} seconds';

	/// en: 'VPN Panel. Preferred location $location. Not secured.'
	String VPNPanelDisconnected({required Object location}) => 'VPN Panel. Preferred location ${location}. Not secured.';

	/// en: 'VPN Panel. Connecting to $location.'
	String VPNPanelConnecting({required Object location}) => 'VPN Panel. Connecting to ${location}.';

	/// en: 'VPN Panel. Connected to $location.'
	String VPNPanelConnected({required Object location}) => 'VPN Panel. Connected to ${location}.';

	/// en: 'VPN Panel. Loading.'
	String get VPNPanelLoading => 'VPN Panel. Loading.';

	/// en: 'More options'
	String get moreOptions => 'More options';

	/// en: 'List with $number items'
	String listWithItems({required Object number}) => 'List with ${number} items';

	/// en: 'Recent connections. $details'
	String recentConnectionWithDetails({required Object details}) => 'Recent connections. ${details}';

	/// en: 'Expand'
	String get expandibleEntryExpand => 'Expand';

	/// en: 'Collapse'
	String get expandibleEntryCollapse => 'Collapse';

	/// en: 'Expanded'
	String get expandibleEntryExpanded => 'Expanded';

	/// en: 'Collapsed'
	String get expandibleEntryCollapsed => 'Collapsed';

	/// en: 'Clear'
	String get clear => 'Clear';

	/// en: '$title. $message'
	String popupWithContent({required Object title, required Object message}) => '${title}. ${message}';

	/// en: 'Link $name'
	String link({required Object name}) => 'Link ${name}';

	/// en: 'Selected'
	String get selected => 'Selected';

	/// en: 'Not selected'
	String get notSelected => 'Not selected';

	/// en: 'NordVPN version $version'
	String appVersion({required Object version}) => 'NordVPN version ${version}';

	/// en: 'Warning. $message'
	String warningMessage({required Object message}) => 'Warning. ${message}';

	/// en: 'Start port'
	String get startPort => 'Start port';

	/// en: 'End port'
	String get endPort => 'End port';
}

// Path: cities
class Translations$cities$en {
	Translations$cities$en._(this._root);

	final Translations _root; // ignore: unused_field

	// Translations

	/// en: 'Accra'
	String get accra => 'Accra';

	/// en: 'Addis Ababa'
	String get addis_ababa => 'Addis Ababa';

	/// en: 'Adelaide'
	String get adelaide => 'Adelaide';

	/// en: 'Albuquerque'
	String get albuquerque => 'Albuquerque';

	/// en: 'Algiers'
	String get algiers => 'Algiers';

	/// en: 'Amman'
	String get amman => 'Amman';

	/// en: 'Amsterdam'
	String get amsterdam => 'Amsterdam';

	/// en: 'Anchorage'
	String get anchorage => 'Anchorage';

	/// en: 'Andorra la Vella'
	String get andorra_la_vella => 'Andorra la Vella';

	/// en: 'Antananarivo'
	String get antananarivo => 'Antananarivo';

	/// en: 'Ashburn'
	String get ashburn => 'Ashburn';

	/// en: 'Astana'
	String get astana => 'Astana';

	/// en: 'Asuncion'
	String get asuncion => 'Asuncion';

	/// en: 'Athens'
	String get athens => 'Athens';

	/// en: 'Atlanta'
	String get atlanta => 'Atlanta';

	/// en: 'Auckland'
	String get auckland => 'Auckland';

	/// en: 'Baghdad'
	String get baghdad => 'Baghdad';

	/// en: 'Baku'
	String get baku => 'Baku';

	/// en: 'Baltimore'
	String get baltimore => 'Baltimore';

	/// en: 'Bandar Seri Begawan'
	String get bandar_seri_begawan => 'Bandar Seri Begawan';

	/// en: 'Bangkok'
	String get bangkok => 'Bangkok';

	/// en: 'Banjul'
	String get banjul => 'Banjul';

	/// en: 'Barcelona'
	String get barcelona => 'Barcelona';

	/// en: 'Beirut'
	String get beirut => 'Beirut';

	/// en: 'Belgrade'
	String get belgrade => 'Belgrade';

	/// en: 'Belmopan'
	String get belmopan => 'Belmopan';

	/// en: 'Berlin'
	String get berlin => 'Berlin';

	/// en: 'Billings'
	String get billings => 'Billings';

	/// en: 'Bishkek'
	String get bishkek => 'Bishkek';

	/// en: 'Bogota'
	String get bogota => 'Bogota';

	/// en: 'Boise'
	String get boise => 'Boise';

	/// en: 'Bordeaux'
	String get bordeaux => 'Bordeaux';

	/// en: 'Boston'
	String get boston => 'Boston';

	/// en: 'Bratislava'
	String get bratislava => 'Bratislava';

	/// en: 'Bridgetown'
	String get bridgetown => 'Bridgetown';

	/// en: 'Brisbane'
	String get brisbane => 'Brisbane';

	/// en: 'Brussels'
	String get brussels => 'Brussels';

	/// en: 'Bucharest'
	String get bucharest => 'Bucharest';

	/// en: 'Budapest'
	String get budapest => 'Budapest';

	/// en: 'Buenos Aires'
	String get buenos_aires => 'Buenos Aires';

	/// en: 'Buffalo'
	String get buffalo => 'Buffalo';

	/// en: 'Burlington'
	String get burlington => 'Burlington';

	/// en: 'Cairo'
	String get cairo => 'Cairo';

	/// en: 'Calgary'
	String get calgary => 'Calgary';

	/// en: 'Caracas'
	String get caracas => 'Caracas';

	/// en: 'Castries'
	String get castries => 'Castries';

	/// en: 'Charleston'
	String get charleston => 'Charleston';

	/// en: 'Charlotte'
	String get charlotte => 'Charlotte';

	/// en: 'Cheyenne'
	String get cheyenne => 'Cheyenne';

	/// en: 'Chicago'
	String get chicago => 'Chicago';

	/// en: 'Chisinau'
	String get chisinau => 'Chisinau';

	/// en: 'Colombo'
	String get colombo => 'Colombo';

	/// en: 'Columbus'
	String get columbus => 'Columbus';

	/// en: 'Copenhagen'
	String get copenhagen => 'Copenhagen';

	/// en: 'Dakar'
	String get dakar => 'Dakar';

	/// en: 'Dallas'
	String get dallas => 'Dallas';

	/// en: 'Denver'
	String get denver => 'Denver';

	/// en: 'Des Moines'
	String get des_moines => 'Des Moines';

	/// en: 'Detroit'
	String get detroit => 'Detroit';

	/// en: 'Dhaka'
	String get dhaka => 'Dhaka';

	/// en: 'Dodoma'
	String get dodoma => 'Dodoma';

	/// en: 'Doha'
	String get doha => 'Doha';

	/// en: 'Douglas'
	String get douglas => 'Douglas';

	/// en: 'Dubai'
	String get dubai => 'Dubai';

	/// en: 'Dublin'
	String get dublin => 'Dublin';

	/// en: 'Dushanbe'
	String get dushanbe => 'Dushanbe';

	/// en: 'Edinburgh'
	String get edinburgh => 'Edinburgh';

	/// en: 'Fargo'
	String get fargo => 'Fargo';

	/// en: 'Frankfurt'
	String get frankfurt => 'Frankfurt';

	/// en: 'Freetown'
	String get freetown => 'Freetown';

	/// en: 'Fujairah'
	String get fujairah => 'Fujairah';

	/// en: 'Gaborone'
	String get gaborone => 'Gaborone';

	/// en: 'George Town'
	String get george_town => 'George Town';

	/// en: 'Georgetown'
	String get georgetown => 'Georgetown';

	/// en: 'Glasgow'
	String get glasgow => 'Glasgow';

	/// en: 'Guatemala City'
	String get guatemala_city => 'Guatemala City';

	/// en: 'Hagatna'
	String get hagatna => 'Hagatna';

	/// en: 'Hamburg'
	String get hamburg => 'Hamburg';

	/// en: 'Hamilton'
	String get hamilton => 'Hamilton';

	/// en: 'Hanoi'
	String get hanoi => 'Hanoi';

	/// en: 'Helsinki'
	String get helsinki => 'Helsinki';

	/// en: 'Ho Chi Minh City'
	String get ho_chi_minh_city => 'Ho Chi Minh City';

	/// en: 'Hong Kong'
	String get hong_kong => 'Hong Kong';

	/// en: 'Honolulu'
	String get honolulu => 'Honolulu';

	/// en: 'Houston'
	String get houston => 'Houston';

	/// en: 'Huntington'
	String get huntington => 'Huntington';

	/// en: 'Indianapolis'
	String get indianapolis => 'Indianapolis';

	/// en: 'Istanbul'
	String get istanbul => 'Istanbul';

	/// en: 'Jackson'
	String get jackson => 'Jackson';

	/// en: 'Jakarta'
	String get jakarta => 'Jakarta';

	/// en: 'Johannesburg'
	String get johannesburg => 'Johannesburg';

	/// en: 'Kabul'
	String get kabul => 'Kabul';

	/// en: 'Kansas City'
	String get kansas_city => 'Kansas City';

	/// en: 'Karachi'
	String get karachi => 'Karachi';

	/// en: 'Kathmandu'
	String get kathmandu => 'Kathmandu';

	/// en: 'Kingston'
	String get kingston => 'Kingston';

	/// en: 'Kuala Lumpur'
	String get kuala_lumpur => 'Kuala Lumpur';

	/// en: 'Kuwait City'
	String get kuwait_city => 'Kuwait City';

	/// en: 'Kyiv'
	String get kyiv => 'Kyiv';

	/// en: 'La Paz'
	String get la_paz => 'La Paz';

	/// en: 'Lagos'
	String get lagos => 'Lagos';

	/// en: 'Las Vegas'
	String get las_vegas => 'Las Vegas';

	/// en: 'Lewiston'
	String get lewiston => 'Lewiston';

	/// en: 'Lilongwe'
	String get lilongwe => 'Lilongwe';

	/// en: 'Lima'
	String get lima => 'Lima';

	/// en: 'Lisbon'
	String get lisbon => 'Lisbon';

	/// en: 'Little Rock'
	String get little_rock => 'Little Rock';

	/// en: 'Ljubljana'
	String get ljubljana => 'Ljubljana';

	/// en: 'London'
	String get london => 'London';

	/// en: 'Los Angeles'
	String get los_angeles => 'Los Angeles';

	/// en: 'Louisville'
	String get louisville => 'Louisville';

	/// en: 'Luanda'
	String get luanda => 'Luanda';

	/// en: 'Lusaka'
	String get lusaka => 'Lusaka';

	/// en: 'Luxembourg'
	String get luxembourg => 'Luxembourg';

	/// en: 'Madrid'
	String get madrid => 'Madrid';

	/// en: 'Male'
	String get male => 'Male';

	/// en: 'Manama'
	String get manama => 'Manama';

	/// en: 'Manchester'
	String get manchester => 'Manchester';

	/// en: 'Manila'
	String get manila => 'Manila';

	/// en: 'Maputo'
	String get maputo => 'Maputo';

	/// en: 'Marseille'
	String get marseille => 'Marseille';

	/// en: 'McAllen'
	String get mcallen => 'McAllen';

	/// en: 'Melbourne'
	String get melbourne => 'Melbourne';

	/// en: 'Mexico'
	String get mexico => 'Mexico';

	/// en: 'Miami'
	String get miami => 'Miami';

	/// en: 'Milan'
	String get milan => 'Milan';

	/// en: 'Milwaukee'
	String get milwaukee => 'Milwaukee';

	/// en: 'Minneapolis'
	String get minneapolis => 'Minneapolis';

	/// en: 'Mogadishu'
	String get mogadishu => 'Mogadishu';

	/// en: 'Monrovia'
	String get monrovia => 'Monrovia';

	/// en: 'Monte Carlo'
	String get monte_carlo => 'Monte Carlo';

	/// en: 'Montevideo'
	String get montevideo => 'Montevideo';

	/// en: 'Montgomery'
	String get montgomery => 'Montgomery';

	/// en: 'Montreal'
	String get montreal => 'Montreal';

	/// en: 'Moroni'
	String get moroni => 'Moroni';

	/// en: 'Mumbai'
	String get mumbai => 'Mumbai';

	/// en: 'Nairobi'
	String get nairobi => 'Nairobi';

	/// en: 'Nashua'
	String get nashua => 'Nashua';

	/// en: 'Nashville'
	String get nashville => 'Nashville';

	/// en: 'Nassau'
	String get nassau => 'Nassau';

	/// en: 'Naypyidaw'
	String get naypyidaw => 'Naypyidaw';

	/// en: 'NDjamena'
	String get ndjamena => 'NDjamena';

	/// en: 'New Haven'
	String get new_haven => 'New Haven';

	/// en: 'New Orleans'
	String get new_orleans => 'New Orleans';

	/// en: 'New York'
	String get new_york => 'New York';

	/// en: 'Nicosia'
	String get nicosia => 'Nicosia';

	/// en: 'Nouakchott'
	String get nouakchott => 'Nouakchott';

	/// en: 'Nuuk'
	String get nuuk => 'Nuuk';

	/// en: 'Oklahoma City'
	String get oklahoma_city => 'Oklahoma City';

	/// en: 'Omaha'
	String get omaha => 'Omaha';

	/// en: 'Osaka'
	String get osaka => 'Osaka';

	/// en: 'Oslo'
	String get oslo => 'Oslo';

	/// en: 'Ouagadougou'
	String get ouagadougou => 'Ouagadougou';

	/// en: 'Palermo'
	String get palermo => 'Palermo';

	/// en: 'Panama City'
	String get panama_city => 'Panama City';

	/// en: 'Paramaribo'
	String get paramaribo => 'Paramaribo';

	/// en: 'Paris'
	String get paris => 'Paris';

	/// en: 'Perth'
	String get perth => 'Perth';

	/// en: 'Phnom Penh'
	String get phnom_penh => 'Phnom Penh';

	/// en: 'Phoenix'
	String get phoenix => 'Phoenix';

	/// en: 'Pittsburgh'
	String get pittsburgh => 'Pittsburgh';

	/// en: 'Podgorica'
	String get podgorica => 'Podgorica';

	/// en: 'Port Louis'
	String get port_louis => 'Port Louis';

	/// en: 'Port Moresby'
	String get port_moresby => 'Port Moresby';

	/// en: 'Port of Spain'
	String get port_of_spain => 'Port of Spain';

	/// en: 'Portland'
	String get portland => 'Portland';

	/// en: 'Porto Novo'
	String get porto_novo => 'Porto Novo';

	/// en: 'Prague'
	String get prague => 'Prague';

	/// en: 'Praia'
	String get praia => 'Praia';

	/// en: 'Providence'
	String get providence => 'Providence';

	/// en: 'Quito'
	String get quito => 'Quito';

	/// en: 'Rabat'
	String get rabat => 'Rabat';

	/// en: 'Reykjavik'
	String get reykjavik => 'Reykjavik';

	/// en: 'Riga'
	String get riga => 'Riga';

	/// en: 'Rome'
	String get rome => 'Rome';

	/// en: 'Saint Helier'
	String get saint_helier => 'Saint Helier';

	/// en: 'Saint John'
	String get saint_john => 'Saint John';

	/// en: 'Saint Louis'
	String get saint_louis => 'Saint Louis';

	/// en: 'Salt Lake City'
	String get salt_lake_city => 'Salt Lake City';

	/// en: 'San Francisco'
	String get san_francisco => 'San Francisco';

	/// en: 'San Jose'
	String get san_jose => 'San Jose';

	/// en: 'San Juan'
	String get san_juan => 'San Juan';

	/// en: 'San Salvador'
	String get san_salvador => 'San Salvador';

	/// en: 'Sanaa'
	String get sanaa => 'Sanaa';

	/// en: 'Santiago'
	String get santiago => 'Santiago';

	/// en: 'Santo Domingo'
	String get santo_domingo => 'Santo Domingo';

	/// en: 'Sao Paulo'
	String get sao_paulo => 'Sao Paulo';

	/// en: 'Sarajevo'
	String get sarajevo => 'Sarajevo';

	/// en: 'Seattle'
	String get seattle => 'Seattle';

	/// en: 'Seoul'
	String get seoul => 'Seoul';

	/// en: 'Singapore'
	String get singapore => 'Singapore';

	/// en: 'Sioux Falls'
	String get sioux_falls => 'Sioux Falls';

	/// en: 'Skopje'
	String get skopje => 'Skopje';

	/// en: 'Sofia'
	String get sofia => 'Sofia';

	/// en: 'Stockholm'
	String get stockholm => 'Stockholm';

	/// en: 'Strasbourg'
	String get strasbourg => 'Strasbourg';

	/// en: 'Suva'
	String get suva => 'Suva';

	/// en: 'Sydney'
	String get sydney => 'Sydney';

	/// en: 'Taipei'
	String get taipei => 'Taipei';

	/// en: 'Tallinn'
	String get tallinn => 'Tallinn';

	/// en: 'Tashkent'
	String get tashkent => 'Tashkent';

	/// en: 'Tbilisi'
	String get tbilisi => 'Tbilisi';

	/// en: 'Tegucigalpa'
	String get tegucigalpa => 'Tegucigalpa';

	/// en: 'Tel Aviv'
	String get tel_aviv => 'Tel Aviv';

	/// en: 'Thimphu'
	String get thimphu => 'Thimphu';

	/// en: 'Tirana'
	String get tirana => 'Tirana';

	/// en: 'Tokyo'
	String get tokyo => 'Tokyo';

	/// en: 'Toronto'
	String get toronto => 'Toronto';

	/// en: 'Trenton'
	String get trenton => 'Trenton';

	/// en: 'Tripoli'
	String get tripoli => 'Tripoli';

	/// en: 'Tunis'
	String get tunis => 'Tunis';

	/// en: 'Ulaanbaatar'
	String get ulaanbaatar => 'Ulaanbaatar';

	/// en: 'Vaduz'
	String get vaduz => 'Vaduz';

	/// en: 'Valletta'
	String get valletta => 'Valletta';

	/// en: 'Vancouver'
	String get vancouver => 'Vancouver';

	/// en: 'Victoria'
	String get victoria => 'Victoria';

	/// en: 'Vienna'
	String get vienna => 'Vienna';

	/// en: 'Vientiane'
	String get vientiane => 'Vientiane';

	/// en: 'Vilnius'
	String get vilnius => 'Vilnius';

	/// en: 'Warsaw'
	String get warsaw => 'Warsaw';

	/// en: 'Wichita'
	String get wichita => 'Wichita';

	/// en: 'Wilmington'
	String get wilmington => 'Wilmington';

	/// en: 'Yerevan'
	String get yerevan => 'Yerevan';

	/// en: 'Zagreb'
	String get zagreb => 'Zagreb';

	/// en: 'Zurich'
	String get zurich => 'Zurich';
}

// Path: countries
class Translations$countries$en {
	Translations$countries$en._(this._root);

	final Translations _root; // ignore: unused_field

	// Translations

	/// en: 'Andorra'
	String get AD => 'Andorra';

	/// en: 'United Arab Emirates'
	String get AE => 'United Arab Emirates';

	/// en: 'Afghanistan'
	String get AF => 'Afghanistan';

	/// en: 'Antigua and Barbuda'
	String get AG => 'Antigua and Barbuda';

	/// en: 'Albania'
	String get AL => 'Albania';

	/// en: 'Armenia'
	String get AM => 'Armenia';

	/// en: 'Angola'
	String get AO => 'Angola';

	/// en: 'Argentina'
	String get AR => 'Argentina';

	/// en: 'Austria'
	String get AT => 'Austria';

	/// en: 'Australia'
	String get AU => 'Australia';

	/// en: 'Azerbaijan'
	String get AZ => 'Azerbaijan';

	/// en: 'Bosnia and Herzegovina'
	String get BA => 'Bosnia and Herzegovina';

	/// en: 'Barbados'
	String get BB => 'Barbados';

	/// en: 'Bangladesh'
	String get BD => 'Bangladesh';

	/// en: 'Belgium'
	String get BE => 'Belgium';

	/// en: 'Burkina Faso'
	String get BF => 'Burkina Faso';

	/// en: 'Bulgaria'
	String get BG => 'Bulgaria';

	/// en: 'Bahrain'
	String get BH => 'Bahrain';

	/// en: 'Benin'
	String get BJ => 'Benin';

	/// en: 'Bermuda'
	String get BM => 'Bermuda';

	/// en: 'Brunei Darussalam'
	String get BN => 'Brunei Darussalam';

	/// en: 'Bolivia'
	String get BO => 'Bolivia';

	/// en: 'Brazil'
	String get BR => 'Brazil';

	/// en: 'Bahamas'
	String get BS => 'Bahamas';

	/// en: 'Bhutan'
	String get BT => 'Bhutan';

	/// en: 'Botswana'
	String get BW => 'Botswana';

	/// en: 'Belize'
	String get BZ => 'Belize';

	/// en: 'Canada'
	String get CA => 'Canada';

	/// en: 'Switzerland'
	String get CH => 'Switzerland';

	/// en: 'Chile'
	String get CL => 'Chile';

	/// en: 'Colombia'
	String get CO => 'Colombia';

	/// en: 'Costa Rica'
	String get CR => 'Costa Rica';

	/// en: 'Cape Verde'
	String get CV => 'Cape Verde';

	/// en: 'Cyprus'
	String get CY => 'Cyprus';

	/// en: 'Czech Republic'
	String get CZ => 'Czech Republic';

	/// en: 'Germany'
	String get DE => 'Germany';

	/// en: 'Denmark'
	String get DK => 'Denmark';

	/// en: 'Dominican Republic'
	String get DO => 'Dominican Republic';

	/// en: 'Algeria'
	String get DZ => 'Algeria';

	/// en: 'Ecuador'
	String get EC => 'Ecuador';

	/// en: 'Estonia'
	String get EE => 'Estonia';

	/// en: 'Egypt'
	String get EG => 'Egypt';

	/// en: 'Spain'
	String get ES => 'Spain';

	/// en: 'Ethiopia'
	String get ET => 'Ethiopia';

	/// en: 'Finland'
	String get FI => 'Finland';

	/// en: 'Fiji'
	String get FJ => 'Fiji';

	/// en: 'France'
	String get FR => 'France';

	/// en: 'United Kingdom'
	String get GB => 'United Kingdom';

	/// en: 'Georgia'
	String get GE => 'Georgia';

	/// en: 'Ghana'
	String get GH => 'Ghana';

	/// en: 'Greenland'
	String get GL => 'Greenland';

	/// en: 'Gambia'
	String get GM => 'Gambia';

	/// en: 'Greece'
	String get GR => 'Greece';

	/// en: 'Guatemala'
	String get GT => 'Guatemala';

	/// en: 'Guam'
	String get GU => 'Guam';

	/// en: 'Guyana'
	String get GY => 'Guyana';

	/// en: 'Hong Kong'
	String get HK => 'Hong Kong';

	/// en: 'Honduras'
	String get HN => 'Honduras';

	/// en: 'Croatia'
	String get HR => 'Croatia';

	/// en: 'Hungary'
	String get HU => 'Hungary';

	/// en: 'Indonesia'
	String get ID => 'Indonesia';

	/// en: 'Ireland'
	String get IE => 'Ireland';

	/// en: 'Israel'
	String get IL => 'Israel';

	/// en: 'Isle of Man'
	String get IM => 'Isle of Man';

	/// en: 'India'
	String get IN => 'India';

	/// en: 'Iraq'
	String get IQ => 'Iraq';

	/// en: 'Iceland'
	String get IS => 'Iceland';

	/// en: 'Italy'
	String get IT => 'Italy';

	/// en: 'Jersey'
	String get JE => 'Jersey';

	/// en: 'Jamaica'
	String get JM => 'Jamaica';

	/// en: 'Jordan'
	String get JO => 'Jordan';

	/// en: 'Japan'
	String get JP => 'Japan';

	/// en: 'Kenya'
	String get KE => 'Kenya';

	/// en: 'Kyrgyzstan'
	String get KG => 'Kyrgyzstan';

	/// en: 'Cambodia'
	String get KH => 'Cambodia';

	/// en: 'Comoros'
	String get KM => 'Comoros';

	/// en: 'South Korea'
	String get KR => 'South Korea';

	/// en: 'Kuwait'
	String get KW => 'Kuwait';

	/// en: 'Cayman Islands'
	String get KY => 'Cayman Islands';

	/// en: 'Kazakhstan'
	String get KZ => 'Kazakhstan';

	/// en: 'Lao People's Democratic Republic'
	String get LA => 'Lao People\'s Democratic Republic';

	/// en: 'Lebanon'
	String get LB => 'Lebanon';

	/// en: 'Saint Lucia'
	String get LC => 'Saint Lucia';

	/// en: 'Liechtenstein'
	String get LI => 'Liechtenstein';

	/// en: 'Sri Lanka'
	String get LK => 'Sri Lanka';

	/// en: 'Liberia'
	String get LR => 'Liberia';

	/// en: 'Lithuania'
	String get LT => 'Lithuania';

	/// en: 'Luxembourg'
	String get LU => 'Luxembourg';

	/// en: 'Latvia'
	String get LV => 'Latvia';

	/// en: 'Libyan Arab Jamahiriya'
	String get LY => 'Libyan Arab Jamahiriya';

	/// en: 'Morocco'
	String get MA => 'Morocco';

	/// en: 'Monaco'
	String get MC => 'Monaco';

	/// en: 'Moldova'
	String get MD => 'Moldova';

	/// en: 'Montenegro'
	String get ME => 'Montenegro';

	/// en: 'Madagascar'
	String get MG => 'Madagascar';

	/// en: 'North Macedonia'
	String get MK => 'North Macedonia';

	/// en: 'Myanmar'
	String get MM => 'Myanmar';

	/// en: 'Mongolia'
	String get MN => 'Mongolia';

	/// en: 'Mauritania'
	String get MR => 'Mauritania';

	/// en: 'Malta'
	String get MT => 'Malta';

	/// en: 'Mauritius'
	String get MU => 'Mauritius';

	/// en: 'Maldives'
	String get MV => 'Maldives';

	/// en: 'Malawi'
	String get MW => 'Malawi';

	/// en: 'Mexico'
	String get MX => 'Mexico';

	/// en: 'Malaysia'
	String get MY => 'Malaysia';

	/// en: 'Mozambique'
	String get MZ => 'Mozambique';

	/// en: 'Nigeria'
	String get NG => 'Nigeria';

	/// en: 'Netherlands'
	String get NL => 'Netherlands';

	/// en: 'Norway'
	String get NO => 'Norway';

	/// en: 'Nepal'
	String get NP => 'Nepal';

	/// en: 'New Zealand'
	String get NZ => 'New Zealand';

	/// en: 'Panama'
	String get PA => 'Panama';

	/// en: 'Peru'
	String get PE => 'Peru';

	/// en: 'Papua New Guinea'
	String get PG => 'Papua New Guinea';

	/// en: 'Philippines'
	String get PH => 'Philippines';

	/// en: 'Pakistan'
	String get PK => 'Pakistan';

	/// en: 'Poland'
	String get PL => 'Poland';

	/// en: 'Puerto Rico'
	String get PR => 'Puerto Rico';

	/// en: 'Portugal'
	String get PT => 'Portugal';

	/// en: 'Paraguay'
	String get PY => 'Paraguay';

	/// en: 'Qatar'
	String get QA => 'Qatar';

	/// en: 'Romania'
	String get RO => 'Romania';

	/// en: 'Serbia'
	String get RS => 'Serbia';

	/// en: 'Seychelles'
	String get SC => 'Seychelles';

	/// en: 'Sweden'
	String get SE => 'Sweden';

	/// en: 'Singapore'
	String get SG => 'Singapore';

	/// en: 'Slovenia'
	String get SI => 'Slovenia';

	/// en: 'Slovakia'
	String get SK => 'Slovakia';

	/// en: 'Sierra Leone'
	String get SL => 'Sierra Leone';

	/// en: 'Senegal'
	String get SN => 'Senegal';

	/// en: 'Somalia'
	String get SO => 'Somalia';

	/// en: 'Suriname'
	String get SR => 'Suriname';

	/// en: 'El Salvador'
	String get SV => 'El Salvador';

	/// en: 'Chad'
	String get TD => 'Chad';

	/// en: 'Thailand'
	String get TH => 'Thailand';

	/// en: 'Tajikistan'
	String get TJ => 'Tajikistan';

	/// en: 'Tunisia'
	String get TN => 'Tunisia';

	/// en: 'Turkey'
	String get TR => 'Turkey';

	/// en: 'Trinidad and Tobago'
	String get TT => 'Trinidad and Tobago';

	/// en: 'Taiwan'
	String get TW => 'Taiwan';

	/// en: 'Tanzania'
	String get TZ => 'Tanzania';

	/// en: 'Ukraine'
	String get UA => 'Ukraine';

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

	/// en: 'Yemen'
	String get YE => 'Yemen';

	/// en: 'South Africa'
	String get ZA => 'South Africa';

	/// en: 'Zambia'
	String get ZM => 'Zambia';
}

// Path: daemon
class Translations$daemon$en {
	Translations$daemon$en._(this._root);

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

	/// en: 'Disabled protocol'
	String get code_3051_title => 'Disabled protocol';

	/// en: 'Unable to connect with the current protocol. Please try a different one using the command: nordvpn set protocol.'
	String get code_3051_msg => 'Unable to connect with the current protocol. Please try a different one using the command: nordvpn set protocol.';

	/// en: 'Here's what to know'
	String get code_3068_title => 'Here\'s what to know';

	/// en: 'Auto-connect is set to a dedicated server, which only works with NordLynx. To use a different protocol, choose another auto-connect location or turn off auto-connect first.'
	String get code_3068_msg => 'Auto-connect is set to a dedicated server, which only works with NordLynx. To use a different protocol, choose another auto-connect location or turn off auto-connect first.';

	/// en: 'Here's what to know'
	String get code_3080_title => 'Here\'s what to know';

	/// en: 'Auto-connect is set to obfuscated servers, which only work with NordWhisper. To use a different protocol, choose another auto-connect location or turn off auto-connect first.'
	String get code_3080_msg => 'Auto-connect is set to obfuscated servers, which only work with NordWhisper. To use a different protocol, choose another auto-connect location or turn off auto-connect first.';

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
class Translations$ui$en {
	Translations$ui$en._(this._root);

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

	/// en: 'No results found for ${searchStr: String}'
	String noResultsFor({required String searchStr}) => 'No results found for ${searchStr}';

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

	/// en: 'Fatal error'
	String get fatalErrorMessage => 'Fatal error';

	/// en: 'Connected'
	String get connected => 'Connected';

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

	/// en: 'Real-time protection'
	String get realTimeProtection => 'Real-time protection';

	/// en: 'Stop online scams and phishing attempts'
	String get realTimeProtectionSubtitle => 'Stop online scams and phishing attempts';

	/// en: 'Stops online scams and phishing attempts while you’re connected to the VPN.'
	String get realTimeProtectionDescription => 'Stops online scams and phishing attempts while you’re connected to the VPN.';

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

	/// en: 'Dedicated IP'
	String get dedicatedIp => 'Dedicated IP';

	/// en: 'Dedicated Server'
	String get dedicatedServer => 'Dedicated Server';

	/// en: 'Double VPN'
	String get doubleVpn => 'Double VPN';

	/// en: 'Onion Over VPN'
	String get onionOverVpn => 'Onion Over VPN';

	/// en: 'Obfuscated'
	String get obfuscated => 'Obfuscated';

	/// en: 'Obfuscated'
	String get obfuscatedServers => 'Obfuscated';

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

	/// en: 'Allowlist, DNS, LAN discovery, firewall'
	String get securityAndPrivacySubtitle => 'Allowlist, DNS, LAN discovery, firewall';

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

	/// en: 'Fastest'
	String get fastest => 'Fastest';

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

	/// en: 'Turn off custom DNS?'
	String get realTimeDisableCustomDNS => 'Turn off custom DNS?';

	/// en: 'Real-time protection works with our default DNS servers only.'
	String get realTimeDisableCustomDNSDescription => 'Real-time protection works with our default DNS servers only.';

	/// en: 'Continue'
	String get continueWord => 'Continue';

	/// en: 'Real-time protection will be turned off'
	String get realTimeProtectionWillTurnOff => 'Real-time protection will be turned off';

	/// en: 'Real-time protection works only with the default DNS. Set a custom DNS server anyway?'
	String get realTimeProtectionWillTurnOffDescription => 'Real-time protection works only with the default DNS. Set a custom DNS server anyway?';

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

	/// en: 'Setting auto-connect to [$target]...'
	String settingAutoconnectTo({required Object target}) => 'Setting auto-connect to [${target}]...';

	/// en: 'Encrypt your traffic twice for extra security'
	String get doubleVpnDesc => 'Encrypt your traffic twice for extra security';

	/// en: 'Use the Onion network with VPN protection'
	String get onionOverVpnDesc => 'Use the Onion network with VPN protection';

	/// en: 'Hide signs of VPN usage on restricted networks'
	String get obfuscatedServersDesc => 'Hide signs of VPN usage on restricted networks';

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

	/// en: 'Some settings weren't applied'
	String get settingsNotApplied => 'Some settings weren\'t applied';

	/// en: 'Some default network settings could not be applied. Reset settings again to retry. If the problem continues, contact support.'
	String get networkSettingsNotAppliedDescription => 'Some default network settings could not be applied. Reset settings again to retry. If the problem continues, contact support.';

	/// en: 'Some default settings could not be applied. The reset finished, but these settings may not be active: $settings. Reset settings again to retry. If the problem continues, contact support.'
	String settingsNotAppliedDescription({required Object settings}) => 'Some default settings could not be applied. The reset finished, but these settings may not be active: ${settings}. Reset settings again to retry. If the problem continues, contact support.';

	/// en: 'NordWhisper'
	String get nordWhisper => 'NordWhisper';

	/// en: 'System'
	String get system => 'System';

	/// en: 'We'll remove private subnets from allowlist'
	String get removePrivateSubnets => 'We\'ll remove private subnets from allowlist';

	/// en: 'Enabling LAN discovery will remove any private subnets from allowlist. Continue?'
	String get removePrivateSubnetsDescription => 'Enabling LAN discovery will remove any private subnets from allowlist. Continue?';

	/// en: 'Remove overlapping subnets?'
	String get removeOverlappingSubnets => 'Remove overlapping subnets?';

	/// en: 'The range you are trying to add includes previously allowed subnets. To prevent conflicts, we’ll delete the redundant ones.'
	String get removeOverlappingSubnetsDescription => 'The range you are trying to add includes previously allowed subnets. To prevent conflicts, we’ll delete the redundant ones.';

	/// en: 'Remove'
	String get removeWord => 'Remove';

	/// en: 'Adding this range may be unsafe'
	String get addingTooWideSubnet => 'Adding this range may be unsafe';

	/// en: 'The subnet range you entered is too large and may allow unprotected traffic to enter or leave your device.'
	String get addingTooWideSubnetDescription => 'The subnet range you entered is too large and may allow unprotected traffic to enter or leave your device.';

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

	/// en: 'Recent Connections'
	String get recentConnections => 'Recent Connections';

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

	/// en: 'Pause for 5 minutes'
	String get pauseFor5Min => 'Pause for 5 minutes';

	/// en: 'Pause for 15 minutes'
	String get pauseFor15Min => 'Pause for 15 minutes';

	/// en: 'Pause for 30 minutes'
	String get pauseFor30Min => 'Pause for 30 minutes';

	/// en: 'Pause for 1 hour'
	String get pauseFor1Hour => 'Pause for 1 hour';

	/// en: 'Pause for 24 hours'
	String get pauseFor24Hours => 'Pause for 24 hours';

	/// en: 'Secured'
	String get secured => 'Secured';

	/// en: 'Not secured'
	String get notSecured => 'Not secured';

	/// en: 'Secure my connection'
	String get secureMyConnection => 'Secure my connection';

	/// en: 'VPN connection resumes in $minutes:$seconds'
	String VPNResumesIn({required Object minutes, required Object seconds}) => 'VPN connection resumes in ${minutes}:${seconds}';

	/// en: 'VPN connection resumes in $hours:$minutes:$seconds'
	String VPNResumesInWithHours({required Object hours, required Object minutes, required Object seconds}) => 'VPN connection resumes in ${hours}:${minutes}:${seconds}';

	/// en: 'Reconnect'
	String get reconnect => 'Reconnect';

	/// en: 'Change VPN settings'
	String get changeVPNsettings => 'Change VPN settings';

	/// en: 'Get help'
	String get getHelp => 'Get help';

	/// en: 'Pause connection'
	String get pauseConnection => 'Pause connection';

	/// en: 'Firewall can’t be turned off while Kill Switch is on'
	String get cannotStopFWWhileKSOnTitle => 'Firewall can’t be turned off while Kill Switch is on';

	/// en: 'Firewall is needed for Kill Switch to work. Turn off Kill Switch first in VPN connection settings.'
	String get cannotStopFWWhileKSOnDesc => 'Firewall is needed for Kill Switch to work. Turn off Kill Switch first in VPN connection settings.';

	/// en: 'Too many connection attempts'
	String get connectionLimitReachedTitle => 'Too many connection attempts';

	/// en: 'Wait a while before trying again. Retrying now can make the waiting period longer. If the issue persists, check our help guide for other possible causes. [Open help guide](https://support.nordvpn.com/hc/en-us/articles/47181405478417-I-get-the-Session-Limit-Reached-error-on-NordVPN?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=ens_error-session_limit&nm=app&ns=nordvpn-linux-gui&nc=ens_error-session_limit)'
	String get connectionLimitReachedBody => 'Wait a while before trying again. Retrying now can make the waiting period longer. If the issue persists, check our help guide for other possible causes.\n[Open help guide](https://support.nordvpn.com/hc/en-us/articles/47181405478417-I-get-the-Session-Limit-Reached-error-on-NordVPN?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=ens_error-session_limit&nm=app&ns=nordvpn-linux-gui&nc=ens_error-session_limit)';

	/// en: '${technology: String} is not available with your current auto-connect settings'
	String incompatibleProtocolWithAutoConnectTitle({required String technology}) => '${technology} is not available with your current auto-connect settings';

	/// en: 'Set auto-connect to a different location or turn it off, then try again.'
	String get incompatibleProtocolWithAutoConnectDesc => 'Set auto-connect to a different location or turn it off, then try again.';
}

/// The flat map containing all translations for locale <en>.
/// Only for edge cases! For simple maps, use the map function of this library.
///
/// The Dart AOT compiler has issues with very large switch statements,
/// so the map is split into smaller functions (512 entries each).
extension on Translations {
	dynamic _flatMapFunction(String path) {
		return switch (path) {
			'a11y.VPNResumesIn' => ({required Object minutes, required Object seconds}) => 'VPN connection resumes in ${minutes} minutes ${seconds} seconds',
			'a11y.VPNResumesInWithHours' => ({required Object hours, required Object minutes, required Object seconds}) => 'VPN connection resumes in ${hours} hours ${minutes} minutes ${seconds} seconds',
			'a11y.VPNPanelDisconnected' => ({required Object location}) => 'VPN Panel. Preferred location ${location}. Not secured.',
			'a11y.VPNPanelConnecting' => ({required Object location}) => 'VPN Panel. Connecting to ${location}.',
			'a11y.VPNPanelConnected' => ({required Object location}) => 'VPN Panel. Connected to ${location}.',
			'a11y.VPNPanelLoading' => 'VPN Panel. Loading.',
			'a11y.moreOptions' => 'More options',
			'a11y.listWithItems' => ({required Object number}) => 'List with ${number} items',
			'a11y.recentConnectionWithDetails' => ({required Object details}) => 'Recent connections. ${details}',
			'a11y.expandibleEntryExpand' => 'Expand',
			'a11y.expandibleEntryCollapse' => 'Collapse',
			'a11y.expandibleEntryExpanded' => 'Expanded',
			'a11y.expandibleEntryCollapsed' => 'Collapsed',
			'a11y.clear' => 'Clear',
			'a11y.popupWithContent' => ({required Object title, required Object message}) => '${title}. ${message}',
			'a11y.link' => ({required Object name}) => 'Link ${name}',
			'a11y.selected' => 'Selected',
			'a11y.notSelected' => 'Not selected',
			'a11y.appVersion' => ({required Object version}) => 'NordVPN version ${version}',
			'a11y.warningMessage' => ({required Object message}) => 'Warning. ${message}',
			'a11y.startPort' => 'Start port',
			'a11y.endPort' => 'End port',
			'cities.accra' => 'Accra',
			'cities.addis_ababa' => 'Addis Ababa',
			'cities.adelaide' => 'Adelaide',
			'cities.albuquerque' => 'Albuquerque',
			'cities.algiers' => 'Algiers',
			'cities.amman' => 'Amman',
			'cities.amsterdam' => 'Amsterdam',
			'cities.anchorage' => 'Anchorage',
			'cities.andorra_la_vella' => 'Andorra la Vella',
			'cities.antananarivo' => 'Antananarivo',
			'cities.ashburn' => 'Ashburn',
			'cities.astana' => 'Astana',
			'cities.asuncion' => 'Asuncion',
			'cities.athens' => 'Athens',
			'cities.atlanta' => 'Atlanta',
			'cities.auckland' => 'Auckland',
			'cities.baghdad' => 'Baghdad',
			'cities.baku' => 'Baku',
			'cities.baltimore' => 'Baltimore',
			'cities.bandar_seri_begawan' => 'Bandar Seri Begawan',
			'cities.bangkok' => 'Bangkok',
			'cities.banjul' => 'Banjul',
			'cities.barcelona' => 'Barcelona',
			'cities.beirut' => 'Beirut',
			'cities.belgrade' => 'Belgrade',
			'cities.belmopan' => 'Belmopan',
			'cities.berlin' => 'Berlin',
			'cities.billings' => 'Billings',
			'cities.bishkek' => 'Bishkek',
			'cities.bogota' => 'Bogota',
			'cities.boise' => 'Boise',
			'cities.bordeaux' => 'Bordeaux',
			'cities.boston' => 'Boston',
			'cities.bratislava' => 'Bratislava',
			'cities.bridgetown' => 'Bridgetown',
			'cities.brisbane' => 'Brisbane',
			'cities.brussels' => 'Brussels',
			'cities.bucharest' => 'Bucharest',
			'cities.budapest' => 'Budapest',
			'cities.buenos_aires' => 'Buenos Aires',
			'cities.buffalo' => 'Buffalo',
			'cities.burlington' => 'Burlington',
			'cities.cairo' => 'Cairo',
			'cities.calgary' => 'Calgary',
			'cities.caracas' => 'Caracas',
			'cities.castries' => 'Castries',
			'cities.charleston' => 'Charleston',
			'cities.charlotte' => 'Charlotte',
			'cities.cheyenne' => 'Cheyenne',
			'cities.chicago' => 'Chicago',
			'cities.chisinau' => 'Chisinau',
			'cities.colombo' => 'Colombo',
			'cities.columbus' => 'Columbus',
			'cities.copenhagen' => 'Copenhagen',
			'cities.dakar' => 'Dakar',
			'cities.dallas' => 'Dallas',
			'cities.denver' => 'Denver',
			'cities.des_moines' => 'Des Moines',
			'cities.detroit' => 'Detroit',
			'cities.dhaka' => 'Dhaka',
			'cities.dodoma' => 'Dodoma',
			'cities.doha' => 'Doha',
			'cities.douglas' => 'Douglas',
			'cities.dubai' => 'Dubai',
			'cities.dublin' => 'Dublin',
			'cities.dushanbe' => 'Dushanbe',
			'cities.edinburgh' => 'Edinburgh',
			'cities.fargo' => 'Fargo',
			'cities.frankfurt' => 'Frankfurt',
			'cities.freetown' => 'Freetown',
			'cities.fujairah' => 'Fujairah',
			'cities.gaborone' => 'Gaborone',
			'cities.george_town' => 'George Town',
			'cities.georgetown' => 'Georgetown',
			'cities.glasgow' => 'Glasgow',
			'cities.guatemala_city' => 'Guatemala City',
			'cities.hagatna' => 'Hagatna',
			'cities.hamburg' => 'Hamburg',
			'cities.hamilton' => 'Hamilton',
			'cities.hanoi' => 'Hanoi',
			'cities.helsinki' => 'Helsinki',
			'cities.ho_chi_minh_city' => 'Ho Chi Minh City',
			'cities.hong_kong' => 'Hong Kong',
			'cities.honolulu' => 'Honolulu',
			'cities.houston' => 'Houston',
			'cities.huntington' => 'Huntington',
			'cities.indianapolis' => 'Indianapolis',
			'cities.istanbul' => 'Istanbul',
			'cities.jackson' => 'Jackson',
			'cities.jakarta' => 'Jakarta',
			'cities.johannesburg' => 'Johannesburg',
			'cities.kabul' => 'Kabul',
			'cities.kansas_city' => 'Kansas City',
			'cities.karachi' => 'Karachi',
			'cities.kathmandu' => 'Kathmandu',
			'cities.kingston' => 'Kingston',
			'cities.kuala_lumpur' => 'Kuala Lumpur',
			'cities.kuwait_city' => 'Kuwait City',
			'cities.kyiv' => 'Kyiv',
			'cities.la_paz' => 'La Paz',
			'cities.lagos' => 'Lagos',
			'cities.las_vegas' => 'Las Vegas',
			'cities.lewiston' => 'Lewiston',
			'cities.lilongwe' => 'Lilongwe',
			'cities.lima' => 'Lima',
			'cities.lisbon' => 'Lisbon',
			'cities.little_rock' => 'Little Rock',
			'cities.ljubljana' => 'Ljubljana',
			'cities.london' => 'London',
			'cities.los_angeles' => 'Los Angeles',
			'cities.louisville' => 'Louisville',
			'cities.luanda' => 'Luanda',
			'cities.lusaka' => 'Lusaka',
			'cities.luxembourg' => 'Luxembourg',
			'cities.madrid' => 'Madrid',
			'cities.male' => 'Male',
			'cities.manama' => 'Manama',
			'cities.manchester' => 'Manchester',
			'cities.manila' => 'Manila',
			'cities.maputo' => 'Maputo',
			'cities.marseille' => 'Marseille',
			'cities.mcallen' => 'McAllen',
			'cities.melbourne' => 'Melbourne',
			'cities.mexico' => 'Mexico',
			'cities.miami' => 'Miami',
			'cities.milan' => 'Milan',
			'cities.milwaukee' => 'Milwaukee',
			'cities.minneapolis' => 'Minneapolis',
			'cities.mogadishu' => 'Mogadishu',
			'cities.monrovia' => 'Monrovia',
			'cities.monte_carlo' => 'Monte Carlo',
			'cities.montevideo' => 'Montevideo',
			'cities.montgomery' => 'Montgomery',
			'cities.montreal' => 'Montreal',
			'cities.moroni' => 'Moroni',
			'cities.mumbai' => 'Mumbai',
			'cities.nairobi' => 'Nairobi',
			'cities.nashua' => 'Nashua',
			'cities.nashville' => 'Nashville',
			'cities.nassau' => 'Nassau',
			'cities.naypyidaw' => 'Naypyidaw',
			'cities.ndjamena' => 'NDjamena',
			'cities.new_haven' => 'New Haven',
			'cities.new_orleans' => 'New Orleans',
			'cities.new_york' => 'New York',
			'cities.nicosia' => 'Nicosia',
			'cities.nouakchott' => 'Nouakchott',
			'cities.nuuk' => 'Nuuk',
			'cities.oklahoma_city' => 'Oklahoma City',
			'cities.omaha' => 'Omaha',
			'cities.osaka' => 'Osaka',
			'cities.oslo' => 'Oslo',
			'cities.ouagadougou' => 'Ouagadougou',
			'cities.palermo' => 'Palermo',
			'cities.panama_city' => 'Panama City',
			'cities.paramaribo' => 'Paramaribo',
			'cities.paris' => 'Paris',
			'cities.perth' => 'Perth',
			'cities.phnom_penh' => 'Phnom Penh',
			'cities.phoenix' => 'Phoenix',
			'cities.pittsburgh' => 'Pittsburgh',
			'cities.podgorica' => 'Podgorica',
			'cities.port_louis' => 'Port Louis',
			'cities.port_moresby' => 'Port Moresby',
			'cities.port_of_spain' => 'Port of Spain',
			'cities.portland' => 'Portland',
			'cities.porto_novo' => 'Porto Novo',
			'cities.prague' => 'Prague',
			'cities.praia' => 'Praia',
			'cities.providence' => 'Providence',
			'cities.quito' => 'Quito',
			'cities.rabat' => 'Rabat',
			'cities.reykjavik' => 'Reykjavik',
			'cities.riga' => 'Riga',
			'cities.rome' => 'Rome',
			'cities.saint_helier' => 'Saint Helier',
			'cities.saint_john' => 'Saint John',
			'cities.saint_louis' => 'Saint Louis',
			'cities.salt_lake_city' => 'Salt Lake City',
			'cities.san_francisco' => 'San Francisco',
			'cities.san_jose' => 'San Jose',
			'cities.san_juan' => 'San Juan',
			'cities.san_salvador' => 'San Salvador',
			'cities.sanaa' => 'Sanaa',
			'cities.santiago' => 'Santiago',
			'cities.santo_domingo' => 'Santo Domingo',
			'cities.sao_paulo' => 'Sao Paulo',
			'cities.sarajevo' => 'Sarajevo',
			'cities.seattle' => 'Seattle',
			'cities.seoul' => 'Seoul',
			'cities.singapore' => 'Singapore',
			'cities.sioux_falls' => 'Sioux Falls',
			'cities.skopje' => 'Skopje',
			'cities.sofia' => 'Sofia',
			'cities.stockholm' => 'Stockholm',
			'cities.strasbourg' => 'Strasbourg',
			'cities.suva' => 'Suva',
			'cities.sydney' => 'Sydney',
			'cities.taipei' => 'Taipei',
			'cities.tallinn' => 'Tallinn',
			'cities.tashkent' => 'Tashkent',
			'cities.tbilisi' => 'Tbilisi',
			'cities.tegucigalpa' => 'Tegucigalpa',
			'cities.tel_aviv' => 'Tel Aviv',
			'cities.thimphu' => 'Thimphu',
			'cities.tirana' => 'Tirana',
			'cities.tokyo' => 'Tokyo',
			'cities.toronto' => 'Toronto',
			'cities.trenton' => 'Trenton',
			'cities.tripoli' => 'Tripoli',
			'cities.tunis' => 'Tunis',
			'cities.ulaanbaatar' => 'Ulaanbaatar',
			'cities.vaduz' => 'Vaduz',
			'cities.valletta' => 'Valletta',
			'cities.vancouver' => 'Vancouver',
			'cities.victoria' => 'Victoria',
			'cities.vienna' => 'Vienna',
			'cities.vientiane' => 'Vientiane',
			'cities.vilnius' => 'Vilnius',
			'cities.warsaw' => 'Warsaw',
			'cities.wichita' => 'Wichita',
			'cities.wilmington' => 'Wilmington',
			'cities.yerevan' => 'Yerevan',
			'cities.zagreb' => 'Zagreb',
			'cities.zurich' => 'Zurich',
			'countries.AD' => 'Andorra',
			'countries.AE' => 'United Arab Emirates',
			'countries.AF' => 'Afghanistan',
			'countries.AG' => 'Antigua and Barbuda',
			'countries.AL' => 'Albania',
			'countries.AM' => 'Armenia',
			'countries.AO' => 'Angola',
			'countries.AR' => 'Argentina',
			'countries.AT' => 'Austria',
			'countries.AU' => 'Australia',
			'countries.AZ' => 'Azerbaijan',
			'countries.BA' => 'Bosnia and Herzegovina',
			'countries.BB' => 'Barbados',
			'countries.BD' => 'Bangladesh',
			'countries.BE' => 'Belgium',
			'countries.BF' => 'Burkina Faso',
			'countries.BG' => 'Bulgaria',
			'countries.BH' => 'Bahrain',
			'countries.BJ' => 'Benin',
			'countries.BM' => 'Bermuda',
			'countries.BN' => 'Brunei Darussalam',
			'countries.BO' => 'Bolivia',
			'countries.BR' => 'Brazil',
			'countries.BS' => 'Bahamas',
			'countries.BT' => 'Bhutan',
			'countries.BW' => 'Botswana',
			'countries.BZ' => 'Belize',
			'countries.CA' => 'Canada',
			'countries.CH' => 'Switzerland',
			'countries.CL' => 'Chile',
			'countries.CO' => 'Colombia',
			'countries.CR' => 'Costa Rica',
			'countries.CV' => 'Cape Verde',
			'countries.CY' => 'Cyprus',
			'countries.CZ' => 'Czech Republic',
			'countries.DE' => 'Germany',
			'countries.DK' => 'Denmark',
			'countries.DO' => 'Dominican Republic',
			'countries.DZ' => 'Algeria',
			'countries.EC' => 'Ecuador',
			'countries.EE' => 'Estonia',
			'countries.EG' => 'Egypt',
			'countries.ES' => 'Spain',
			'countries.ET' => 'Ethiopia',
			'countries.FI' => 'Finland',
			'countries.FJ' => 'Fiji',
			'countries.FR' => 'France',
			'countries.GB' => 'United Kingdom',
			'countries.GE' => 'Georgia',
			'countries.GH' => 'Ghana',
			'countries.GL' => 'Greenland',
			'countries.GM' => 'Gambia',
			'countries.GR' => 'Greece',
			'countries.GT' => 'Guatemala',
			'countries.GU' => 'Guam',
			'countries.GY' => 'Guyana',
			'countries.HK' => 'Hong Kong',
			'countries.HN' => 'Honduras',
			'countries.HR' => 'Croatia',
			'countries.HU' => 'Hungary',
			'countries.ID' => 'Indonesia',
			'countries.IE' => 'Ireland',
			'countries.IL' => 'Israel',
			'countries.IM' => 'Isle of Man',
			'countries.IN' => 'India',
			'countries.IQ' => 'Iraq',
			'countries.IS' => 'Iceland',
			'countries.IT' => 'Italy',
			'countries.JE' => 'Jersey',
			'countries.JM' => 'Jamaica',
			'countries.JO' => 'Jordan',
			'countries.JP' => 'Japan',
			'countries.KE' => 'Kenya',
			'countries.KG' => 'Kyrgyzstan',
			'countries.KH' => 'Cambodia',
			'countries.KM' => 'Comoros',
			'countries.KR' => 'South Korea',
			'countries.KW' => 'Kuwait',
			'countries.KY' => 'Cayman Islands',
			'countries.KZ' => 'Kazakhstan',
			'countries.LA' => 'Lao People\'s Democratic Republic',
			'countries.LB' => 'Lebanon',
			'countries.LC' => 'Saint Lucia',
			'countries.LI' => 'Liechtenstein',
			'countries.LK' => 'Sri Lanka',
			'countries.LR' => 'Liberia',
			'countries.LT' => 'Lithuania',
			'countries.LU' => 'Luxembourg',
			'countries.LV' => 'Latvia',
			'countries.LY' => 'Libyan Arab Jamahiriya',
			'countries.MA' => 'Morocco',
			'countries.MC' => 'Monaco',
			'countries.MD' => 'Moldova',
			'countries.ME' => 'Montenegro',
			'countries.MG' => 'Madagascar',
			'countries.MK' => 'North Macedonia',
			'countries.MM' => 'Myanmar',
			'countries.MN' => 'Mongolia',
			'countries.MR' => 'Mauritania',
			'countries.MT' => 'Malta',
			'countries.MU' => 'Mauritius',
			'countries.MV' => 'Maldives',
			'countries.MW' => 'Malawi',
			'countries.MX' => 'Mexico',
			'countries.MY' => 'Malaysia',
			'countries.MZ' => 'Mozambique',
			'countries.NG' => 'Nigeria',
			'countries.NL' => 'Netherlands',
			'countries.NO' => 'Norway',
			'countries.NP' => 'Nepal',
			'countries.NZ' => 'New Zealand',
			'countries.PA' => 'Panama',
			'countries.PE' => 'Peru',
			'countries.PG' => 'Papua New Guinea',
			'countries.PH' => 'Philippines',
			'countries.PK' => 'Pakistan',
			'countries.PL' => 'Poland',
			'countries.PR' => 'Puerto Rico',
			'countries.PT' => 'Portugal',
			'countries.PY' => 'Paraguay',
			'countries.QA' => 'Qatar',
			'countries.RO' => 'Romania',
			'countries.RS' => 'Serbia',
			'countries.SC' => 'Seychelles',
			'countries.SE' => 'Sweden',
			'countries.SG' => 'Singapore',
			'countries.SI' => 'Slovenia',
			'countries.SK' => 'Slovakia',
			'countries.SL' => 'Sierra Leone',
			'countries.SN' => 'Senegal',
			'countries.SO' => 'Somalia',
			'countries.SR' => 'Suriname',
			'countries.SV' => 'El Salvador',
			'countries.TD' => 'Chad',
			'countries.TH' => 'Thailand',
			'countries.TJ' => 'Tajikistan',
			'countries.TN' => 'Tunisia',
			'countries.TR' => 'Turkey',
			'countries.TT' => 'Trinidad and Tobago',
			'countries.TW' => 'Taiwan',
			'countries.TZ' => 'Tanzania',
			'countries.UA' => 'Ukraine',
			'countries.US' => 'United States',
			'countries.UY' => 'Uruguay',
			'countries.UZ' => 'Uzbekistan',
			'countries.VE' => 'Venezuela',
			'countries.VN' => 'Vietnam',
			'countries.YE' => 'Yemen',
			'countries.ZA' => 'South Africa',
			'countries.ZM' => 'Zambia',
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
			'daemon.code_3051_title' => 'Disabled protocol',
			'daemon.code_3051_msg' => 'Unable to connect with the current protocol. Please try a different one using the command: nordvpn set protocol.',
			'daemon.code_3068_title' => 'Here\'s what to know',
			'daemon.code_3068_msg' => 'Auto-connect is set to a dedicated server, which only works with NordLynx. To use a different protocol, choose another auto-connect location or turn off auto-connect first.',
			'daemon.code_3080_title' => 'Here\'s what to know',
			'daemon.code_3080_msg' => 'Auto-connect is set to obfuscated servers, which only work with NordWhisper. To use a different protocol, choose another auto-connect location or turn off auto-connect first.',
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
			'ui.noResultsFor' => ({required String searchStr}) => 'No results found for ${searchStr}',
			'ui.waitingToConnectToDaemon' => 'Connecting to the daemon...',
			'ui.fetchingData' => 'Fetching data',
			'ui.failedToFetchData' => 'Failed to fetch data',
			'ui.retry' => 'Retry',
			'ui.quickConnect' => 'Quick Connect',
			'ui.register' => 'Register',
			'ui.signIn' => 'Sign in',
			'ui.disconnect' => 'Disconnect',
			'ui.double_vpn' => 'Double VPN',
			'ui.fatalErrorMessage' => 'Fatal error',
			'ui.connected' => 'Connected',
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
			_ => null,
		} ?? switch (path) {
			'ui.launchAppAtStartup' => 'Launch at Startup',
			'ui.vpnProtocol' => 'VPN Protocol',
			'ui.notificationsStatus' => 'VPN Connection Status Notifications',
			'ui.firewall' => 'Firewall',
			'ui.firewallDescription' => 'Allow the use of the system firewall. When enabled, you can attach a firewall mark to VPN packets for custom firewall rules.',
			'ui.resetToDefaults' => 'Reset all app settings to default',
			'ui.firewallMark' => 'Firewall mark',
			'ui.reset' => 'Reset',
			'ui.confirm' => 'Confirm',
			'ui.cancel' => 'Cancel',
			'ui.killSwitchDescription' => 'Disable internet access if the VPN connection drops to secure your data from accidental exposure.',
			'ui.realTimeProtection' => 'Real-time protection',
			'ui.realTimeProtectionSubtitle' => 'Stop online scams and phishing attempts',
			'ui.realTimeProtectionDescription' => 'Stops online scams and phishing attempts while you’re connected to the VPN.',
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
			'ui.dedicatedIp' => 'Dedicated IP',
			'ui.dedicatedServer' => 'Dedicated Server',
			'ui.doubleVpn' => 'Double VPN',
			'ui.onionOverVpn' => 'Onion Over VPN',
			'ui.obfuscated' => 'Obfuscated',
			'ui.obfuscatedServers' => 'Obfuscated',
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
			'ui.securityAndPrivacySubtitle' => 'Allowlist, DNS, LAN discovery, firewall',
			'ui.appearance' => 'Appearance',
			'ui.light' => 'Light',
			'ui.dark' => 'Dark',
			'ui.showNotifications' => 'Show notifications',
			'ui.vpnConnection' => 'VPN connection',
			'ui.vpnConnectionSubtitle' => 'Auto-connect, Kill Switch, protocol',
			'ui.autoConnectDescription' => 'Automatically connect to the fastest available server or your chosen server location when the app starts.',
			'ui.fastestServer' => 'Fastest server',
			'ui.fastest' => 'Fastest',
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
			'ui.add' => 'Add',
			'ui.customDnsEntries' => ({required Object n}) => 'Custom DNS: ${n}/3',
			'ui.addUpTo3DnsServers' => 'Add up to 3 DNS servers',
			'ui.nothingHereYet' => 'Nothing here yet',
			'ui.addCustomDnsDescription' => 'To activate custom DNS, add at least one DNS server.',
			'ui.realTimeDisableCustomDNS' => 'Turn off custom DNS?',
			'ui.realTimeDisableCustomDNSDescription' => 'Real-time protection works with our default DNS servers only.',
			'ui.continueWord' => 'Continue',
			'ui.realTimeProtectionWillTurnOff' => 'Real-time protection will be turned off',
			'ui.realTimeProtectionWillTurnOffDescription' => 'Real-time protection works only with the default DNS. Set a custom DNS server anyway?',
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
			'ui.settingAutoconnectTo' => ({required Object target}) => 'Setting auto-connect to [${target}]...',
			'ui.doubleVpnDesc' => 'Encrypt your traffic twice for extra security',
			'ui.onionOverVpnDesc' => 'Use the Onion network with VPN protection',
			'ui.obfuscatedServersDesc' => 'Hide signs of VPN usage on restricted networks',
			'ui.save' => 'Save',
			'ui.close' => 'Close',
			'ui.to' => 'to',
			'ui.useCustomDns' => 'Use custom DNS',
			'ui.useCustomDnsDescription' => 'Add up to three DNS servers.',
			'ui.enterDnsAddress' => 'Enter DNS server address',
			'ui.duplicatedDnsServer' => 'This server is already on the list.',
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
			'ui.settingsNotApplied' => 'Some settings weren\'t applied',
			'ui.networkSettingsNotAppliedDescription' => 'Some default network settings could not be applied. Reset settings again to retry. If the problem continues, contact support.',
			'ui.settingsNotAppliedDescription' => ({required Object settings}) => 'Some default settings could not be applied. The reset finished, but these settings may not be active: ${settings}. Reset settings again to retry. If the problem continues, contact support.',
			'ui.nordWhisper' => 'NordWhisper',
			'ui.system' => 'System',
			'ui.removePrivateSubnets' => 'We\'ll remove private subnets from allowlist',
			'ui.removePrivateSubnetsDescription' => 'Enabling LAN discovery will remove any private subnets from allowlist. Continue?',
			'ui.removeOverlappingSubnets' => 'Remove overlapping subnets?',
			'ui.removeOverlappingSubnetsDescription' => 'The range you are trying to add includes previously allowed subnets. To prevent conflicts, we’ll delete the redundant ones.',
			'ui.removeWord' => 'Remove',
			'ui.addingTooWideSubnet' => 'Adding this range may be unsafe',
			'ui.addingTooWideSubnetDescription' => 'The subnet range you entered is too large and may allow unprotected traffic to enter or leave your device.',
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
			'ui.recentConnections' => 'Recent Connections',
			'ui.standardVpnServer' => 'Standard VPN Servers',
			'ui.reconnectNow' => 'Reconnect now',
			'ui.reconnectToChangeProtocol' => 'Reconnect to change protocol',
			'ui.reconnectToChangeProtocolDescription' => 'To apply this change, we\'ll reconnect you to the VPN.',
			'ui.reconnectToApplyChanges' => 'Reconnect to apply changes',
			'ui.reconnectToApplyChangesDescription' => 'Your new settings will take effect after you reconnect to the VPN.',
			'ui.gotIt' => 'Got it',
			'ui.pauseFor5Min' => 'Pause for 5 minutes',
			'ui.pauseFor15Min' => 'Pause for 15 minutes',
			'ui.pauseFor30Min' => 'Pause for 30 minutes',
			'ui.pauseFor1Hour' => 'Pause for 1 hour',
			'ui.pauseFor24Hours' => 'Pause for 24 hours',
			'ui.secured' => 'Secured',
			'ui.notSecured' => 'Not secured',
			'ui.secureMyConnection' => 'Secure my connection',
			'ui.VPNResumesIn' => ({required Object minutes, required Object seconds}) => 'VPN connection resumes in ${minutes}:${seconds}',
			'ui.VPNResumesInWithHours' => ({required Object hours, required Object minutes, required Object seconds}) => 'VPN connection resumes in ${hours}:${minutes}:${seconds}',
			'ui.reconnect' => 'Reconnect',
			'ui.changeVPNsettings' => 'Change VPN settings',
			'ui.getHelp' => 'Get help',
			'ui.pauseConnection' => 'Pause connection',
			'ui.cannotStopFWWhileKSOnTitle' => 'Firewall can’t be turned off while Kill Switch is on',
			'ui.cannotStopFWWhileKSOnDesc' => 'Firewall is needed for Kill Switch to work. Turn off Kill Switch first in VPN connection settings.',
			'ui.connectionLimitReachedTitle' => 'Too many connection attempts',
			'ui.connectionLimitReachedBody' => 'Wait a while before trying again. Retrying now can make the waiting period longer. If the issue persists, check our help guide for other possible causes.\n[Open help guide](https://support.nordvpn.com/hc/en-us/articles/47181405478417-I-get-the-Session-Limit-Reached-error-on-NordVPN?utm_medium=app&utm_source=nordvpn-linux-gui&utm_campaign=ens_error-session_limit&nm=app&ns=nordvpn-linux-gui&nc=ens_error-session_limit)',
			'ui.incompatibleProtocolWithAutoConnectTitle' => ({required String technology}) => '${technology} is not available with your current auto-connect settings',
			'ui.incompatibleProtocolWithAutoConnectDesc' => 'Set auto-connect to a different location or turn it off, then try again.',
			_ => null,
		};
	}
}
