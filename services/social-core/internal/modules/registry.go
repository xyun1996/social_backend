package modules

import "sort"

// Descriptor documents a bounded context that will move into the consolidated
// social-core runtime during the rebuild.
type Descriptor struct {
	Name         string   `json:"name"`
	Scope        string   `json:"scope"`
	DependsOn    []string `json:"depends_on,omitempty"`
	Phase        string   `json:"phase"`
	ProductGrade bool     `json:"product_grade"`
}

// Registry is the product-rebuild catalog for social-core modules.
type Registry struct {
	descriptors []Descriptor
}

// NewRegistry returns the canonical module set for phase A of the product
// rebuild.
func NewRegistry() Registry {
	descriptors := []Descriptor{
		{Name: "identity", Scope: "login, token lifecycle, session principal", Phase: "phase-a", ProductGrade: true},
		{Name: "social", Scope: "friends, blocks, relationship read models", DependsOn: []string{"identity"}, Phase: "phase-a", ProductGrade: true},
		{Name: "invite", Scope: "cross-domain invite lifecycle", DependsOn: []string{"identity", "social"}, Phase: "phase-a", ProductGrade: true},
		{Name: "private-chat", Scope: "direct messaging and conversation summaries", DependsOn: []string{"identity", "social"}, Phase: "phase-a", ProductGrade: true},
		{Name: "guild-basics", Scope: "guild membership, governance, and guild chat entry points", DependsOn: []string{"identity", "invite"}, Phase: "phase-a", ProductGrade: false},
		{Name: "party-basics", Scope: "party formation, readiness, and invite-backed membership", DependsOn: []string{"identity", "invite"}, Phase: "phase-a", ProductGrade: false},
	}

	sort.Slice(descriptors, func(i, j int) bool {
		return descriptors[i].Name < descriptors[j].Name
	})

	return Registry{descriptors: descriptors}
}

// Descriptors returns a copy of the module catalog so callers can't mutate the
// registry in place.
func (r Registry) Descriptors() []Descriptor {
	result := make([]Descriptor, len(r.descriptors))
	copy(result, r.descriptors)
	return result
}
