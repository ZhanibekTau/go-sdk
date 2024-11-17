package structures

type RabbitConfig struct {
	Host  string `mapstructure:"RABBITMQ_HOST"`
	Port  string `mapstructure:"RABBITMQ_PORT"`
	User  string `mapstructure:"RABBITMQ_USER"`
	Pass  string `mapstructure:"RABBITMQ_PASSWORD"`
	VHost string `mapstructure:"RABBITMQ_VHOST"`
}
