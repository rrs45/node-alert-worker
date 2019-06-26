package options

import (
	"flag"
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//NewConfigFromFile parses config file  
func NewConfigFromFile(configFile string) (*viper.Viper, error) {
	//dir, file := filepath.Split(configFile)
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(configFile)
	//v.AddConfigPath(dir)
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}

//Config defines configuration parameters
type Config struct {
	File string
}

//GetConfig returna new config file
func GetConfig() *Config {
	return &Config{}
}

//AddFlags takes config file input
func (c *Config) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.File, "file", "/home/rajsingh/go/src/github.com/box-node-alert-worker/config/config.toml", "Configuration file path")
}

//ValidOrDie validates some of the config parameters
func ValidOrDie(ago *viper.Viper) {
	log.Infof("%+v",ago.AllSettings())
	_, err := time.ParseDuration(ago.GetString("general.cache_expire_interval"))
	if err != nil {
		log.Errorf("Options - Incorrect general.cache_expire_interval: %v ", err)
		log.Panic("Incorrect options")
	}
}