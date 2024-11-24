package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Тест на поведение кешей
func TestCache(t *testing.T) {
	client := NewDefaultClient()

	require.True(t, client.Start())

	require.True(t, client.PingOrderAssigner())
	// require.True(t, client.PingOrderExecutor()) // TODO
	require.True(t, client.PingSources())

	t.Run("Default work", func(t *testing.T) {
		MakeAssign(t, client, "1", "1")

		before, err := client.SourceCounters()
		require.NoError(t, err)

		MakeCancel(t, client, "1")
		MakeAssign(t, client, "1", "1")

		after, err := client.SourceCounters()
		require.NoError(t, err)

		CheckCountersDidntIncreaseForCachedSources(t, client, before, after)
	})

	t.Run("Turn off cached sources", func(t *testing.T) {
		MakeAssign(t, client, "2", "2")

		before, err := client.SourceCounters()
		require.NoError(t, err)

		MakeCancel(t, client, "2")

		TurnOffCacheableSources(t, client)

		MakeAssign(t, client, "2", "2")

		after, err := client.SourceCounters()
		require.NoError(t, err)

		CheckCountersDidntIncreaseForCachedSources(t, client, before, after)

		TurnOnCacheableSources(t, client)
	})

	require.True(t, client.Down())
}