package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Тест на успешное поднятие сервиса
func TestStart(t *testing.T) {
	client := NewDefaultClient()

	require.True(t, client.StartIfNotWorking())

	require.True(t, client.PingOrderAssigner())
	require.True(t, client.PingSources())

}
