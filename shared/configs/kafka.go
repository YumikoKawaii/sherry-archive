package configs

type KafkaConfig struct {
	Brokers        []string `name:"brokers" help:"Kafka brokers to write logs" env:"KAFKA_BROKERS" default:"localhost:9092" yaml:"brokers"`
	UseCompression bool     `name:"use-compression" help:"Turn on/off Kafka compression" env:"KAFKA_USE_COMPRESSION" default:"true" yaml:"use-compression"`
}
