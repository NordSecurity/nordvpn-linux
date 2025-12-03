package recents

import (
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
)

// filter provides a chainable interface for filtering recent connections
type filter struct {
	target     Model
	candidates []Model
	// flags for criteria
	excludeSpecificName bool
	excludeSpecificID   bool
	excludeTechnologies bool
	includeTechnologies []core.ServerTechnology
}

// newFilter creates a new filter for finding matching recent connections
func newFilter(target Model, candidates []Model) *filter {
	return &filter{
		target:     target,
		candidates: slices.Clone(candidates),
	}
}

// withSpecificServerOnlyFor configures the filter to match specific server fields (ID and name)
// only when the target connection type is one of the specified rules.
// For all other connection types, specific server fields are ignored during matching.
func (f *filter) withSpecificServerOnlyFor(rules []config.ServerSelectionRule) *filter {
	matchesRule := slices.Contains(rules, f.target.ConnectionType)
	if !matchesRule {
		// Connection type is NOT in the list - exclude specific server fields
		f.excludeSpecificID = true
		f.excludeSpecificName = true
	}
	return f
}

// withoutTechnologies excludes all technology-based criteria from the comparison.
// Use this when you want the match logic to ignore server technologies entirely.
// Calling this clears any previously set withTechnologies filter to avoid conflicting states.
func (f *filter) withoutTechnologies() *filter {
	f.includeTechnologies = nil
	f.excludeTechnologies = true
	return f
}

// withTechnologies requires that candidates support all of the specified server technologies.
// Use this when you need strict matching based on supported technologies.
// Calling this clears any previously set withoutTechnologies filter to avoid conflicting states.
func (f *filter) withTechnologies(serverTechs []core.ServerTechnology) *filter {
	f.excludeTechnologies = false
	f.includeTechnologies = serverTechs
	return f
}

// apply applies all configured filters and returns all matching models
func (f *filter) apply() []Model {
	var result []Model
	for _, candidate := range f.candidates {
		if f.matches(candidate) {
			result = append(result, candidate)
		}
	}
	return result
}

// matches checks if a single candidate matches the target based on filter criteria
func (f *filter) matches(m Model) bool {
	if m.Country != f.target.Country {
		return false
	}
	if m.City != f.target.City {
		return false
	}
	if m.Group != f.target.Group {
		return false
	}
	if m.CountryCode != f.target.CountryCode {
		return false
	}
	if m.ConnectionType != f.target.ConnectionType {
		return false
	}
	if m.IsVirtual != f.target.IsVirtual {
		return false
	}

	if !f.excludeSpecificName && m.SpecificServerName != f.target.SpecificServerName {
		return false
	}
	if !f.excludeSpecificID && m.SpecificServer != f.target.SpecificServer {
		return false
	}

	if len(f.includeTechnologies) > 0 {
		// Model candidate must contain all required technologies
		for _, reqTech := range f.includeTechnologies {
			if !slices.Contains(m.ServerTechnologies, reqTech) {
				return false
			}
		}
	} else if !f.excludeTechnologies {
		// Technologies must be identical
		if !slices.Equal(m.ServerTechnologies, f.target.ServerTechnologies) {
			return false
		}
	}

	return true
}
