// короч данпуз слуууушай:

// 1) простейший синхронный сценарий:
//     один поток делает assign_order,
//     потом забирает через acquire_order (возможно ретраит потому что у нас эвенчуал залупа).
//     Повторяет так несколько раз (ну видимо рандомит запросы как-то, не одно и то же долбит в очко)
// 2) такой же синхронный сценрий, тока вместо acquire_order пусть будет cancel_order (также нужны рэтраи)
// 3) несколько потоков конкурентно (инфра) и пишем и забираем, и отменяем
// 4) проверить невалидные варианты использования:
//     assign существующего заказа - 400,
//     cancel несуществующего или уже заканселившегося или забранного заказа - 400,
//     cancel старого заказа (старше 10 мин) - 400,
//     acquire несуществующего или отмененного заказа - 400

package test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogic(t *testing.T) {
	client := NewDefaultClient()

	// require.True(t, client.Start())
	require.True(t, client.PingOrderAssigner())
	// require.True(t, client.PingOrderExecutor()) // TODO
	require.True(t, client.PingSources())

	t.Run("Simple one-thread assign -- acquire", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			orderID := strconv.Itoa(i)
			executorID := strconv.Itoa(i)
			MakeAssign(t, client, orderID, executorID)
			MakeAcquire(t, client, executorID)
		}
	})

	// t.Run("Simple one-thread assign -- cancel", func(t *testing.T) {
	// 	for orderID := 10; orderID < 20; orderID++ {
	// 		executorID := orderID
	// 		MakeAssign(t, client, strconv.Itoa(orderID), strconv.Itoa(executorID))
	// 		MakeCancel(t, client, strconv.Itoa(orderID))
	// 	}
	// })

	// t.Run("Concurrent", func(t *testing.T) {
	// 	for orderID := 20; orderID < 30; orderID++ {
	// 		t.Run("Concurrent assign -- acquire", func(t *testing.T) {
	// 			t.Parallel()
	// 			executorID := orderID
	// 			MakeAssign(t, client, strconv.Itoa(orderID), strconv.Itoa(executorID))
	// 			MakeAcquire(t, client, strconv.Itoa(executorID))
	// 		})
	// 	}
	// 	for orderID := 30; orderID < 40; orderID++ {
	// 		t.Run("Concurrent assign -- cancel", func(t *testing.T) {
	// 			t.Parallel()
	// 			executorID := orderID
	// 			MakeAssign(t, client, strconv.Itoa(orderID), strconv.Itoa(executorID))
	// 			MakeCancel(t, client, strconv.Itoa(orderID))
	// 		})
	// 	}
	// })

	// t.Run("Invalid scenarios", func(t *testing.T) {
	// 	t.Run("Double assign", func(t *testing.T) {
	// 		MakeAssign(t, client, "40", "40")

	// 		code, err := client.AssignOrder("40", "40")
	// 		require.NoError(t, err)
	// 		require.Equal(t, http.StatusBadRequest, code)
	// 	})

	// 	t.Run("Cancel after acquire", func(t *testing.T) {
	// 		MakeAssign(t, client, "41", "41")
	// 		MakeAcquire(t, client, "41")

	// 		code, err := client.CancelOrder("41")
	// 		require.NoError(t, err)
	// 		require.Equal(t, http.StatusBadRequest, code)
	// 	})

	// 	t.Run("Cancel after cancel", func(t *testing.T) {
	// 		MakeAssign(t, client, "42", "42")
	// 		MakeCancel(t, client, "42")

	// 		code, err := client.CancelOrder("42")
	// 		require.NoError(t, err)
	// 		require.Equal(t, http.StatusBadRequest, code)
	// 	})

	// 	t.Run("Cancel non-existent order", func(t *testing.T) {
	// 		code, err := client.CancelOrder("non-existent")
	// 		require.NoError(t, err)
	// 		require.Equal(t, http.StatusBadRequest, code)
	// 	})

	// 	t.Run("Acquire non-assign order", func(t *testing.T) {
	// 		code, err := client.AcquireOrder("43")
	// 		require.NoError(t, err)
	// 		require.Equal(t, http.StatusBadRequest, code)
	// 	})

	// 	t.Run("Acquire order by non-existent executor", func(t *testing.T) {
	// 		code, err := client.AcquireOrder("non-existent")
	// 		require.NoError(t, err)
	// 		require.Equal(t, http.StatusBadRequest, code)
	// 	})

	// 	t.Run("Acquire after cancel", func(t *testing.T) {
	// 		MakeAssign(t, client, "44", "44")
	// 		MakeCancel(t, client, "44")

	// 		code, err := client.AcquireOrder("44")
	// 		require.NoError(t, err)
	// 		require.Equal(t, http.StatusBadRequest, code)
	// 	})
	// })

	// require.True(t, client.Down())
}
