package config

import (
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

var env Env

type Env struct {
	App string `env:"APPLICATION_NAME" env-required:"true"`
	Env string `env:"ENV" env-required:"true"`
	Url string `env:"URL" env-required:"true"`
}

func Get() (*Env, error) {
	err := cleanenv.ReadConfig(".env", &env)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if err != nil {
		err = cleanenv.ReadEnv(&env)
	}

	return &env, nil
}
