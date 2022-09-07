package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const configFileName = "config.yml"

type ApiGRPCServer struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	TimeoutConnection int    `yaml:"timeoutConnection"`
}

type ApiHTTPServer struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type RMQ struct {
	Use      bool   `yaml:"use"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
	Queue    string `yaml:"queue"`
}

type Kafka struct {
	Host  string `yaml:"host"`
	Port  string `yaml:"port"`
	Topic string `yaml:"topic"`
}

type Config struct {
	ApiGRPCServer ApiGRPCServer `yaml:"apiGRPCServer"`
	ApiHTTPServer ApiHTTPServer `yaml:"apiHTTPServer"`
	RMQ           RMQ           `yaml:"rmq"`
	Kafka         Kafka         `yaml:"kafka"`
}

func NewConfig() (*Config, error) {
	var cfg = new(Config)
	if err := cfg.initFromFile(configFileName); err != nil {
		return nil, err
	}

	return cfg, nil
}

// initFromFile init Config from yml file.
func (c *Config) initFromFile(filePath string) error {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("config file close: %v", err)
		}
	}()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&c); err != nil {
		return err
	}

	return nil
}

func (c *Config) GetApiGRPCServerAddr() string {
	return fmt.Sprintf("%v:%v", c.ApiGRPCServer.Host, c.ApiGRPCServer.Port)
}

func (c *Config) GetApiGRPCServerTimeout() time.Duration {
	timeout := time.Duration(c.ApiGRPCServer.TimeoutConnection) * time.Second
	return timeout
}

func (c *Config) GetApiHTTPServerAddr() string {
	return fmt.Sprintf("%v:%v", c.ApiHTTPServer.Host, c.ApiHTTPServer.Port)
}

// GetRMQuri return connect uri in format: amqp://guest:guest@localhost:5672/
func (c *Config) GetRMQuri() string {
	return fmt.Sprintf("amqp://%v:%v@%v:%v/", c.RMQ.Login, c.RMQ.Password, c.RMQ.Host, c.RMQ.Port)
}

func (c *Config) GetRMQqueue() string {
	return c.RMQ.Queue
}

func (c *Config) GetKafkaUri() string {
	return fmt.Sprintf("%v:%v", c.Kafka.Host, c.Kafka.Port)
}

func (c *Config) GetKafkaTopic() string {
	return c.Kafka.Topic
}
