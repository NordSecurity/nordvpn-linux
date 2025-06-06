package config

type ServerSelectionRule int

const (
	ServerSelectionRuleNone ServerSelectionRule = iota
	ServerSelectionRuleRecommended
	ServerSelectionRuleCity
	ServerSelectionRuleCountry
	ServerSelectionRuleSpecificServer
	ServerSelectionRuleGroup
	ServerSelectionRuleCountryWithGroup
	ServerSelectionRuleSpecificServerWithGroup
)
