package daemon

import (
	"context"
	"sort"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

var countryCodeMap = map[string]string{
	"Slovakia":               "sk",
	"Bosnia_And_Herzegovina": "ba",
	"Malaysia":               "my",
	"Romania":                "ro",
	"Brazil":                 "br",
	"Ireland":                "ie",
	"Ukraine":                "ua",
	"Thailand":               "th",
	"South_Africa":           "za",
	"Sweden":                 "se",
	"Finland":                "fi",
	"Spain":                  "es",
	"Czech_Republic":         "cz",
	"India":                  "in",
	"Singapore":              "sg",
	"North_Macedonia":        "mk",
	"Indonesia":              "id",
	"Belgium":                "be",
	"United_Kingdom":         "uk",
	"United_States":          "us",
	"Taiwan":                 "tw",
	"Luxembourg":             "lu",
	"Israel":                 "il",
	"Japan":                  "jp",
	"Netherlands":            "nl",
	"Denmark":                "dk",
	"Moldova":                "md",
	"France":                 "fr",
	"New_Zealand":            "nz",
	"Australia":              "au",
	"Austria":                "at",
	"Chile":                  "cl",
	"Vietnam":                "vn",
	"Croatia":                "hr",
	"Hungary":                "hu",
	"Hong_Kong":              "hk",
	"Georgia":                "ge",
	"Iceland":                "is",
	"Portugal":               "pt",
	"Poland":                 "pl",
	"Switzerland":            "ch",
	"Estonia":                "ee",
	"Greece":                 "gr",
	"Argentina":              "ar",
	"Italy":                  "it",
	"Latvia":                 "lv",
	"Slovenia":               "si",
	"Norway":                 "no",
	"Canada":                 "ca",
	"Mexico":                 "mx",
	"South_Korea":            "kr",
	"Albania":                "al",
	"Serbia":                 "rs",
	"Germany":                "de",
	"Costa_Rica":             "cr",
	"Bulgaria":               "bg",
	"Turkey":                 "tr",
	"Cyprus":                 "cy",
}

// Countries provides country command and country autocompletion.
func (r *RPC) Countries(ctx context.Context, in *pb.CountriesRequest) (*pb.Payload, error) {
	var countryNames []string
	for country := range r.dm.GetAppData().CountryNames[in.GetObfuscate()][in.GetProtocol()].Iter() {
		countryNames = append(countryNames, country.(string))
	}
	sort.Strings(countryNames)
	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: countryNames,
	}, nil
}

func (r *RPC) FrontendCountries(ctx context.Context, in *pb.CountriesRequest) (*pb.CountriesResponse, error) {
	countries := r.dm.GetAppData().CountryNames[in.GetObfuscate()][in.GetProtocol()]
	var resp []*pb.Country
	for countryName := range countries.Iter() {
		country := &pb.Country{Name: countryName.(string)}
		if code, ok := countryCodeMap[countryName.(string)]; ok {
			country.Code = code
		} else {
			country.Code = "??"
		}
		resp = append(resp, country)
	}

	sort.Slice(resp, func(i, j int) bool {
		if resp[i].Name < resp[j].Name {
			return true
		}
		return false
	})

	return &pb.CountriesResponse{Countries: resp}, nil
}
