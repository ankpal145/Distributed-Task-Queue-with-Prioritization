package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Configuration struct {
	Name    string
	Server  ServerConfig
	Redis   RedisConfig
	Worker  WorkerConfig
	Task    TaskConfig
	Metrics MetricsConfig
	Logging LoggingConfig
	Dynamic DynamicConfig
}

type ServerConfig struct {
	RESTPort int
	GRPCPort int
	Host     string
}

type RedisConfig struct {
	Host                string
	Port                int
	Password            string
	DB                  int
	PoolSize            int
	MinIdleConnections  int
	DialTimeout         time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	ClusterMode         bool
	ClusterNodes        []string
	PerNodePoolSize     int
	PerNodeMinIdleConns int
}

type WorkerConfig struct {
	Concurrency       int
	HeartbeatInterval time.Duration
	WorkerTimeout     time.Duration
	CleanupInterval   time.Duration
}

type TaskConfig struct {
	DefaultTimeout    time.Duration
	DefaultMaxRetries int
	RetryBaseDelay    time.Duration
	RetryMaxDelay     time.Duration
	CleanupInterval   time.Duration
}

type MetricsConfig struct {
	Enabled bool
	Port    int
	Path    string
}

type LoggingConfig struct {
	Level  string
	Format string
}

type DynamicConfig struct {
	EnableHealthCheck bool
	DebugMode         bool
}

var config *Configuration

func PanicIfError(err error) {
	if err != nil {
		panic(fmt.Errorf("unable to load config: %w", err))
	}
}

func CheckKey(key string) {
	if !viper.IsSet(key) {
		PanicIfError(fmt.Errorf("%s key is not set", key))
	}
}

func GetStringOrPanic(key string) string {
	CheckKey(key)
	return viper.GetString(key)
}

func GetIntOrPanic(key string) int {
	CheckKey(key)
	return viper.GetInt(key)
}

func GetInt64OrPanic(key string) int64 {
	CheckKey(key)
	return int64(viper.GetInt(key))
}

func GetBoolOrPanic(key string) bool {
	CheckKey(key)
	return viper.GetBool(key)
}

func GetDurationOrPanic(key string) time.Duration {
	CheckKey(key)
	return viper.GetDuration(key)
}

func GetStringArray(key string) []string {
	value := viper.GetString(key)
	if value == "" {
		return []string{}
	}
	stringArray := strings.Split(value, ",")
	for i, str := range stringArray {
		stringArray[i] = strings.TrimSpace(str)
	}
	return stringArray
}

func GetStringOrDefault(key string, defaultValue string) string {
	if !viper.IsSet(key) {
		return defaultValue
	}
	return viper.GetString(key)
}

func GetIntOrDefault(key string, defaultValue int) int {
	if !viper.IsSet(key) {
		return defaultValue
	}
	return viper.GetInt(key)
}

func GetBoolOrDefault(key string, defaultValue bool) bool {
	if !viper.IsSet(key) {
		return defaultValue
	}
	return viper.GetBool(key)
}

func GetDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if !viper.IsSet(key) {
		return defaultValue
	}
	return viper.GetDuration(key)
}

