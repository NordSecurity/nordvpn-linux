package config

type ServerSelectionRule int

const (
	ServerSelectionRule_RECOMMENDED ServerSelectionRule = iota
	ServerSelectionRule_CITY
	ServerSelectionRule_COUNTRY
	ServerSelectionRule_SPECIFIC_SERVER
	ServerSelectionRule_GROUP
	ServerSelectionRule_COUNTRY_WITH_GROUP
	ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP
)

func (r ServerSelectionRule) String() string {
	switch r {
	case ServerSelectionRule_RECOMMENDED:
		return "RECOMMENDED"
	case ServerSelectionRule_CITY:
		return "CITY"
	case ServerSelectionRule_COUNTRY:
		return "COUNTRY"
	case ServerSelectionRule_SPECIFIC_SERVER:
		return "SPECIFIC_SERVER"
	case ServerSelectionRule_GROUP:
		return "GROUP"
	case ServerSelectionRule_COUNTRY_WITH_GROUP:
		return "COUNTRY_WITH_GROUP"
	case ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP:
		return "SPECIFIC_SERVER_WITH_GROUP"
	default:
		return "UNKNOWN"
	}
}
