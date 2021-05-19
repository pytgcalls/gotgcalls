package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

var port, logMode int
var client *GoSocketClient

func main() {
	flag.Parse()
	envVars := flag.Args()
	port, _ = strconv.Atoi(strings.Split(envVars[0],"=")[1])
	logMode, _ = strconv.Atoi(strings.Split(envVars[1],"=")[1])
	client = GoSocket(onMessage)
	client.Run()
}
func onMessage(data map[string]interface{}) {
	if logMode > 0 {
		fmt.Println("REQUEST: ", data)
	}
	if data["action"] == "join_call" {
		connections := RTCPeerConnection(
			normalizeInt(data["chat_id"]),
			normalizeString(data["invite_hash"]),
		)
		stream := Stream(
			data["file_path"].(string),
			16,
			normalizeInt(data["bitrate"]),
			1,
			logMode,
			normalizeInt(data["buffer_long"]),
		)
		if stream != nil{
			result := connections.joinCall()
			if result{
				stream.start(connections.Track())
			}
		}
	}
}