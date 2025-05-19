package configs

type KafkaConfig struct {
	Brokers        []string `name:"brokers" help:"Kafka brokers to write logs" env:"KAFKA_BROKERS" default:"localhost:9092" yaml:"brokers"`
	UseCompression bool     `name:"use-compression" help:"Turn on/off Kafka compression" env:"KAFKA_USE_COMPRESSION" default:"true" yaml:"use-compression"`
	GroupId        string   `name:"group-id" help:"Consumer group id" env:"CONSUMER_GROUP_ID" default:"sherry-archive" yaml:"group-id"`
	BatchSize      int      `name:"consume-batch-size" help:"Consumer batch size" env:"CONSUMER_BATCH_SIZE" default:"1000" yaml:"batch-size"`
	CycleInSec     int      `name:"consume-cycle-in-sec" help:"Consumer cycle in sec" env:"CONSUMER_CYCLE_IN_SEC" default:"5" yaml:"cycle-in-sec"`
}
