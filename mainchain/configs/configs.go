package configs

import (
	"fmt"
	"github.com/spf13/viper"
)
var (
	FLConfFilePath string
	ClientID int
)

type Configs struct {
	FlConfigViper *viper.Viper
	HostConfigViper *viper.Viper
	ModelAccessViper *viper.Viper
	MeasureTimeViper *viper.Viper
}

var GlobalConfig = Configs{}

// ReadConfigs
// @Description: 读取所有配置文件
// @receiver c
// @param path
// @return error
func (c *Configs) ReadConfigs(path string) error {
	c.FlConfigViper = viper.New()
	c.FlConfigViper.SetConfigName("fl_conf")
	c.FlConfigViper.SetConfigType("json")
	c.FlConfigViper.AddConfigPath(path)
	c.FlConfigViper.AddConfigPath("./configs")
	c.FlConfigViper.AddConfigPath(".")
	if err := c.FlConfigViper.ReadInConfig(); err != nil {
		return fmt.Errorf("Fatal error config file: %w \n", err)
	}
	c.HostConfigViper = viper.New()
	c.HostConfigViper.SetConfigName("host_conf")
	c.HostConfigViper.SetConfigType("json")
	c.HostConfigViper.AddConfigPath(path)
	c.HostConfigViper.AddConfigPath("./configs")
	c.HostConfigViper.AddConfigPath(".")
	if err := c.HostConfigViper.ReadInConfig(); err != nil {
		return fmt.Errorf("Fatal error config file: %w \n", err)
	}
	c.ModelAccessViper = viper.New()
	c.ModelAccessViper.SetConfigName("model_access")
	c.ModelAccessViper.SetConfigType("json")
	c.MeasureTimeViper = viper.New()
	return nil
}
