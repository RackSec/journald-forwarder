package kafka

import (
	sarama "gopkg.in/Shopify/sarama.v1"
	"strings"
)

var producers []producerWithTopic

type producerWithTopic struct {
	Producer sarama.AsyncProducer
	Topic    string
}

// LoadProducers connects to the specified Kafka brokers,
// preparing asynchronous producers. If we fail to connect
// to any of the Kafka brokers, we return an error.
func LoadProducers(brokers string, topic string) error {

	// End early, because producers were already loaded
	if len(producers) != 0 {
		return nil
	}

	configs := readConfigs(brokers, topic)
	producers = make([]producerWithTopic, len(configs))

	for i, config := range configs {
		producer, err := newAsyncKafkaProducer(config.Brokers)
		if err != nil {
			return err
		}
		producers[i] = producerWithTopic{
			Producer: producer,
			Topic:    config.Topic,
		}
	}

	return nil
}

// Send produces an event on each provided Kafka producer. You should
// call LoadProducers before calling this, or else it won't have anywhere to
// send the message.
func Send(raw string) {
	for _, p := range producers {
		p.Producer.Input() <- &sarama.ProducerMessage{
			Topic: p.Topic,
			Value: sarama.StringEncoder([]byte(raw)),
		}
	}
}

func newAsyncKafkaProducer(brokers []string) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal

	return sarama.NewAsyncProducer(brokers, config)
}

// readConfigs takes argument variables to parse Kafka brokers and
// topics. Each producer requires both a broker and topic, although you can
// provide multiple brokers in a comma-delimited list. You can provide an
// arbitrary number of producers.
//
// Example:
//     brokers="192.168.100.50:9092"
//     ... or ...
//     brokers="192.168.100.51:9092,192.168.100.52:9092"
//
//	   topic="awesomeness"
func readConfigs(brokers string, topic string) []producerConfig {
	configs := make([]producerConfig, 0)

	configs = append(configs, producerConfig{
		Brokers: strings.Split(brokers, ","),
		Topic:   topic,
	})

	return configs
}

type producerConfig struct {
	Brokers []string
	Topic   string
}
