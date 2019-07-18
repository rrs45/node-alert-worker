package options

import (
	"flag"
	"time"
	"os"
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
	KubeAPIURL string
}

//GetConfig returna new config file
func GetConfig() *Config {
	return &Config{}
}

//AddFlags takes config file input
func (c *Config) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.File, "file", "/home/rajsingh/go/src/github.com/box-node-alert-worker/config/config.toml", "Configuration file path")
	fs.StringVar(&c.KubeAPIURL, "apiserver-override", "", "URL of the kubernetes api server")
}

//ValidOrDie validates some of the config parameters
func ValidOrDie(awo *viper.Viper) {
	log.Infof("Options - %+v",awo.AllSettings())
	_, err := time.ParseDuration(awo.GetString("general.cache_expire_interval"))
	if err != nil {
		log.Errorf("Options - Incorrect general.cache_expire_interval: %v ", err)
		log.Panic("Incorrect options")
	}
	if _, err1 := os.Stat(awo.GetString("scripts.dir")); os.IsNotExist(err) {
		log.Errorf("Options - Incorrect scripts.dir: %v ", err1)
		log.Panic("Incorrect options")
	}
}