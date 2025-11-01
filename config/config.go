package config

import (
	"fmt"
	"os"

	"github.com/jinzhu/configor"
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
	FontPath      string
	StartImagePath string
	StartAudioPath string
	GoodImagePath string
}{}

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