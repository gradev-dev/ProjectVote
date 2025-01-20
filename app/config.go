package app

import (
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"io/ioutil"
	"os"
	"strings"
)

var env Env

type Env struct {
	App          string `env:"APPLICATION_NAME" env-required:"true"`
	Env          string `env:"ENV" env-required:"true"`
	Url          string `env:"URL" env-required:"true"`
	JiraUrl      string `env:"JIRA_BASE_URL" env-required:"true"`
	JiraAPIToken string `env:"JIRA_API_TOKEN" env-required:"true"`
}

func GetEnv() (*Env, error) {
	err := cleanenv.ReadConfig(".env", &env)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if err != nil {
		err = cleanenv.ReadEnv(&env)
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
