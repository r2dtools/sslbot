package config

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/unknwon/com"
	"gopkg.in/yaml.v3"
)

const (
	defaultPort               = 60150
	defaultCaServer           = "https://acme-v02.api.letsencrypt.org/directory"
	defaultVarDir             = "/usr/local/r2dtools/sslbot/var"
	defaultCertBotDataDir     = "/etc/letsencrypt/live"
	defaultCertBotBin         = "certbot"
	defaultNginxRoot          = "/etc/nginx"
	defaultNginxAcmeCommonDir = "/var/www/html/"
)

var isDevMode = true
var Version string

type Config struct {
	LogFile            string
	Port               int
	Token              string
	IsDevMode          bool
	Version            string
	LegoBin            string
	CaServer           string
	ConfigFilePath     string
	VarDir             string
	CertBotEnabled     bool
	CertBotBin         string
	CertBotWokrDir     string
	NginxAcmeCommonDir string
	rootPath           string
}

func GetConfig() (*Config, error) {
	var rootPath string

	if isDevMode {
		wd, err := os.Getwd()

		if err != nil {
			return nil, err
		}

		rootPath = wd

		if filepath.Base(wd) == "cmd" {
			rootPath = filepath.Dir(wd)
		}
	} else {
		executable, err := os.Executable()

		if err != nil {
			return nil, err
		}

		rootPath = filepath.Dir(executable)
	}

	configFilePath := filepath.Join(rootPath, "config.yaml")

	viper.AddConfigPath(filepath.Dir(configFilePath))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("sslbot")

	viper.SetDefault(PortOpt, defaultPort)
	viper.SetDefault(CaServerOpt, defaultCaServer)
	viper.SetDefault(VarDirOpt, defaultVarDir)
	viper.SetDefault(CertBotWorkDirOpt, defaultCertBotDataDir)
	viper.SetDefault(CertBotBinOpt, defaultCertBotBin)
	viper.SetDefault(NginxAcmeCommonDirOpt, defaultNginxAcmeCommonDir)
	viper.SetDefault(NginxRootOpt, defaultNginxRoot)

	if com.IsFile(configFilePath) {
		configFile, err := os.OpenFile(configFilePath, os.O_RDONLY, 0644)

		if err != nil {
			panic(err)
		}

		defer configFile.Close()

		if err := viper.ReadConfig(configFile); err != nil {
			panic(err)
		}
	}

	if Version == "" {
		Version = "dev"
	}

	if isDevMode {
		viper.Set(VarDirOpt, filepath.Join(rootPath, "var"))
	}

	config := &Config{
		LogFile:        filepath.Join(rootPath, "sslbot.log"),
		LegoBin:        filepath.Join(rootPath, "lego"),
		ConfigFilePath: configFilePath,
		rootPath:       rootPath,
		IsDevMode:      isDevMode,
		Version:        Version,
	}
	setDynamicParams(config)

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		setDynamicParams(config)
	})

	return config, nil
}

func (c *Config) GetPathInsideVarDir(path ...string) string {
	parts := []string{c.VarDir}
	parts = append(parts, path...)

	return filepath.Join(parts...)
}

func (c *Config) ToMap() map[string]string {
	settings := viper.AllSettings()
	options := make(map[string]string)

	for key, value := range settings {
		if strValue, ok := value.(string); ok {
			options[key] = strValue
		}
	}

	return options
}

func (c *Config) SetParam(name string, value any) error {
	data, err := os.ReadFile(c.ConfigFilePath)

	if err != nil {
		return err
	}

	confMap := make(map[string]any)
	err = yaml.Unmarshal(data, confMap)

	if err != nil {
		return err
	}

	confMap[name] = value
	data, err = yaml.Marshal(confMap)

	if err != nil {
		return err
	}

	return os.WriteFile(c.ConfigFilePath, data, 0644)
}

func CreateConfigFileIfNotExists(config *Config) error {
	if com.IsFile(config.ConfigFilePath) {
		return nil
	}

	file, err := os.Create(config.ConfigFilePath)

	if err != nil {
		return err
	}

	defer file.Close()

	return nil
}

func setDynamicParams(c *Config) {
	c.Port = viper.GetInt(PortOpt)
	c.Token = viper.GetString(TokenOpt)
	c.CaServer = viper.GetString(CaServerOpt)
	c.VarDir = viper.GetString(VarDirOpt)
	c.CertBotEnabled = viper.GetBool(CertBotEnabledOpt)
	c.CertBotBin = viper.GetString(CertBotBinOpt)
	c.CertBotWokrDir = viper.GetString(CertBotWorkDirOpt)
	c.NginxAcmeCommonDir = viper.GetString(NginxAcmeCommonDirOpt)
}
