package event

import (
	"context"
	"fmt"
	"os"

	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/solum-sp/aps-be-common/common/utils"
)

func CreateTopicIfNotExist(adminClient *kafka.AdminClient, topicName string, numPartitions int, replicationFactor int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		panic("ParseDuration(3s)")
	}
	// Create the topic if it does not exist
	results, err := utils.Retry(retryCount, retryInterval, func() ([]kafka.TopicResult, error) {
		return adminClient.CreateTopics(
			ctx,
			[]kafka.TopicSpecification{{
				Topic:             topicName,
				NumPartitions:     numPartitions,
				ReplicationFactor: replicationFactor,
			}},
			kafka.SetAdminOperationTimeout(maxDur))
	})
	if err != nil {
		fmt.Printf("Failed to create topic: %v\n", err)
		os.Exit(1)
	}
	// Print results
	for _, result := range results {
		fmt.Printf("Topic created: %v\n", result)
	}
	return nil
}
