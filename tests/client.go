// Ну погнали нахуй

package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	sources "malomopa/internal/sources"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	AssignerAddress string
	ExecutorAddress string
	SourcesAddress  string
}

func NewDefaultClient() Client {
	return Client{
		AssignerAddress: "http://localhost:5252",
		ExecutorAddress: "http://localhost:5253",
		SourcesAddress:  "http://localhost:1337",
	}
}

func (c *Client) PingOrderAssigner() bool {
	resp, err := http.Get(c.AssignerAddress + "/ping")
	return err == nil && resp.StatusCode == http.StatusOK
}

func (c *Client) PingOrderExecutor() bool {
	resp, err := http.Get(c.ExecutorAddress + "/ping")
	return err == nil && resp.StatusCode == http.StatusOK
}

func (c *Client) PingSources() bool {
	resp, err := http.Get(c.SourcesAddress + "/ping")
	return err == nil && resp.StatusCode == http.StatusOK
}

// How to simulate network failure:
//
// docker network disconnect <NETWORK> <CONTAINER>
// docker network connect <NETWORK> <CONTAINER>

const (
	AssignerContainer  = "order_assigner"
	ExecutorContainter = "order_executor"
	SourcesContainer   = "fake_sources"
	Network            = "my_network"
)

var ScyllaNodesContainers = [...]string{"scylla-node1", "scylla-node2", "scylla-node3"}

func disconnectService(container string) bool {
	cmd := exec.Command("docker", "network", "disconnect", Network, container)
	_, err := cmd.CombinedOutput()
	return err == nil
}

func connectService(container string) bool {
	cmd := exec.Command("docker", "network", "connect", Network, container)
	_, err := cmd.CombinedOutput()
	return err == nil
}

// Что тут происходит?
//
//	Когда докер считает, что ноды поднялись, на самом деле им еще нужно время для создания кластера. В это время они недоступны
//	Я не знаю, если ли хороший способ проверить состояние кластера
//	Мой способ: исполнить команду `docker exec scylla-node1 nodetool status` и посмотреть, сколько в ней строчек
//	            Эта команда показывает в текстовом виде состояние кластера, а точнее то, собрала ли 1 нода кластер и кто в нем есть
//	            Если в команде 10 строчек, я считаю, что все ок. (У меня столько строчек в хорошем выводе)
//	            Также команда может вернуть ошибку (если нода еще не встала), поэтому на ошибку смотреть плохо, только логируем)
//	Как можно улучшить:
//	  1. Каким-то образом понимать, когда хотя бы первая нода встала, чтобы ошибки от команды можно было обрабатывать
//	  2. Нормально парсить вывод команды, а не смотреть на число строчек
func waitScylla() bool {
	time.Sleep(20 * time.Second)
	log.Println("Waiting scylla...")
	for {
		time.Sleep(10 * time.Second)
		cmd := exec.Command("sh", "-c", "docker exec scylla-node1 nodetool status | wc -l")
		output, err := cmd.Output()
		if err != nil {
			log.Printf("`docker exec -it scylla-node1 nodetool status | wc -l` got error: %s", err.Error())
			continue
		}

		outputStr := strings.TrimSpace(string(output))
		count, err := strconv.Atoi(outputStr)
		if err != nil {
			log.Printf("bad output of `docker exec -it scylla-node1 nodetool status | wc -l`: %s", outputStr)
			return false
		}
		if count == 10 { // ?? XD
			time.Sleep(20 * time.Second)
			log.Printf("got `wc -l` = %v, returning", count)
			return true
		}
	}
}

// Накатываем миграцию (создаем БД и табличку)
func migrateData() bool {
	for it := 0; it < 10; it++ {
		cmd := exec.Command("docker", "exec", ScyllaNodesContainers[0], "cqlsh", "-f", "/mutant-data.txt")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Migration failed: %s", string(output))
			time.Sleep(10 * time.Second)
			continue
		}
		return true
	}
	return false
}

