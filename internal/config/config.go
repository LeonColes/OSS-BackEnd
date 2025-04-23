package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server struct {
		Port string `mapstructure:"port"`
		Mode string `mapstructure:"mode"`
	} `mapstructure:"server"`

	Database struct {
		Driver          string `mapstructure:"driver"`
		DSN             string `mapstructure:"dsn"`
		MaxIdleConns    int    `mapstructure:"max_idle_conns"`
		MaxOpenConns    int    `mapstructure:"max_open_conns"`
		ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	} `mapstructure:"database"`

	Redis struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`

	MinIO struct {
		Endpoint       string `mapstructure:"endpoint"`
		AccessKey      string `mapstructure:"access_key"`
		SecretKey      string `mapstructure:"secret_key"`
		UseSSL         bool   `mapstructure:"use_ssl"`
		BucketLocation string `mapstructure:"bucket_location"`
	} `mapstructure:"minio"`

	JWT struct {
		Secret             string `mapstructure:"secret"`
		ExpireHours        int    `mapstructure:"expire_hours"`
		RefreshExpireHours int    `mapstructure:"refresh_expire_hours"`
	} `mapstructure:"jwt"`

	Storage struct {
		UploadPath   string   `mapstructure:"upload_path"`
		TempPath     string   `mapstructure:"temp_path"`
		MaxFileSize  int64    `mapstructure:"max_file_size"`
		AllowedTypes []string `mapstructure:"allowed_types"`
	} `mapstructure:"storage"`

	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
		Output string `mapstructure:"output"`
	} `mapstructure:"log"`
}

var config *Config

// 加载配置文件
func Load() (*Config, error) {
	// 若已加载配置，则直接返回
	if config != nil {
		return config, nil
	}

	// 设置配置文件路径
	viper.SetConfigName("config")    // 配置文件名称(不带扩展名)
	viper.SetConfigType("yaml")      // 配置文件类型
	viper.AddConfigPath("./configs") // 配置文件路径

	// 尝试读取开发环境配置
	if _, err := os.Stat("./configs/config.dev.yaml"); err == nil {
		viper.SetConfigName("config.dev")
	}

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// 解析配置到结构体
	config = &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	// 从环境变量加载配置
	loadFromEnv()

	return config, nil
}

// 从环境变量加载配置
func loadFromEnv() {
	// 数据库配置
	if host := os.Getenv("DB_HOST"); host != "" {
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "3306"
		}
		user := os.Getenv("DB_USER")
		if user == "" {
			user = "root"
		}
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "oss_system"
		}

		// 重新构建DSN
		config.Database.DSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True",
			user, password, host, port, dbname)
	}

	// Redis配置
	if host := os.Getenv("REDIS_HOST"); host != "" {
		port := os.Getenv("REDIS_PORT")
		if port == "" {
			port = "6379"
		}
		config.Redis.Addr = fmt.Sprintf("%s:%s", host, port)

		if password := os.Getenv("REDIS_PASSWORD"); password != "" {
			config.Redis.Password = password
		}

		if db := os.Getenv("REDIS_DB"); db != "" {
			if dbInt, err := strconv.Atoi(db); err == nil {
				config.Redis.DB = dbInt
			}
		}
	}

	// MinIO配置
	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		config.MinIO.Endpoint = endpoint
	}
	if accessKey := os.Getenv("MINIO_ACCESS_KEY"); accessKey != "" {
		config.MinIO.AccessKey = accessKey
	}
	if secretKey := os.Getenv("MINIO_SECRET_KEY"); secretKey != "" {
		config.MinIO.SecretKey = secretKey
	}
	if useSSL := os.Getenv("MINIO_USE_SSL"); useSSL != "" {
		config.MinIO.UseSSL = strings.ToLower(useSSL) == "true"
	}

	// JWT密钥
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.JWT.Secret = secret
	}
}

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	v := viper.New()

	// 设置配置文件路径
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// 读取环境变量
	v.AutomaticEnv()
	v.SetEnvPrefix("OSS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认配置
			fmt.Println("警告: 配置文件未找到，使用默认配置")
		} else {
			return nil, fmt.Errorf("读取配置文件错误: %w", err)
		}
	}

	// 加载配置到结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置错误: %w", err)
	}

	// 设置默认值
	setDefaults(&cfg)

	return &cfg, nil
}

// setDefaults 设置配置默认值
func setDefaults(cfg *Config) {
	// 服务器默认配置
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}

	// 数据库默认配置
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 100
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 3600 // 默认1小时
	}

	// JWT默认配置
	if cfg.JWT.ExpireHours == 0 {
		cfg.JWT.ExpireHours = 24 // 默认24小时
	}
	if cfg.JWT.RefreshExpireHours == 0 {
		cfg.JWT.RefreshExpireHours = 168 // 默认7天
	}

	// 存储默认配置
	if cfg.Storage.UploadPath == "" {
		cfg.Storage.UploadPath = "./uploads"
	}
	if cfg.Storage.TempPath == "" {
		cfg.Storage.TempPath = "./temp"
	}
	if cfg.Storage.MaxFileSize == 0 {
		cfg.Storage.MaxFileSize = 10 * 1024 * 1024 // 默认10MB
	}

	// 日志默认配置
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "text"
	}
	if cfg.Log.Output == "" {
		cfg.Log.Output = "stdout"
	}
}
