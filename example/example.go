package example

import (
	"sync"

	"github.com/ohait/forego/ctx"
)

type Store struct {
	m    sync.Mutex
	Data map[string]any
}

type Get struct {
	any `doc:"yooo hooo"`

	XFF string `api:"header,X-Forwarded-For"`
	UID string `api:"auth,required"`

	Store *Store

	Key   string `api:"in,out" json:"key"`
	Value any    `api:"out" json:"value"`
}

func (this *Get) Do(c ctx.C) error {
	this.Value = this.Store.Data[this.Key]
	return nil
}

type Set struct {
	Store *Store

	Key   string `api:"in,out" json:"key"`
	Value any    `api:"in" json:"value"`
	Prev  any    `api:"out" json:"prev"`
}

func (this *Set) Do(c ctx.C) error {
	this.Prev = this.Store.Data[this.Key]
	this.Store.Data[this.Key] = this.Value
	return nil
}
