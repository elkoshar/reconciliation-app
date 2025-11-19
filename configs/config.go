package config

import (
	"github.com/spf13/viper"
)

var (
	config *Config
)

// option defines configuration option
type option struct {
	configFolder string
	configFile   string
	configType   string
}

// Init initializes `config` from the default config file.
// use `WithConfigFile` to specify the location of the config file
func Init(opts ...Option) error {
	opt := &option{
		configFolder: getDefaultConfigFolder(),
		configFile:   getDefaultConfigFile(),
		configType:   getDefaultConfigType(),
	}

	for _, optFunc := range opts {
		optFunc(opt)
	}

	// Config File Path
	viper.AddConfigPath(opt.configFolder)
	// Config File Name
	viper.SetConfigName(opt.configFile)
	// Config File Type
	viper.SetConfigType(opt.configType)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	// Reading variables without using the model for Kube SecretKey
	config = new(Config)

	//set default value for all config
	setDefault()

	err = viper.Unmarshal(&config)
	if err != nil {
		return err
	}

	return config.postprocess()
}

// Option define an option for config package
type Option func(*option)

// WithConfigFolder set `config` to use the given config folder
func WithConfigFolder(configFolder string) Option {
	return func(opt *option) {
		opt.configFolder = configFolder
	}
}

// WithConfigFile set `config` to use the given config file
func WithConfigFile(configFile string) Option {
	return func(opt *option) {
		opt.configFile = configFile
	}
}

// WithConfigType set `config` to use the given config type
func WithConfigType(configType string) Option {
	return func(opt *option) {
		opt.configType = configType
	}
}

// getDefaultConfigFolder get default config folder.
func getDefaultConfigFolder() string {
	configPath := "./configs/"

	return configPath
}

// getDefaultConfigFile get default config file.
func getDefaultConfigFile() string {
	return "config"
}

// getDefaultConfigType get default config type.
func getDefaultConfigType() string {
	return "yaml"
}

// Get config
func Get() *Config {
	if config == nil {
		config = &Config{}
	}
	return config
}

func setDefault() {
	viper.SetDefault("HTTP_READ_TIMEOUT", "10s")
	viper.SetDefault("HTTP_WRITE_TIMEOUT", "10s")
	viper.SetDefault("HTTP_INBOUND_TIMEOUT", "10s")
	viper.SetDefault("HTTP_TIMEOUT", "10s")
	viper.SetDefault("HTTP_DEBUG", true)
	viper.SetDefault("HTTP_MAX_IDLE_CONNECTIONS", 100)
	viper.SetDefault("HTTP_MAX_IDLE_CONNECTIONS_PER_HOST", 100)
	viper.SetDefault("HTTP_IDLE_CONNECTION_TIMEOUT", "10s")
}

// postprocess several config
func (c *Config) postprocess() error {
	return nil
}
