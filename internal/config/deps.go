package config

import (
	"fmt"
	"sort"
)

// StartOrder returns enabled service names in dependency order
// (dependencies before dependents) using Kahn's algorithm.
// Returns an error if a cycle is detected.
func StartOrder(services map[string]*ServiceEntry) ([]string, error) {
	// Collect enabled service names.
	enabled := make(map[string]bool, len(services))
	for name, entry := range services {
		if entry.Enabled {
			enabled[name] = true
		}
	}

	// Build in-degree counts and adjacency (within enabled set only).
	inDegree := make(map[string]int, len(enabled))
	dependents := make(map[string][]string) // dep → list of services that depend on it
	for name := range enabled {
		inDegree[name] = 0
	}
	for name := range enabled {
		def, ok := Registry[name]
		if !ok {
			continue
		}
		for _, dep := range def.DependsOn {
			if enabled[dep] {
				inDegree[name]++
				dependents[dep] = append(dependents[dep], name)
			}
		}
	}

	// Seed queue with zero-degree nodes (sorted for determinism).
	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}
	sort.Strings(queue)

	var order []string
	for len(queue) > 0 {
		// Pop first element.
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)

		// Collect and sort newly freed nodes for determinism.
		var freed []string
		for _, dep := range dependents[node] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				freed = append(freed, dep)
			}
		}
		sort.Strings(freed)
		queue = append(queue, freed...)
	}

	if len(order) != len(enabled) {
		return nil, fmt.Errorf("dependency cycle detected among enabled services")
	}
	return order, nil
}

// MissingDeps returns the dependency names of the given service that are
// not currently enabled. Useful when the API enables a service.
func MissingDeps(name string, services map[string]*ServiceEntry) []string {
	def, ok := Registry[name]
	if !ok {
		return nil
	}
	var missing []string
	for _, dep := range def.DependsOn {
		entry, exists := services[dep]
		if !exists || !entry.Enabled {
			missing = append(missing, dep)
		}
	}
	return missing
}

// EnabledDependents returns the names of enabled services that depend on
// the given service. Useful when the API disables a service.
func EnabledDependents(name string, services map[string]*ServiceEntry) []string {
	var deps []string
	for svcName, entry := range services {
		if !entry.Enabled {
			continue
		}
		def, ok := Registry[svcName]
		if !ok {
			continue
		}
		for _, dep := range def.DependsOn {
			if dep == name {
				deps = append(deps, svcName)
				break
			}
		}
	}
	sort.Strings(deps)
	return deps
}
