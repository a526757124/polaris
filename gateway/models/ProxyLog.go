package models

import (
	"time"
	"github.com/devfeel/polaris/models"
)

type ProxyLog struct {
	RetCode           int
	RetMsg            string
	RequestUrl        string
	CallInfo 	  	  []*models.TargetApiInfo
	RawResponseFlag	  bool
	LoadConfigTime  time.Time
	IntervalTime      int64
	ContentType       string
	Message           interface{}
}