package config

import (
	"fmt"
	"os"

	"github.com/jinzhu/configor"
	"gopkg.in/yaml.v3"
)

const (
	ContextUserClaimKey = "userClaim"
	ContextDBKey        = "DB"
)

var Config = struct {
	HttpPort    string
	Environment string
	Database    struct {
		Driver           string
		User             string
		Connection       string
		ConnectionString string
	}
	Service struct {
		Name string
	}
	FontPath       string
	StartAudioPath string
	Paths          struct {
		TempDir       string
		TempImagesDir string
		TempAudioDir  string
		TempVideosDir string
		FinalVideoDir string
		TemplateDir   string
		Templates     struct {
			Vertical      string
			BackgroundImg string
			Title         string
			StartImg      string
			GoodImg       string
			StartComment string
		}
	}
}{}

// CliConfig holds the configuration for the CLI application, read from config.yaml
type CliConfig struct {
	Video struct {
		Type string `yaml:"type"`
		Date string `yaml:"date"`
	} `yaml:"video"`
}

// LoadCliConfig reads the config.yaml file and returns the configuration.
func LoadCliConfig(path string) (*CliConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// If the file doesn't exist, return a default config.
		if os.IsNotExist(err) {
			return &CliConfig{
				Video: struct {
					Type string `yaml:"type"`
					Date string `yaml:"date"`
				}{
					Type: "W",
					Date: "today",
				},
			}, nil
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg CliConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

func InitConfig(cfg string) {
	configor.Load(&Config, cfg)
}

// 서비스 무관하게 공통으로 사용하는 부분
func ConfigureEnvironment(path string, env ...string) {
	configor.Load(&Config, path+"config/config.json") //배포 환경에 따른 설정 파일(json)을 로딩한다.
	properties := make(map[string]string)

	for _, key := range env {
		arg := os.Getenv(key)
		if len(arg) == 0 {
			panic(fmt.Errorf("No %s system env variable\n", key))
		}
		properties[key] = arg
	}

	afterPropertiesSet(properties)
}

// 서비스별 처리 로직이 달라지는 부분.
func afterPropertiesSet(properties map[string]string) {
	if properties["STUDY_GENIE_DB_PASSWORD"] != "" {
		Config.Database.ConnectionString = fmt.Sprintf("%s:%s%s", Config.Database.User, properties["STUDY_GENIE_DB_PASSWORD"], Config.Database.Connection)
	} else {
		Config.Database.ConnectionString = Config.Database.Connection
	}
}
