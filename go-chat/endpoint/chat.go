package endpoint

import (
	"encoding/json"
	"fmt"
	"go-chat/cache"
	"log"

	"github.com/gofrs/uuid"
)

type message struct {
	data []byte
	room string
	conn *connection
}

// UserLogin ID 尚未使用
type UserLogin struct {
	IsSuccess bool
	ID        string
}

type msgdetail struct {
	MsgType int    `json:"msgType"`
	Msg     string `json:"msg"`
}

type subscription struct {
	conn *connection
	room string
}

type messageResp struct {
	MsgType        int                `json:"msgType"`
	Msg            string             `json:"msg"`
	OnlineUserList *map[string]string `json:"userlist"`
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type room struct {
	// Registered connections.
	rooms map[string]map[*connection]UserLogin

	// Inbound messages from the connections.
	broadcast chan message

	// Register requests from the connections.
	register chan subscription

	// Unregister requests from connections.
	unregister chan subscription
}

// 聊天室常數
const (
	LOGIN = iota
	CHAT
	NAMEREPEAT
	ERROR
	SOMEONEJOIN
)

// ONLINEUSERLIST redis key
const ONLINEUSERLIST = "OnlineUser"

// RoomSet 房間所需資訊
var RoomSet = room{
	broadcast:  make(chan message, 100),
	register:   make(chan subscription, 100),
	unregister: make(chan subscription, 100),
	rooms:      make(map[string]map[*connection]UserLogin),
}

// OnlineUser 紀錄名稱
var OnlineUser = make(map[*connection]string)

func (r *room) Run() {
	for {
		select {
		case s := <-r.register:
			connections := r.rooms[s.room]
			if connections == nil {
				connections = make(map[*connection]UserLogin)
				r.rooms[s.room] = connections
			}
			uuid, _ := uuid.NewV4()
			userlogin := UserLogin{
				IsSuccess: true,
				ID:        uuid.String(),
			}
			r.rooms[s.room][s.conn] = userlogin
			OnlineUser[s.conn] = ""

		case s := <-r.unregister:
			connections := r.rooms[s.room]
			if connections != nil {
				if _, ok := connections[s.conn]; ok {
					cache.DelOnlineUser(ONLINEUSERLIST, OnlineUser[s.conn])
					OnlineUser[s.conn] = ""
					delete(connections, s.conn)
					close(s.conn.send)
					if len(connections) == 0 {
						delete(r.rooms, s.room)
					}
				}
			}

		case m := <-r.broadcast:
			connections := r.rooms[m.room]
			var msgdet msgdetail
			if err := json.Unmarshal(m.data, &msgdet); err != nil {
				log.Println(err)
				return
			}
			var sendData []byte
			resp := getResponseData(msgdet, m)
			respData, _ := json.Marshal(resp)
			sendData = respData

			if resp.MsgType == NAMEREPEAT || resp.MsgType == ERROR {
				if _, ok := connections[m.conn]; ok {
					select {
					case m.conn.send <- sendData:
					default:
						close(m.conn.send)
						delete(connections, m.conn)
					}
				}
			} else {
				for c := range connections {
					select {
					case c.send <- sendData:
					default:
						close(c.send)
						delete(connections, c)
						if len(connections) == 0 {
							delete(r.rooms, m.room)
						}
					}
				}
			}

		}
	}
}

func getResponseData(msgdet msgdetail, m message) (resp messageResp) {
	if msgdet.MsgType == LOGIN {
		checkName := cache.GetOnlineUser(ONLINEUSERLIST, msgdet.Msg)
		if checkName != "" {
			resp.MsgType = NAMEREPEAT
			resp.Msg = "暱稱重複"
		} else {
			isSuccess := cache.SetOnlineUser(ONLINEUSERLIST, msgdet.Msg, msgdet.Msg)
			if !isSuccess {
				resp.MsgType = ERROR
				resp.Msg = "Server Error"
			} else {
				resp.MsgType = SOMEONEJOIN
				resp.Msg = fmt.Sprint(msgdet.Msg, " 已加入")
				OnlineUser[m.conn] = msgdet.Msg
				onlineUserList := cache.GetAllOnlineUser(ONLINEUSERLIST)
				resp.OnlineUserList = &onlineUserList
			}
			cache.GetAllOnlineUser(ONLINEUSERLIST)
		}
	} else {
		resp.MsgType = CHAT
		resp.Msg = fmt.Sprint(OnlineUser[m.conn], " : ", msgdet.Msg)
	}
	return
}
