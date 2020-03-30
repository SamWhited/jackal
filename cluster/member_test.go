/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMembers_Member(t *testing.T) {
	var ms Members

	ms = append(ms, Member{"a1", "h1", "10"})
	ms = append(ms, Member{"a2", "h2", "20"})

	m := ms.Member("a3")
	require.Nil(t, m)

	m = ms.Member("a2")
	require.NotNil(t, m)

	require.Equal(t, m.Host, "h2")
}
