package config

import (
	"reflect"
	"testing"
)

func TestStartOrderChain(t *testing.T) {
	services := map[string]*ServiceEntry{
		"mmdvmhost":  {Enabled: true},
		"dmrgateway": {Enabled: true},
		"dmr2ysf":    {Enabled: true},
	}
	order, err := StartOrder(services)
	if err != nil {
		t.Fatalf("StartOrder: %v", err)
	}

	// mmdvmhost has no deps, dmrgateway has no deps, dmr2ysf depends on dmrgateway.
	// Verify dmrgateway comes before dmr2ysf.
	idx := make(map[string]int, len(order))
	for i, name := range order {
		idx[name] = i
	}
	if idx["dmrgateway"] >= idx["dmr2ysf"] {
		t.Errorf("dmrgateway (pos %d) should come before dmr2ysf (pos %d)", idx["dmrgateway"], idx["dmr2ysf"])
	}
}

func TestStartOrderAllEnabled(t *testing.T) {
	all := make(map[string]*ServiceEntry, len(Registry))
	for name := range Registry {
		all[name] = &ServiceEntry{Enabled: true}
	}
	order, err := StartOrder(all)
	if err != nil {
		t.Fatalf("StartOrder with all services: %v", err)
	}
	if len(order) != len(Registry) {
		t.Errorf("got %d services, want %d", len(order), len(Registry))
	}

	// Verify every dependency comes before its dependent.
	idx := make(map[string]int, len(order))
	for i, name := range order {
		idx[name] = i
	}
	for _, name := range order {
		def := Registry[name]
		for _, dep := range def.DependsOn {
			if idx[dep] >= idx[name] {
				t.Errorf("%s (pos %d) depends on %s (pos %d) but comes first", name, idx[name], dep, idx[dep])
			}
		}
	}
}

func TestStartOrderSingleNoDeps(t *testing.T) {
	services := map[string]*ServiceEntry{
		"ysfparrot": {Enabled: true},
	}
	order, err := StartOrder(services)
	if err != nil {
		t.Fatalf("StartOrder: %v", err)
	}
	if !reflect.DeepEqual(order, []string{"ysfparrot"}) {
		t.Errorf("got %v, want [ysfparrot]", order)
	}
}

func TestStartOrderEmptyMap(t *testing.T) {
	order, err := StartOrder(map[string]*ServiceEntry{})
	if err != nil {
		t.Fatalf("StartOrder: %v", err)
	}
	if len(order) != 0 {
		t.Errorf("expected empty order, got %v", order)
	}
}

func TestMissingDeps(t *testing.T) {
	services := map[string]*ServiceEntry{
		"dmr2ysf": {Enabled: true},
	}
	missing := MissingDeps("dmr2ysf", services)
	if !reflect.DeepEqual(missing, []string{"dmrgateway"}) {
		t.Errorf("MissingDeps = %v, want [dmrgateway]", missing)
	}
}

func TestMissingDepsAllPresent(t *testing.T) {
	services := map[string]*ServiceEntry{
		"dmrgateway": {Enabled: true},
		"dmr2ysf":    {Enabled: true},
	}
	missing := MissingDeps("dmr2ysf", services)
	if len(missing) != 0 {
		t.Errorf("MissingDeps = %v, want empty", missing)
	}
}

func TestMissingDepsNoDeps(t *testing.T) {
	services := map[string]*ServiceEntry{
		"mmdvmhost": {Enabled: true},
	}
	missing := MissingDeps("mmdvmhost", services)
	if len(missing) != 0 {
		t.Errorf("MissingDeps = %v, want empty", missing)
	}
}

func TestEnabledDependents(t *testing.T) {
	services := map[string]*ServiceEntry{
		"dmrgateway": {Enabled: true},
		"dmr2ysf":    {Enabled: true},
		"dmr2nxdn":   {Enabled: true},
		"ysfgateway": {Enabled: true},
	}
	deps := EnabledDependents("dmrgateway", services)
	expected := []string{"dmr2nxdn", "dmr2ysf"}
	if !reflect.DeepEqual(deps, expected) {
		t.Errorf("EnabledDependents = %v, want %v", deps, expected)
	}
}

func TestEnabledDependentsNone(t *testing.T) {
	services := map[string]*ServiceEntry{
		"dmrgateway": {Enabled: true},
		"ysfgateway": {Enabled: true},
	}
	deps := EnabledDependents("dmrgateway", services)
	if len(deps) != 0 {
		t.Errorf("EnabledDependents = %v, want empty", deps)
	}
}

func TestEnabledDependentsDisabledNotIncluded(t *testing.T) {
	services := map[string]*ServiceEntry{
		"dmrgateway": {Enabled: true},
		"dmr2ysf":    {Enabled: false},
	}
	deps := EnabledDependents("dmrgateway", services)
	if len(deps) != 0 {
		t.Errorf("EnabledDependents = %v, want empty (dmr2ysf is disabled)", deps)
	}
}