func Load(commandArgs []string) {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("name", "taskqueue")
	viper.AutomaticEnv()

	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.AddConfigPath("./profiles")
	viper.SetConfigType("yaml")

	if len(commandArgs) > 2 {
		configName := commandArgs[2]
		viper.SetConfigName(configName)

		err := viper.ReadInConfig()
		if err != nil {
			logrus.Fatalf("error while reading config file: %v", err)
			return
		}
	} else {
		viper.SetConfigName("config")
		if err := viper.ReadInConfig(); err != nil {
			logrus.Warnf("No config file found, using defaults: %v", err)
		}
	}

	if len(commandArgs) > 3 {
		loadDynamicConfig(commandArgs[3])
	}

	config = &Configuration{
		Name: GetStringOrDefault("name", "taskqueue"),
		Server: ServerConfig{
			RESTPort: GetIntOrDefault("server.rest_port", 8080),
			GRPCPort: GetIntOrDefault("server.grpc_port", 9090),
			Host:     GetStringOrDefault("server.host", "0.0.0.0"),
		},
		Redis: RedisConfig{
			Host:                GetStringOrDefault("redis.host", "localhost"),
			Port:                GetIntOrDefault("redis.port", 6379),
			Password:            GetStringOrDefault("redis.password", ""),
			DB:                  GetIntOrDefault("redis.db", 0),
			PoolSize:            GetIntOrDefault("redis.pool_size", 10),
			MinIdleConnections:  GetIntOrDefault("redis.min_idle_connections", 5),
			DialTimeout:         GetDurationOrDefault("redis.dial_timeout", 5*time.Second),
			ReadTimeout:         GetDurationOrDefault("redis.read_timeout", 3*time.Second),
			WriteTimeout:        GetDurationOrDefault("redis.write_timeout", 3*time.Second),
			ClusterMode:         GetBoolOrDefault("redis.cluster_mode", false),
			ClusterNodes:        GetStringArray("redis.cluster_nodes"),
			PerNodePoolSize:     GetIntOrDefault("redis.per_node_pool_size", 10),
			PerNodeMinIdleConns: GetIntOrDefault("redis.per_node_min_idle_conns", 5),
		},
		Worker: WorkerConfig{
			Concurrency:       GetIntOrDefault("worker.concurrency", 10),
			HeartbeatInterval: GetDurationOrDefault("worker.heartbeat_interval", 5*time.Second),
			WorkerTimeout:     GetDurationOrDefault("worker.worker_timeout", 30*time.Second),
			CleanupInterval:   GetDurationOrDefault("worker.cleanup_interval", 1*time.Minute),
		},
		Task: TaskConfig{
			DefaultTimeout:    GetDurationOrDefault("task.default_timeout", 5*time.Minute),
			DefaultMaxRetries: GetIntOrDefault("task.default_max_retries", 3),
			RetryBaseDelay:    GetDurationOrDefault("task.retry_base_delay", 1*time.Second),
			RetryMaxDelay:     GetDurationOrDefault("task.retry_max_delay", 5*time.Minute),
			CleanupInterval:   GetDurationOrDefault("task.cleanup_interval", 10*time.Minute),
		},
		Metrics: MetricsConfig{
			Enabled: GetBoolOrDefault("metrics.enabled", true),
			Port:    GetIntOrDefault("metrics.port", 2112),
			Path:    GetStringOrDefault("metrics.path", "/metrics"),
		},
		Logging: LoggingConfig{
			Level:  GetStringOrDefault("logging.level", "info"),
			Format: GetStringOrDefault("logging.format", "json"),
		},
		Dynamic: DynamicConfig{
			EnableHealthCheck: GetBoolOrDefault("dynamic.enable_health_check", true),
			DebugMode:         GetBoolOrDefault("dynamic.debug_mode", false),
		},
	}
}

func loadDynamicConfig(dynamicFilePath string) {
	viper.SetConfigName(dynamicFilePath)
	err := viper.MergeInConfig()
	if err != nil {
		logrus.Warnf("error while merging dynamic config file: %v", err)
		return
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		logrus.Infof("Dynamic configuration file changed: %s", in.Name)
		reloadDynamicConfig()
	})
}

func reloadDynamicConfig() {
	if config != nil {
		config.Dynamic.EnableHealthCheck = GetBoolOrDefault("dynamic.enable_health_check", true)
		config.Dynamic.DebugMode = GetBoolOrDefault("dynamic.debug_mode", false)
		logrus.Infof("Dynamic config reloaded: %+v", config.Dynamic)
	}
}

func NewConfig(module string) *Configuration {
	if config == nil {
		logrus.Fatal("Config not loaded. Call Load() first.")
	}
	cfg := *config
	cfg.Name = cfg.Name + "-" + module
	return &cfg
}

func GetConfig() *Configuration {
	if config == nil {
		logrus.Fatal("Config not loaded. Call Load() first.")
	}
	return config
}

func (c *Configuration) Validate() error {
	if c.Server.RESTPort < 1 || c.Server.RESTPort > 65535 {
		return fmt.Errorf("invalid REST port: %d", c.Server.RESTPort)
	}
	if c.Server.GRPCPort < 1 || c.Server.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", c.Server.GRPCPort)
	}
	if !c.Redis.ClusterMode && c.Redis.Host == "" {
		return fmt.Errorf("Redis host is required")
	}
	if c.Redis.ClusterMode && len(c.Redis.ClusterNodes) == 0 {
		return fmt.Errorf("Redis cluster nodes are required when cluster mode is enabled")
	}
	if c.Worker.Concurrency < 1 {
		return fmt.Errorf("worker concurrency must be at least 1")
	}
	return nil
}

func (c *Configuration) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}
