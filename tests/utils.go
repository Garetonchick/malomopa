package test

import (
	sources "malomopa/internal/sources"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TurnOffCacheableSources(t *testing.T, client Client) {
	require.True(t, client.TurnOffConfigsSource())
	require.True(t, client.TurnOffZonesInfoSource())
}

func TurnOnCacheableSources(t *testing.T, client Client) {
	require.True(t, client.TurnOnConfigsSource())
	require.True(t, client.TurnOnZonesInfoSource())
}

func CheckCountersDidntIncreaseForCachedSources(t *testing.T, client Client, before, after *sources.HandlersCountersResponse) {
	require.Equal(t, before.GeneralInfoCounter+1, after.GeneralInfoCounter)
	require.Equal(t, before.ZoneInfoCounter, after.ZoneInfoCounter) // cached
	require.Equal(t, before.ExecutorProfileCounter+1, after.ExecutorProfileCounter)
	require.Equal(t, before.ConfigsCounter, after.ConfigsCounter) // cached
	require.Equal(t, before.TollRoadsInfoCountter+1, after.TollRoadsInfoCountter)
}

func MakeAssign(t *testing.T, client Client, orderID, executorID string) {
	code, err := client.AssignOrder(orderID, executorID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
}

func MakeCancel(t *testing.T, client Client, orderID string) {
	resp, err := client.CancelOrder(orderID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.code)
}
