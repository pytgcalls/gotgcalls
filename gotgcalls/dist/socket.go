package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrr/fastws"
	"github.com/valyala/fasthttp"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

type GoSocketClient struct{
	handler       func(json map[string]interface{})
	channelResult map[string]chan string
	channelSend   map[string]chan string
}

func GoSocket(handler func(json map[string]interface{})) *GoSocketClient {
	client := &GoSocketClient{
		handler: handler,
		channelResult: make(map[string]chan string),
		channelSend: make(map[string]chan string),
	}
	//goland:noinspection GoUnreachableCode
	return client
}
func (g GoSocketClient) Run(){
	log.Fatal(fasthttp.ListenAndServe(fmt.Sprintf(":%d", port), g.webHandler))
}

func (g GoSocketClient) webHandler(ctx *fasthttp.RequestCtx)  {
	switch string(ctx.Path()) {
	case "/go_socket":
		upgr := fastws.Upgrader{
			Handler:  g.wsHandler,
			Compress: true,
		}
		upgr.Upgrade(ctx)
	default:
		if ctx.IsPost(){
			var bodyPost map[string]interface{}
			jsonFix := strings.ReplaceAll(string(ctx.PostBody()), "\r\n","")
			jsonFix = strings.ReplaceAll(jsonFix, " ","")
			jsonFix = strings.ReplaceAll(jsonFix, ",}","}")
			err := json.Unmarshal([]byte(jsonFix), &bodyPost)
			if err == nil{
				sessionID := g.generateSessionID(16)
				messageRequest, _ := json.Marshal(map[string]interface{}{
					"path": string(ctx.Path()),
					"post": bodyPost,
					"session_id": sessionID,
				})
				g.channelSend[sessionID] = make(chan string)
				g.channelResult[sessionID] = make(chan string)
				g.channelSend[sessionID] <- string(messageRequest)
				messageResult := <-g.channelResult[sessionID]
				delete(g.channelResult, sessionID)
				var resultUM map[string]interface{}
				_ = json.Unmarshal([]byte(messageResult), &resultUM)
				statusCode := normalizeInt(resultUM["status"].(map[string]interface{})["code"])
				if statusCode == 404{
					ctx.Error("Error: 404 Not Found", fasthttp.StatusNotFound)
				}else if statusCode == 500{
					ctx.Error("Error: 500 Internal Server Error", fasthttp.StatusInternalServerError)
				}else{
					ctx.SetContentType("application/json")
					ctx.SetStatusCode(fasthttp.StatusOK)
					bodyResult, _ := json.Marshal(resultUM["result"])
					ctx.SetBody(bodyResult)
					ctx.SetConnectionClose()
				}
			}else{
				ctx.Error("Error: 400 Bad Request", fasthttp.StatusBadRequest)
			}
		}else{
			ctx.Error("Error: 405 Method Not Allowed", fasthttp.StatusMethodNotAllowed)
		}
	}
}

func (g GoSocketClient) generateSessionID(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (g GoSocketClient) wsHandler(conn *fastws.Conn) {
	var err error
	_, err = conn.WriteString("CONNECTED")
	var msg []byte
	go func() {
		for {
			randomNum := rand.Intn(5 - 3 + 1) + 3
			_, err = conn.Write([]byte("PING"))
			time.Sleep(time.Second * time.Duration(randomNum))
		}
	}()
	go func() {
		for {
			for key, element2 := range g.channelSend {
				request := <-element2
				delete(g.channelSend, key)
				_, err = conn.Write([]byte(request))
			}
		}
	}()
	for {
		_, msg, err = conn.ReadMessage(msg[:0])
		if err != nil {
			if err != fastws.EOF {
				log.Fatal(fmt.Fprintf(os.Stderr, "error reading message: %s\n", err))
			}
		}
		messageText := string(msg)
		if messageText != "PONG" && messageText != "RECEIVED" {
			var data map[string]interface{}
			err = json.Unmarshal([]byte(messageText), &data)
			if err == nil{
				if data["status"] != nil{
					sessionID := data["session_id"].(string)
					g.channelResult[sessionID] <- messageText
				}else{
					go func() {
						g.handler(data)
					}()
				}
			}
		}

		_, err = conn.Write([]byte("RECEIVED"))
		if err != nil {
			log.Fatal(fmt.Fprintf(os.Stderr, "error writing message: %s\n", err))
		}
	}
}
func (g *GoSocketClient) requestData(bodyPost map[string]interface{}, path string) *string{
	sessionID := g.generateSessionID(16)
	messageRequest, _ := json.Marshal(map[string]interface{}{
		"path": path,
		"post": bodyPost,
		"session_id": sessionID,
	})
	g.channelSend[sessionID] = make(chan string)
	g.channelResult[sessionID] = make(chan string)
	g.channelSend[sessionID] <- string(messageRequest)
	messageResult := <-g.channelResult[sessionID]
	delete(g.channelResult, sessionID)
	var resultUM map[string]interface{}
	_ = json.Unmarshal([]byte(messageResult), &resultUM)
	statusCode := normalizeInt(resultUM["status"].(map[string]interface{})["code"])
	if statusCode != 200{
		return nil
	}else{
		return &messageResult
	}
}
