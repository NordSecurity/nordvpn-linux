package recents

import (
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
)

// Filter provides a chainable interface for filtering recent connections
type Filter struct {
	target     Model
	candidates []Model
	// flags for criteria
	excludeSpecificName bool
	excludeSpecificID   bool
	excludeTechnologies bool
	includeTechnologies []core.ServerTechnology
}

// NewFilter creates a new filter for finding matching recent connections
func NewFilter(target Model, candidates []Model) *Filter {
	return &Filter{
		target:     target,
		candidates: slices.Clone(candidates),
	}
}

// WithoutSpecificServerFor excludes specific server fields from matching when the target
// connection type is NOT one of the specified rules.
// Use this to ignore specific server fields for location-based connections.
func (f *Filter) WithoutSpecificServerFor(rules []config.ServerSelectionRule) *Filter {
	matchesRule := slices.Contains(rules, f.target.ConnectionType)
	if !matchesRule {
		// Connection type is NOT in the list - exclude specific server fields
		f.excludeSpecificID = true
		f.excludeSpecificName = true
	}
	return f
}

// WithoutTechnologies excludes all technology-based criteria from the comparison.
// Use this when you want the match logic to ignore server technologies entirely.
// Calling this clears any previously set WithTechnologies filter to avoid conflicting states.
func (f *Filter) WithoutTechnologies() *Filter {
	f.includeTechnologies = nil
	f.excludeTechnologies = true
	return f
}

// WithTechnologies requires that candidates support all of the specified server technologies.
// Use this when you need strict matching based on supported technologies.
// Calling this clears any previously set WithoutTechnologies filter to avoid conflicting states.
func (f *Filter) WithTechnologies(serverTechs []core.ServerTechnology) *Filter {
	f.excludeTechnologies = false
	f.includeTechnologies = serverTechs
	return f
}

// Apply applies all configured filters and returns all matching models
func (f *Filter) Apply() []Model {
	var result []Model
	for _, candidate := range f.candidates {
		if f.matches(candidate) {
			result = append(result, candidate)
		}
	}
	return result
}

// matches checks if a single candidate matches the target based on filter criteria
func (f *Filter) matches(m Model) bool {
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
