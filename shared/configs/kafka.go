package configs

type KafkaConfig struct {
	Brokers        []string `name:"brokers" help:"Kafka brokers to write logs" env:"KAFKA_BROKERS" default:"localhost:9092" yaml:"brokers"`
	Version        string   `name:"kafka-version" help:"Kafka version" env:"KAFKA_VERSION" default:"2.1.1" yaml:"version"`
	Verbose        bool     `name:"verbose" help:"Turn on/off Sarama logging" env:"KAFKA_VERBOSE" default:"false" yaml:"verbose"`
	UseCompression bool     `name:"use-compression" help:"Turn on/off Kafka compression" env:"KAFKA_USE_COMPRESSION" default:"true" yaml:"use-compression"`
}
