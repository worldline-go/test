package kafkautils

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/worldline-go/wkafka"
)

type Kafka struct {
	Client *wkafka.Client
	Admin  *kadm.Client
	Config wkafka.Config
}

type KafkaTest struct {
	*Kafka
}

type Topic struct {
	Name string

	// Partitions defaults to -1, which means the broker default.
	Partitions int32
	// ReplicationFactor defaults to -1, which means the broker default.
	ReplicationFactor int16
}

type ModifiedPartitioner struct {
	kgo.Partitioner
}

func (p *ModifiedPartitioner) ForTopic(string) kgo.TopicPartitioner {
	return &ModifiedTopicPartitioner{
		TopicPartitioner: p.Partitioner.ForTopic(""),
	}
}

type ModifiedTopicPartitioner struct {
	kgo.TopicPartitioner
}

func (p *ModifiedTopicPartitioner) PartitionByBackup(r *kgo.Record, n int, backup kgo.TopicBackupIter) int {
	if r.Partition == -1 {
		// If partition is -1, use the round robin batch-wise partitioner
		if partitioner, _ := p.TopicPartitioner.(kgo.TopicBackupPartitioner); partitioner != nil {
			return partitioner.PartitionByBackup(r, n, backup)
		}
	}

	// If partition is >= 0, use the specified partition ID
	return int(r.Partition)
}

func New(ctx context.Context, cfg wkafka.Config, opts ...Option) (*Kafka, error) {
	partitoner := ModifiedPartitioner{kgo.UniformBytesPartitioner(64<<10, true, true, nil)}

	o := option{
		WkafkaOpts: []wkafka.Option{
			wkafka.WithPingRetry(true),
			wkafka.WithKGOOptions(
				// Use custom partitioner that treats
				// - PartitionID = -1 just like the kgo.StickyKeyPartitioner() would do (round robin batch-wise)
				// - PartitionID >= 0 Use the partitionID as specified in the record struct
				kgo.RecordPartitioner(&partitoner),
			),
		},
	}

	o.apply(opts...)

	client, err := wkafka.New(ctx, cfg, o.WkafkaOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	return &Kafka{
		Client: client,
		Admin:  kadm.NewClient(client.Kafka),
		Config: cfg,
	}, nil
}

func NewTest(t *testing.T, cfg wkafka.Config, opts ...Option) *KafkaTest {
	t.Helper()

	k, err := New(t.Context(), cfg, opts...)
	if err != nil {
		t.Fatal(err)
	}

	return &KafkaTest{
		Kafka: k,
	}
}

// ///////////////////////////////////////////////////////////////////

func (k *KafkaTest) DeleteGroups(t *testing.T, groups ...string) {
	if err := k.Kafka.deleteGroups(t, t.Context(), groups...); err != nil {
		t.Fatal(err)
	}
}

func (k *Kafka) DeleteGroups(ctx context.Context, groups ...string) error {
	return k.deleteGroups(nil, ctx, groups...)
}

func (k *Kafka) deleteGroups(t *testing.T, ctx context.Context, groups ...string) error {
	if t != nil {
		t.Helper()
	}

	if _, err := k.Admin.DeleteGroups(ctx, groups...); err != nil {
		return fmt.Errorf("failed to delete groups: %w", err)
	}

	return nil
}

func (k *KafkaTest) DeleteTopics(t *testing.T, topics ...string) {
	if err := k.Kafka.deleteTopics(t, t.Context(), topics...); err != nil {
		t.Fatal(err)
	}
}

func (k *Kafka) DeleteTopics(ctx context.Context, topics ...string) error {
	return k.deleteTopics(nil, ctx, topics...)
}

func (k *Kafka) deleteTopics(t *testing.T, ctx context.Context, topics ...string) error {
	if t != nil {
		t.Helper()
	}

	if _, err := k.Admin.DeleteTopics(ctx, topics...); err != nil {
		return fmt.Errorf("failed to delete topics: %w", err)
	}

	return nil
}

func (k *KafkaTest) CreateTopics(t *testing.T, topics ...Topic) []kadm.CreateTopicResponse {
	responses, err := k.Kafka.createTopics(t, t.Context(), topics...)
	if err != nil {
		t.Fatal(err)
	}

	return responses
}

func (k *Kafka) CreateTopics(ctx context.Context, topics ...Topic) ([]kadm.CreateTopicResponse, error) {
	return k.createTopics(nil, ctx, topics...)
}

func (k *Kafka) createTopics(t *testing.T, ctx context.Context, topics ...Topic) ([]kadm.CreateTopicResponse, error) {
	if t != nil {
		t.Helper()
	}

	responses := make([]kadm.CreateTopicResponse, 0, len(topics))
	for _, topic := range topics {
		partitions := topic.Partitions
		if partitions == 0 {
			partitions = -1
		}

		replicationFactor := topic.ReplicationFactor
		if replicationFactor == 0 {
			replicationFactor = -1
		}

		response, err := k.Admin.CreateTopic(ctx, partitions, replicationFactor, nil, topic.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to create topic %s: %w", topic.Name, err)
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// Publish publishes messages to the specified topic.
//   - If the message is a byte slice, it will be sent as is.
//   - If the message is any other type, it will be marshaled to JSON.
//   - If the message is a wkafka.Record, than parition field is used, set to -1 to use the round robin batch partitioner.
func (k *KafkaTest) Publish(t *testing.T, topic string, messages ...any) {
	if err := k.Kafka.publish(t, t.Context(), topic, messages...); err != nil {
		t.Fatal(err)
	}
}

// Publish publishes messages to the specified topic.
//   - If the message is a byte slice, it will be sent as is.
//   - If the message is any other type, it will be marshaled to JSON.
//   - If the message is a wkafka.Record, than parition field is used, set to -1 to use the round robin batch partitioner.
func (k *Kafka) Publish(ctx context.Context, topic string, messages ...any) error {
	return k.publish(nil, ctx, topic, messages...)
}

func (k *Kafka) publish(t *testing.T, ctx context.Context, topic string, messages ...any) error {
	if t != nil {
		t.Helper()
	}

	records := make([]*wkafka.Record, 0, len(messages))
	headers := []wkafka.Header{
		{
			Key:   wkafka.HeaderServiceKey,
			Value: k.Client.ClientID(),
		},
	}

	for _, msg := range messages {
		switch v := msg.(type) {
		case wkafka.Record:
			v.Headers = append(v.Headers, headers...)
			v.Topic = topic

			records = append(records, &v)
		case []byte:
			records = append(records, &wkafka.Record{Topic: topic, Value: v, Headers: headers, Partition: -1})
		default:
			// Convert the message to a byte slice
			jsonData, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("failed to marshal message: %w", err)
			}

			records = append(records, &wkafka.Record{Topic: topic, Value: jsonData, Headers: headers, Partition: -1})
		}
	}

	// Produce the records to Kafka
	if err := k.Client.ProduceRaw(ctx, records); err != nil {
		return fmt.Errorf("failed to produce messages: %w", err)
	}

	return nil
}
