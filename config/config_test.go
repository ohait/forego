package config_test

import (
	"testing"

	"github.com/ohait/forego/config"
	"github.com/ohait/forego/test"
)

func TestConfig(t *testing.T) {
	c := test.Context(t)

	cfg := config.Must(c, struct {
		Listen string `config:"listen,default=:8080"`
		Local  bool   `config:"local,default=true"`
	}{}, func(key string) string {
		switch key {
		default:
			return ""
		}
	})
	test.EqualsJSON(t, ":8080", cfg.Listen)
}
