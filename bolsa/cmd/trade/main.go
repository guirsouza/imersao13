package main

import (
	"encoding/json"
	"fmt"
	"sync"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/guirsouza/imersao13/go/internal/infra/kafka"
	"github.com/guirsouza/imersao13/go/internal/market/dto"
	"github.com/guirsouza/imersao13/go/internal/market/entity"
	"github.com/guirsouza/imersao13/go/internal/market/transformer"
)

func main() {

	ordersIn := make(chan *entity.Order)
	ordersOut := make(chan *entity.Order)

	waitGroup := &sync.WaitGroup{}
	defer waitGroup.Wait()

	kafkaMsgChannel := make(chan *ckafka.Message)
	producerConfigMap := &ckafka.ConfigMap{
		"bootstrap.servers": "host.docker.internal:9094",
	}
	consumerConfigMap := &ckafka.ConfigMap{
		"bootstrap.servers": "host.docker.internal:9094",
		"group.id":          "myGroup",
		"auto.offset.reset": "latest",
	}

	producer := kafka.NewKafkaProducer(producerConfigMap)
	consumer := kafka.NewConsumer(consumerConfigMap, []string{"input"})

	// thread 2
	go consumer.Consume(kafkaMsgChannel)

	// recebe do canal do kafka, manda pro input, processa e joga pro output, depois publca
	book := entity.NewBook(ordersIn, ordersOut, waitGroup)

	// thread 3
	go book.Trade()

	go func() {
		for msg := range kafkaMsgChannel {
			waitGroup.Add(1)
			fmt.Println(string(msg.Value))

			tradeInput := dto.TradeInput{}
			err := json.Unmarshal(msg.Value, &tradeInput)
			if err != nil {
				panic(err)
			}
			order := transformer.TransformInput(tradeInput)
			ordersIn <- order
		}
	}()

	for res := range ordersOut {
		output := transformer.TransformOutput(res)
		outputJson, err := json.MarshalIndent(output, "", "   ")
		fmt.Println(string(outputJson))

		if err != nil {
			fmt.Println(err)
		}

		producer.Publish(outputJson, []byte("orders"), "output")

	}
}
