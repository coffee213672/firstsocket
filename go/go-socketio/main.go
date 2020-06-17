package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	socketio "github.com/googollee/go-socket.io"
)

type loginInfo struct {
	Name string `json:"name"`
}

type errorInfo struct {
	ErrorMsg string `json:"errorMeg"`
}

var onlineUser = make(map[string]string)

func main() {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.OnConnect("/", func(s socketio.Conn) error {
		return nil
	})

	server.OnEvent("", "login", func(s socketio.Conn, msg string) {
		var loginUser loginInfo
		if err := json.Unmarshal([]byte(msg), &loginUser); err != nil {
			log.Println(err)
		}

		fmt.Println(loginUser.Name)

		_, ok := onlineUser[loginUser.Name]
		if !ok {
			onlineUser[loginUser.Name] = loginUser.Name
		} else {
			// errorEvent(s, "暱稱重複")
			errMsg := errorInfo{
				ErrorMsg: "暱稱重複",
			}
			jsonData, err := json.Marshal(errMsg)
			if err != nil {
				log.Println(err)
				return
			}
			s.Emit("err", string(jsonData))
		}

		return
	})

	server.OnEvent("", "chat message", func(s socketio.Conn, msg string) {
		s.Emit("chat message", msg)
		return
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})

	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	log.Println("Serving at localhost:8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func errorEvent(s socketio.Conn, msg string) {
	errMsg := errorInfo{
		ErrorMsg: msg,
	}
	jsonData, err := json.Marshal(errMsg)
	if err != nil {
		log.Println(err)
		return
	}
	s.Emit("error", string(jsonData))
	return
}
