package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/davecgh/go-spew/spew"

	"github.com/spf13/viper"
)

type config struct {
	AppEnv   string
	Database struct {
		User                 string
		Password             string
		Addr                 string
		Net                  string
		DBName               string
		Port                 string
		AllowNativePasswords bool
		Params               struct {
			ParseTime string
			Charset   string
			Loc       string
			TLS       string
		}
	}
	JwtTokenSecret     string
	RefreshTokenSecret string
	AppName            string
	Server             struct {
		Address string
	}
	AWS struct {
		Region          string
		S3Bucket        string
		AccessKeyID     string
		SecretAccessKey string
	}
	RapidAPI struct {
		APIKey        string
		BaseURL       string
		MonthlyQuota  int
		TimeoutSeconds int
	}
	Email struct {
		SMTPHost     string
		SMTPPort     int
		Username     string
		Password     string
		FromAddress  string
		AdminEmail   string
	}
	Cron struct {
		ProfileFetcherSchedule string
		QuotaResetSchedule     string
		BatchSize              int
	}
}

// C is config variable
var C config

// Application Environment name
const (
	Development = "development"
	Test        = "test"
	E2E         = "e2e"
	Staging     = "staging"
	Production  = "production"
)

// ReadConfigOption is a config option
type ReadConfigOption struct {
	AppEnv string
}

// ReadConfig configures config file
func ReadConfig(option ReadConfigOption) {
	Config := &C

	e := appEnv(option)

	if e == Test {
		setTest()
	} else if e == E2E {
		setE2E()
	} else if e == Staging {
		setStaging()
	} else if e == Development {
		setDev()
	} else {
		setProd()
	}

	viper.SetConfigType("yml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		log.Fatalln(err)
	}

	if err := viper.Unmarshal(&Config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	spew.Dump(C)
}

func appEnv(option ReadConfigOption) string {
	if option.AppEnv != "" {
		return option.AppEnv
	}
	if os.Getenv("APP_ENV") != "" {
		return os.Getenv("APP_ENV")
	}

	return Development
}

func rootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func setDev() {
	viper.AddConfigPath(filepath.Join(rootDir(), "config"))
	viper.SetConfigName("config")
}

func setTest() {
	viper.AddConfigPath(filepath.Join(rootDir(), "config"))
	viper.SetConfigName("config.test")
}

func setE2E() {
	viper.AddConfigPath(filepath.Join(rootDir(), "config"))
	viper.SetConfigName("config.e2e")
}

func setStaging() {
	viper.AddConfigPath(filepath.Join(rootDir(), "config"))
	viper.SetConfigName("config.staging")
}

func setProd() {
	viper.AddConfigPath(filepath.Join(rootDir(), "config"))
	viper.SetConfigName("config.production")
}
