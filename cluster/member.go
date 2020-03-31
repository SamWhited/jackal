/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

// Member represents a single cluster member.
type Member struct {
	// AllocationID defines a unique cluster identifier.
	AllocationID string

	// Host defines member host ip.
	Host string

	// Port defines member host port.
	Port string
}

// Members groups all cluster members.
type Members []Member

// AllocationIDs returns all members allocation identfiiers.
func (m Members) AllocationIDs() []string {
	res := make([]string, len(m))
	for i := range m {
		res[i] = m[i].AllocationID
	}
	return res
}

// Member returns the member reference associated to some allocation identifier.
func (m Members) Member(allocationID string) *Member {
	for _, m := range m {
		if m.AllocationID == allocationID {
			return &m
		}
	}
	return nil
}
