package main

import (
	"net/http"

	"github.com/chzyer/flagx"

	"gopkg.in/logex.v1"
)

type Config struct {
	Listen string `flag:"[0]"`
	Mongo  string `flag:";def=:27171;usage=connect to mongo"`
	Db     string `flag:"db;def=pocket"`
	Key    string
	Crt    string
}

func main() {
	cfg := new(Config)
	flagx.Parse(cfg)

	InitMongo(cfg.Mongo, cfg.Db)
	mux := http.NewServeMux()
	Handler(mux)
	if cfg.Key != "" {
		go func() {
			err := http.ListenAndServeTLS(":443", cfg.Crt, cfg.Key, mux)
			if err != nil {
				logex.Error(err)
			}
		}()
	}
	if err := http.ListenAndServe(":8011", mux); err != nil {
		logex.Error(err)
	}
}
