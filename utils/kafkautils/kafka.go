package kafkautils

import (
	"encoding/json"
	"testing"

	"github.com/twmb/franz-go/pkg/kadm"
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

func New(t *testing.T, cfg wkafka.Config, opts ...Option) *Kafka {
	t.Helper()

	o := option{
		WkafkaOpts: []wkafka.Option{wkafka.WithPingRetry(true)},
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
		case []byte:
			records = append(records, &wkafka.Record{Topic: topic, Value: v, Headers: headers})
		default:
			// Convert the message to a byte slice
			jsonData, err := json.Marshal(v)
			if err != nil {
				t.Fatal("failed to marshal message:", err)
			}

			records = append(records, &wkafka.Record{Topic: topic, Value: jsonData, Headers: headers})
		}
	}

	// Produce the records to Kafka
	if err := k.Client.ProduceRaw(t.Context(), records); err != nil {
		t.Fatal("failed to produce messages:", err)
	}
}
