package endpoint

import (
	"encoding/json"
	"fmt"
	"log"
)

type message struct {
	data []byte
	room string
	conn *connection
}

type msgdetail struct {
	MsgType int    `json:"msgType"`
	Msg     string `json:"msg"`
}

type subscription struct {
	conn *connection
	room string
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type room struct {
	// Registered connections.
	rooms map[string]map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan message

	// Register requests from the connections.
	register chan subscription

	// Unregister requests from connections.
	unregister chan subscription
}

const (
	LOGIN = iota
	CHAT
)

// RoomSet 房間所需資訊
var RoomSet = room{
	broadcast:  make(chan message, 100),
	register:   make(chan subscription, 100),
	unregister: make(chan subscription, 100),
	rooms:      make(map[string]map[*connection]bool),
}

// OnlineUser 紀錄名稱
var OnlineUser = make(map[*connection]string)

func (r *room) Run() {
	for {
		select {
		case s := <-r.register:
			connections := r.rooms[s.room]
			if connections == nil {
				connections = make(map[*connection]bool)
				r.rooms[s.room] = connections
			}
			r.rooms[s.room][s.conn] = true
			OnlineUser[s.conn] = ""
		case s := <-r.unregister:
			connections := r.rooms[s.room]
			if connections != nil {
				if _, ok := connections[s.conn]; ok {
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
			if msgdet.MsgType == LOGIN {
				OnlineUser[m.conn] = msgdet.Msg
				msgstr := fmt.Sprint(msgdet.Msg, " 已加入")
				sendData = []byte(msgstr)
			} else {
				msgAddName := fmt.Sprint(OnlineUser[m.conn], " : ", msgdet.Msg)
				sendData = []byte(msgAddName)
			}

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
