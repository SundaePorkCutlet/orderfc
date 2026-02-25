package config

type Config struct {
	App      AppConfig      `yaml:"app" validate:"required"`
	Database DatabaseConfig `yaml:"database" validate:"required"`
	Redis    RedisConfig    `yaml:"redis" validate:"required"`
	Kafka    KafkaConfig    `yaml:"kafka" validate:"required"`
	Product  ProductConfig  `yaml:"product" validate:"required"`
	Tracing  TracingConfig  `yaml:"tracing"`
}

type TracingConfig struct {
	Endpoint    string `yaml:"endpoint" mapstructure:"endpoint"`
	ServiceName string `yaml:"service_name" mapstructure:"service_name"`
	Enabled     bool   `yaml:"enabled" mapstructure:"enabled"`
}

type AppConfig struct {
	Port string `yaml:"port" validate:"required"`
}

type ProductConfig struct {
	Host string `yaml:"host" validate:"required"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" validate:"required"`
	Port     string `yaml:"port" validate:"required"`
	User     string `yaml:"user" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	Name     string `yaml:"name" validate:"required"`
}

type RedisConfig struct {
	Host     string `yaml:"host" validate:"required"`
	Port     string `yaml:"port" validate:"required"`
	Password string `yaml:"password" validate:"required"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers" validate:"required"`
	Topic   string   `yaml:"topic" validate:"required"`
}
