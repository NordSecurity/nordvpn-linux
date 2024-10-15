package core

const (
	HeaderDigest = "x-digest"

	// CDNURL is the url for NordCDN
	CDNURL = "https://downloads.nordcdn.com"

	// InsightsURL defines url to get information about ip
	// Used by JobInsights every 30mins to set the user country
	InsightsURL = "/v1/helpers/ips/insights"

	// PlanURL defines endpoint to fetch plans
	PlanURL = "/v1/plans?filters[plans.active]=1&filters[plans.type]=linux"

	// ServersURL defines url to get servers list
	// Used as a fallback if /v1/servers/recommendations returns
	// an empty list or a http error
	ServersURL = "/v1/servers"

	// ServersCountriesURL defines url to get servers countries list
	// Used by JobCountries every 6h to populate /var/lib/nordvpn/data/countries.dat
	ServersCountriesURL = ServersURL + "/countries"

	// RecommendedServersURL defines url for recommended servers list
	RecommendedServersURL = ServersURL + "/recommendations"

	// notificationTokenURL defines url to retrieve Notification Center credentials
	notificationTokenURL = "/v1/notifications/tokens"

	// notificationTokenRevokeURL defines url to revoke Notification Center credentials
	notificationTokenRevokeURL = "/v1/notifications/tokens/revoke"

	// UsersURL defines url to create a new user
	UsersURL = "/v1/users"

	urlOAuth2Login  = UsersURL + "/oauth/login"
	urlOAuth2Logout = UsersURL + "/oauth/logout"
	urlOAuth2Token  = UsersURL + "/oauth/token"

	// TokensURL defines url to get user token
	TokensURL = UsersURL + "/tokens" // #nosec

	// ServicesURL defines url to check user's current/expired services
	ServicesURL = UsersURL + "/services"

	// urlOrders defines URL to list user's orders
	urlOrders = UsersURL + "/orders"

	// urlOrders defines URL to list user's payments
	urlPayments = UsersURL + "/payments"

	// CredentialsURL defines url to generate openvpn credentials
	CredentialsURL = ServicesURL + "/credentials"

	// CurrentUserURL defines url to check user's metadata
	CurrentUserURL = UsersURL + "/current"

	// TokenRenewURL defines url to renew user's token
	TokenRenewURL = UsersURL + "/tokens/renew" // #nosec

	TrustedPassTokenURL = UsersURL + "/oauth/tokens/trusted"

	MFAStatusURL = UsersURL + "/oauth/mfa/status"

	// ServersURLConnectQuery is all servers query optimized
	// for minimal dm size required
	// to create servers maps and calculate their penalty scores
	// so instead of downloading 15mb we download 1.5mb
	// and when connecting, download all info about specific server
	//
	// no problems with this logic so far
	ServersURLConnectQuery = "?limit=1073741824" +
		"&filters[servers.status]=online" +
		"&fields[servers.id]" +
		"&fields[servers.name]" +
		"&fields[servers.hostname]" +
		"&fields[servers.station]" +
		"&fields[servers.status]" +
		"&fields[servers.load]" +
		"&fields[servers.created_at]" +
		"&fields[servers.groups.id]" +
		"&fields[servers.groups.title]" +
		"&fields[servers.technologies.id]" +
		"&fields[servers.technologies.metadata]" +
		"&fields[servers.technologies.pivot.status]" +
		"&fields[servers.specifications.identifier]" +
		"&fields[servers.specifications.values.value]" +
		"&fields[servers.locations.country.name]" +
		"&fields[servers.locations.country.code]" +
		"&fields[servers.locations.country.city.name]" +
		"&fields[servers.locations.country.city.latitude]" +
		"&fields[servers.locations.country.city.longitude]" +
		"&fields[servers.locations.country.city.hub_score]" +
		"&fields[servers.ips]"

	RecommendedServersURLConnectQuery = "?limit=%d" +
		"&filters[servers.status]=online" +
		"&filters[servers_technologies]=%d" +
		"&filters[servers_technologies][pivot][status]=online" +
		"&fields[servers.id]" +
		"&fields[servers.name]" +
		"&fields[servers.hostname]" +
		"&fields[servers.station]" +
		"&fields[servers.status]" +
		"&fields[servers.load]" +
		"&fields[servers.created_at]" +
		"&fields[servers.groups.id]" +
		"&fields[servers.groups.title]" +
		"&fields[servers.technologies.id]" +
		"&fields[servers.technologies.metadata]" +
		"&fields[servers.technologies.pivot.status]" +
		"&fields[servers.specifications.identifier]" +
		"&fields[servers.specifications.values.value]" +
		"&fields[servers.locations.country.name]" +
		"&fields[servers.locations.country.code]" +
		"&fields[servers.locations.country.city.name]" +
		"&fields[servers.locations.country.city.latitude]" +
		"&fields[servers.locations.country.city.longitude]" +
		"&coordinates[longitude]=%f&coordinates[latitude]=%f" +
		"&fields[servers.ips]"

	RecommendedServersCountryFilter = "&filters[country_id]=%d"
	RecommendedServersCityFilter    = "&filters[country_city_id]=%d"
	RecommendedServersGroupsFilter  = "&filters[servers_groups]=%d"

	// ServersURLSpecificQuery defines query params for a specific server
	ServersURLSpecificQuery = "?filters[servers.id]=%d"

	// ovpnTemplateURL defines url to ovpn server template
	ovpnTemplateURL = "/configs/templates/ovpn/1.0/template.xslt"

	// ovpnObfsTemplateURL defines url to ovpn obfuscated server template
	ovpnObfsTemplateURL = "/configs/templates/ovpn_xor/1.0/template.xslt"

	// threatProtectionLiteURL defines url of the cybersec file
	threatProtectionLiteURL = "/configs/dns/cybersec.json"

	// DebFileinfoURLFormat is the path to debian repository's package information
	DebFileinfoURLFormat = "/deb/%s/debian/dists/stable/main/binary-%s/Packages.gz"

	// RpmRepoMdURLFormat is the path to rpm repository's information
	RpmRepoMdURLFormat = "/yum/%s/centos/%s/repodata/%s"

	// RpmRepoMdURL is the path to rpm repository's information file
	RpmRepoMdURL = "repomd.xml"

	// RepoTypeProduction defines production repo type
	RepoTypeProduction = "nordvpn"

	// RepoTypeTest defines non-production (qa, development) repo type
	RepoTypeTest = "nordvpn-test"
)
