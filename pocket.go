package main

import (
	"net/http"
	"time"

	"github.com/chzyer/flagx"

	"gopkg.in/logex.v1"
)

type Config struct {
	Mongo  string `flag:";def=:27171;usage=connect to mongo"`
	Db     string `flag:"db;def=pocket"`
	Listen string `flag:";def=:8011;usage=listen"`

	Key string
	Crt string
}

var (
	HttpsEnable = true
)

func main() {
	cfg := new(Config)
	flagx.Parse(cfg)

	InitMongo(cfg.Mongo, cfg.Db)
	mux := http.NewServeMux()
	Handler(mux)
	if cfg.Key != "" {
		done := make(chan bool)
		go func() {
			err := http.ListenAndServeTLS(":443", cfg.Crt, cfg.Key, mux)
			if err != nil {
				HttpsEnable = false
				logex.Error(err)
			}
			done <- err == nil
		}()
		select {
		case <-time.After(time.Second):
		case d := <-done:
			if !d {
				break
			}
			mux := http.NewServeMux()
			RedirectHandler(mux)
			if err := http.ListenAndServe(cfg.Listen, mux); err != nil {
				logex.Error(err)
			}
		}
	}
	if err := http.ListenAndServe(cfg.Listen, mux); err != nil {
		logex.Error(err)
	}
}
