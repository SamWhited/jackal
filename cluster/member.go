/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

type Member struct {
	AllocationID string
	Host         string
	Port         string
}

type Members []Member

func (m Members) Member(allocationID string) *Member {
	for _, m := range m {
		if m.AllocationID == allocationID {
			return &m
		}
	}
	return nil
}
