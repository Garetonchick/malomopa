package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Тест на успешное завершение сервиса
func TestEnd(t *testing.T) {
	client := NewDefaultClient()
	require.True(t, client.Down())
}
