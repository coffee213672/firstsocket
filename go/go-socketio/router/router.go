package router

import (
	"encoding/json"
	"fmt"
	"go-socketio/model"
	"log"
	"net/http"
	"sync"

	socketio "github.com/googollee/go-socket.io"
)

var onlineUserID = make(map[string]string)
var onlineUser = make(map[string]string)
var writeLock sync.Mutex

// InitRouter ...
func InitRouter() {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.OnConnect("/", func(s socketio.Conn) error {
		s.Join("main")
		fmt.Println("connected:", s.ID())
		fmt.Println("connect-namespace", s.Namespace())
		return nil
	})

	server.OnEvent("", "login", func(s socketio.Conn, msg string) {
		var loginUser model.LoginInfo
		if err := json.Unmarshal([]byte(msg), &loginUser); err != nil {
			log.Println(err)
		}

		fmt.Println(loginUser.Name)

		_, ok := onlineUser[loginUser.Name]
		if !ok {
			writeLock.Lock()
			onlineUser[loginUser.Name] = loginUser.Name
			onlineUserID[s.ID()] = loginUser.Name
			writeLock.Unlock()
		} else {
			// errorEvent(s, "暱稱重複")
			errMsg := model.ErrorInfo{
				ErrorMsg: "暱稱重複",
			}
			jsonData, err := json.Marshal(errMsg)
			if err != nil {
				log.Println(err)
				return
			}
			s.Emit("repeat", string(jsonData))
		}
		server.BroadcastToRoom("", "main", "userlist", onlineUserID)
		return
	})

	server.OnEvent("", "chat message", func(s socketio.Conn, msg string) {
		s.SetContext(msg)
		// s.Emit("chat message", msg)
		fmt.Println("talk:", s.ID())
		fmt.Println("talk-namespace", s.Namespace())
		totalMsg := fmt.Sprint(onlineUserID[s.ID()], " : ", msg)
		fmt.Println(totalMsg)
		server.BroadcastToRoom("", "main", "chat message", totalMsg)
		return
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		_, ok := onlineUserID[s.ID()]
		if ok {
			writeLock.Lock()
			delete(onlineUser, onlineUserID[s.ID()])
			delete(onlineUserID, s.ID())
			writeLock.Unlock()
		}
		log.Println("closed", reason)
	})

	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	log.Println("Serving at localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
