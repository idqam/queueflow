package autoscaler

import (
	"context"
	"errors"

	"owen/queueflow/internal/kafka"

	kafkago "github.com/segmentio/kafka-go"
)

type KafkaLagFetcher struct {
	Client  *kafkago.Client
	GroupID string
	Topics  []string
}

func NewKafkaFetcher(brokers []string, groupID string, topics []string) (*KafkaLagFetcher, error) {
	if len(brokers) == 0 {
		return nil, errors.New("kafka brokers required")
	}
	if groupID == "" {
		return nil, errors.New("kafka groupID required")
	}
	if len(topics) == 0 {
		topics = kafka.AllJobTopics
	}

	client := &kafkago.Client{Addr: kafkago.TCP(brokers...)}

	return &KafkaLagFetcher{
		Client:  client,
		GroupID: groupID,
		Topics:  topics,
	}, nil
}

func (f *KafkaLagFetcher) FetchLag(ctx context.Context) (map[string]int64, int64, error) {
	if f == nil || f.Client == nil {
		return nil, 0, errors.New("kafka lag fetcher not initialized")
	}

	if len(f.Topics) == 0 {
		f.Topics = kafka.AllJobTopics
	}

	metaResp, err := f.Client.Metadata(ctx, &kafkago.MetadataRequest{Topics: f.Topics})
	if err != nil {
		return nil, 0, err
	}

	partitionMap := make(map[string][]int)
	for _, t := range metaResp.Topics {
		if t.Error != nil {
			continue
		}
		if len(t.Partitions) == 0 {
			continue
		}
		for _, p := range t.Partitions {
			partitionMap[t.Name] = append(partitionMap[t.Name], p.ID)
		}
	}

	if len(partitionMap) == 0 {
		return nil, 0, errors.New("no kafka topic partitions found")
	}

	listOffsetsRequest := &kafkago.ListOffsetsRequest{Topics: map[string][]kafkago.OffsetRequest{}}
	for topic, partitions := range partitionMap {
		for _, partition := range partitions {
			listOffsetsRequest.Topics[topic] = append(listOffsetsRequest.Topics[topic], kafkago.LastOffsetOf(partition))
		}
	}

	latest, err := f.Client.ListOffsets(ctx, listOffsetsRequest)
	if err != nil {
		return nil, 0, err
	}

	committed, err := f.Client.OffsetFetch(ctx, &kafkago.OffsetFetchRequest{GroupID: f.GroupID, Topics: partitionMap})
	if err != nil {
		return nil, 0, err
	}

	perTopicLag := make(map[string]int64)
	var totalLag int64

	for topic, partitions := range latest.Topics {
		for _, p := range partitions {
			if p.Error != nil {
				continue
			}

			committedOffset := int64(0)
			if committedTopic, ok := committed.Topics[topic]; ok {
				for _, cp := range committedTopic {
					if cp.Partition == p.Partition {
						committedOffset = cp.CommittedOffset
						break
					}
				}
			}

			lag := p.LastOffset - committedOffset
			if lag < 0 {
				lag = 0
			}

			perTopicLag[topic] += lag
			totalLag += lag
		}
	}

	return perTopicLag, totalLag, nil
}

func (f *KafkaLagFetcher) FetchTotalLag(ctx context.Context) (int64, error) {
	_, total, err := f.FetchLag(ctx)
	return total, err
}

