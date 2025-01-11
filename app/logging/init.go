package logging

import (
	"Planning_poker/app"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/MatusOllah/slogcolor"
	slogelastic "github.com/nicus101/slog-elastic"
	slogmulti "github.com/samber/slog-multi"
)

func InitLogging(env *app.Env) error {
	if env.Flags.InsecureTls {
		disableTLSVerification()
	}

	handlers := []slog.Handler{
		slogcolor.NewHandler(os.Stderr, slogcolor.DefaultOptions),
	}

	if env.EsLog.Enabled {
		esHandler, err := initEsLogs(env)
		if err != nil {
			return err
		}
		handlers = append(handlers, esHandler)
	}

	slog.SetDefault(slog.New(slogmulti.Fanout(handlers...)))

	return nil
}

func initEsLogs(env *app.Env) (slog.Handler, error) {
	envEsLog := env.EsLog

	if envEsLog.Address == "" || envEsLog.Index == "" {
		return nil, fmt.Errorf("addresses or index not set")
	}

	slogEsCfg := slogelastic.Config{
		Address: envEsLog.Address,
		User:    envEsLog.Username,
		Pass:    envEsLog.Password,
		Index:   envEsLog.Index,
	}

	if err := slogEsCfg.ConnectEsLog(); err != nil {
		return nil, fmt.Errorf("cannot connect elastic: %w", err)
	}

	return slogEsCfg.NewElasticHandler(), nil
}

func disableTLSVerification() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}
