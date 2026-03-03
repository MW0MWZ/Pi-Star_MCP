package config

import (
	"sort"
	"testing"
)

func TestRegistryDependsOnTargetsExist(t *testing.T) {
	for name, def := range Registry {
		for _, dep := range def.DependsOn {
			if _, ok := Registry[dep]; !ok {
				t.Errorf("service %q depends on %q which is not in the registry", name, dep)
			}
		}
	}
}

func TestRegistryNoDuplicateNames(t *testing.T) {
	seen := make(map[string]bool, len(Registry))
	for name, def := range Registry {
		if name != def.Name {
			t.Errorf("registry key %q does not match ServiceDef.Name %q", name, def.Name)
		}
		if seen[name] {
			t.Errorf("duplicate registry entry: %q", name)
		}
		seen[name] = true
	}
}

func TestRegistryNoEmptyFields(t *testing.T) {
	for name, def := range Registry {
		if def.Name == "" {
			t.Errorf("service %q has empty Name", name)
		}
		if def.DisplayName == "" {
			t.Errorf("service %q has empty DisplayName", name)
		}
		if def.DefaultBinaryPath == "" {
			t.Errorf("service %q has empty DefaultBinaryPath", name)
		}
		if def.ConfigArgStyle == ArgPort {
			if def.DefaultArgs == "" {
				t.Errorf("service %q has ArgPort style but empty DefaultArgs", name)
			}
		} else if def.DefaultConfigPath == "" {
			t.Errorf("service %q has empty DefaultConfigPath", name)
		}
	}
}

func TestRegistryNoCycles(t *testing.T) {
	// Enable all services and verify topological sort succeeds.
	all := make(map[string]*ServiceEntry, len(Registry))
	for name := range Registry {
		all[name] = &ServiceEntry{Enabled: true}
	}
	order, err := StartOrder(all)
	if err != nil {
		t.Fatalf("dependency cycle detected: %v", err)
	}
	if len(order) != len(Registry) {
		t.Errorf("StartOrder returned %d services, want %d", len(order), len(Registry))
	}
}

func TestServiceNamesReturnsSortedCompleteList(t *testing.T) {
	names := ServiceNames()
	if len(names) != len(Registry) {
		t.Fatalf("ServiceNames() returned %d names, want %d", len(names), len(Registry))
	}
	if !sort.StringsAreSorted(names) {
		t.Error("ServiceNames() is not sorted")
	}
	for _, name := range names {
		if _, ok := Registry[name]; !ok {
			t.Errorf("ServiceNames() contains %q which is not in Registry", name)
		}
	}
}

func TestLookupService(t *testing.T) {
	def, ok := LookupService("mmdvmhost")
	if !ok {
		t.Fatal("LookupService(mmdvmhost) returned false")
	}
	if def.DisplayName != "MMDVMHost" {
		t.Errorf("DisplayName = %q, want MMDVMHost", def.DisplayName)
	}

	_, ok = LookupService("nonexistent")
	if ok {
		t.Error("LookupService(nonexistent) should return false")
	}
}
