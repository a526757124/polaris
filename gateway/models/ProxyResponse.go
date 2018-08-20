package models

import (
	"time"
)
type ProxyResponse struct {
	RetCode          int
	RetMsg           string
	LoadConfigTime 		 time.Time
	IntervalTime     int64
	ContentType      string
	Message          interface{}
}
