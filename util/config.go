package util

/*
	使用viper从配置文件app.env中读取参数
*/
import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver                   string        `mapstructure:"DB_DRIVER"`
	DBSource                   string        `mapstructure:"DB_SOURCE"`
	ServerAddress              string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey          string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDurationMinutes time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

// 加载环境变量/配置文件，并将其解析到 Config 结构体中
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv() // 自动从环境变量中读取配置,Viper会优先使用环境变量的值

	err = viper.ReadInConfig() // 读取配置文件
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config) // 将读取到的配置内容反序列化为 Config 结构体

	return
}
