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
		Secret            string `mapstructure:"secret"`
		ExpireHours       int    `mapstructure:"expire_hours"`
		RefreshExpireHours int   `mapstructure:"refresh_expire_hours"`
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
	viper.SetConfigName("config") // 配置文件名称(不带扩展名)
	viper.SetConfigType("yaml")   // 配置文件类型
	viper.AddConfigPath("./configs")   // 配置文件路径
	
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