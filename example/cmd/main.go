package main

import (
	"os"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/example"
	"github.com/ohait/forego/http"
	"github.com/ohait/forego/shutdown"
)

func main() {
	c, cf := ctx.Background()
	log.Warnf(c, "init")
	defer log.Warnf(c, "exit")

	s := http.NewServer(c)
	store := example.Store{
		Data: map[string]any{
			"true": true,
		},
	}
	s.RegisterAPI(c, "/api/v1/get", &example.Get{Store: &store})
	s.RegisterAPI(c, "/api/v1/set", &example.Set{Store: &store})

	addr, err := s.Listen(c, "127.0.0.1:0")
	if err != nil {
		log.Errorf(c, "err")
		os.Exit(-1)
	}

	log.Infof(c, "listening %v", addr.String())

	shutdown.WaitForSignal(c, cf)
}
