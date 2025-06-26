package kafkautils

import (
	"encoding/json"
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

func New(t *testing.T, cfg wkafka.Config, opts ...Option) *Kafka {
	t.Helper()

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

	client, err := wkafka.New(t.Context(), cfg, o.WkafkaOpts...)
	if err != nil {
		t.Fatal("failed to create Kafka client:", err)
	}

	return &Kafka{
		Client: client,
		Admin:  kadm.NewClient(client.Kafka),
		Config: cfg,
	}
}

// ///////////////////////////////////////////////////////////////////

func (k *Kafka) DeleteGroups(t *testing.T, groups ...string) {
	t.Helper()
	_, err := k.Admin.DeleteGroups(t.Context(), groups...)

	if err != nil {
		t.Fatal("failed to delete groups:", err)
	}
}

func (k *Kafka) DeleteTopics(t *testing.T, topics ...string) {
	t.Helper()
	_, err := k.Admin.DeleteTopics(t.Context(), topics...)

	if err != nil {
		t.Fatal("failed to delete topics:", err)
	}
}

func (k *Kafka) CreateTopics(t *testing.T, topics ...Topic) []kadm.CreateTopicResponse {
	t.Helper()

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

		response, err := k.Admin.CreateTopic(t.Context(), partitions, replicationFactor, nil, topic.Name)
		if err != nil {
			t.Fatal("failed to create topic:", err)
		}

		responses = append(responses, response)
	}

	return responses
}

// Publish publishes messages to the specified topic.
//   - If the message is a byte slice, it will be sent as is.
//   - If the message is any other type, it will be marshaled to JSON.
//   - If the message is a wkafka.Record, than parition field is used, set to -1 to use the round robin batch partitioner.
func (k *Kafka) Publish(t *testing.T, topic string, messages ...any) {
	t.Helper()

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
				t.Fatal("failed to marshal message:", err)
			}

			records = append(records, &wkafka.Record{Topic: topic, Value: jsonData, Headers: headers, Partition: -1})
		}
	}

	// Produce the records to Kafka
	if err := k.Client.ProduceRaw(t.Context(), records); err != nil {
		t.Fatal("failed to produce messages:", err)
	}
}
