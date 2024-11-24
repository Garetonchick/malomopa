package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Тест на успешное поднятие сервиса
func TestStartAndStopServices(t *testing.T) {
	client := NewDefaultClient()

	require.True(t, client.Start())

	require.True(t, client.PingOrderAssigner())
	// require.True(t, client.PingOrderExecutor()) // Этого вроде нет :)
	require.True(t, client.PingSources())

	require.True(t, client.Down())
}
