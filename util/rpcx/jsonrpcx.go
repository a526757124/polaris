package rpcx

import (
	"net/rpc/jsonrpc"
	"encoding/json"
	"time"
)

// CallJsonRPC common call remote json rpc api
func CallJsonRPC(serverUrl string, callName string, body []byte) ([]byte, int64, error){
	startTime := time.Now()
	var intervalTime int64
	client, err := jsonrpc.Dial("tcp", serverUrl)
	if err != nil {
		intervalTime = int64(time.Now().Sub(startTime) / time.Millisecond)
		return nil, intervalTime, err
	}
	args := map[string]interface{}{}
	err = json.Unmarshal(body, &args)
	if err != nil{
		intervalTime = int64(time.Now().Sub(startTime) / time.Millisecond)
		return nil, intervalTime, err
	}
	var reply interface{}
	err = client.Call(callName, args, &reply)
	if err != nil {
		intervalTime = int64(time.Now().Sub(startTime) / time.Millisecond)
		return nil, intervalTime, err
	}
	br, err := json.Marshal(reply)
	if err != nil {
		intervalTime = int64(time.Now().Sub(startTime) / time.Millisecond)
		return nil, intervalTime, err
	}
	intervalTime = int64(time.Now().Sub(startTime) / time.Millisecond)
	return br, intervalTime, nil
}