// Поднимаем все сервисы и ждем базу с миграцией
func (c *Client) Start() bool {
	cmd := exec.Command("docker", "compose", "build")
	err := cmd.Run()
	if err != nil {
		return false
	}
	cmd = exec.Command("docker", "compose", "up")
	err = cmd.Start()
	if err != nil {
		return false
	}
	if !waitScylla() {
		log.Fatal("Scylla hasn't been wake up :C")
		return false
	}
	if !migrateData() {
		log.Fatal("Data migration failed :C")
		return false
	}
	return true
}

// Останавливаем вообще все
func (c *Client) Down() bool {
	cmd := exec.Command("docker", "compose", "down")
	_, err := cmd.CombinedOutput()
	return err == nil
}

func (c *Client) Restart() bool {
	c.Down()
	return c.Start()
}

func (c *Client) DisconnectSources() bool {
	return disconnectService(SourcesContainer)
}

func (c *Client) ConnectSources() bool {
	return connectService(SourcesContainer)
}

func (c *Client) DisconnectNode(id int) bool {
	return disconnectService(ScyllaNodesContainers[id])
}

func (c *Client) ConnectNode(id int) bool {
	return connectService(ScyllaNodesContainers[id])
}

func (c *Client) TurnOffConfigsSource() bool {
	resp, err := http.Post(c.SourcesAddress+"/configs_off", "application/json", nil)
	if err != nil {
		log.Printf("Turning off `Configs` returned error: %s", err.Error())
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func (c *Client) TurnOnConfigsSource() bool {
	resp, err := http.Post(c.SourcesAddress+"/configs_on", "application/json", nil)
	if err != nil {
		log.Printf("Turning on `Configs` returned error: %s", err.Error())
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func (c *Client) TurnOffZonesInfoSource() bool {
	resp, err := http.Post(c.SourcesAddress+"/zone_info_off", "application/json", nil)
	if err != nil {
		log.Printf("Turning off `ZonesInfo` returned error: %s", err.Error())
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func (c *Client) TurnOnZonesInfoSource() bool {
	resp, err := http.Post(c.SourcesAddress+"/zone_info_on", "application/json", nil)
	if err != nil {
		log.Printf("Turning on `ZonesInfo` returned error: %s", err.Error())
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func (c *Client) SourceCounters() (*sources.HandlersCountersResponse, error) {
	resp, err := http.Get(c.SourcesAddress + "/counters")
	if err != nil {
		log.Printf("`GetCounters` returned error: %s", err.Error())
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("`GetCounters` returned ", resp.StatusCode)
		return nil, errors.New("Counter got not 200")
	}
	decoder := json.NewDecoder(resp.Body)
	var counters sources.HandlersCountersResponse
	err = decoder.Decode(&counters)
	if err != nil {
		return nil, err
	}
	return &counters, nil
}

type AcquireResponse struct {
	orderPayload *OrderPayload
	code         int
}

func (c *Client) AcquireOrder(executorID string) (*AcquireResponse, error) {
	resp, err := http.Post(fmt.Sprintf("%s/v1/acquire_order?executor-id=%s", c.ExecutorAddress, executorID), "application/json", nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return &AcquireResponse{code: resp.StatusCode}, nil
	}
	decoder := json.NewDecoder(resp.Body)
	var payload OrderPayload
	err = decoder.Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &AcquireResponse{
		orderPayload: &payload,
		code:         resp.StatusCode,
	}, nil
}

type CancelResponse struct {
	orderPayload *OrderPayload
	code         int
}

func (c *Client) CancelOrder(orderID string) (*CancelResponse, error) {
	resp, err := http.Post(fmt.Sprintf("%s/v1/cancel_order?order-id=%s", c.AssignerAddress, orderID), "application/json", nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return &CancelResponse{code: resp.StatusCode}, nil
	}
	decoder := json.NewDecoder(resp.Body)
	var payload OrderPayload
	err = decoder.Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &CancelResponse{
		orderPayload: &payload,
		code:         resp.StatusCode,
	}, nil
}

func (c *Client) AssignOrder(orderID, executorID string) (int, error) {
	resp, err := http.Post(fmt.Sprintf("%s/v1/assign_order?order-id=%s&executor-id=%s", c.AssignerAddress, orderID, executorID), "application/json", nil)
	return resp.StatusCode, err
}
