package app

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

var env Env

type Env struct {
	App string `env:"APPLICATION_NAME" env-required:"true"`
	Env string `env:"ENV" env-required:"true"`
	Url string `env:"URL" env-required:"true"`

	JiraUrl      string `env:"JIRA_BASE_URL" env-required:"true"`
	JiraAPIToken string `env:"JIRA_API_TOKEN" env-required:"true"`

	EsLog struct {
		Enabled  bool   `env:"ES_LOG_ENABLED" env-default:"false"`
		Address  string `env:"ES_LOG_ADDRESS"`
		Username string `env:"ES_LOG_USER"`
		Password string `env:"ES_LOG_PASS"`
		Index    string `env:"ES_LOG_INDEX"`
	}

	Flags struct {
		InsecureTls bool `env:"FLAG_INSECURE_TLS" env-default:"false"`
	}
}

func GetEnv() (*Env, error) {
	err := cleanenv.ReadConfig(".env", &env)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	err = cleanenv.ReadEnv(&env)
	if err != nil {
		return nil, err
	}

	return &env, nil
}

func GetVersion() string {
	file, err := os.Open(".version")
	if err != nil {
		return ""
	}

	defer file.Close()

	versionFromFile, versionFromFileErr := ioutil.ReadAll(file)
	if versionFromFileErr != nil {
		return ""
	}

	return strings.TrimSuffix(string(versionFromFile), "\n")
}
