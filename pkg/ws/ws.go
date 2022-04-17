package ws

import (
	"net/http"
	"sync"

	"github.com/nektro/mantle/pkg/db"

	"github.com/gorilla/websocket"
	"github.com/nektro/go.etc/dbt"
	"github.com/nektro/go.etc/store"
	"github.com/valyala/fastjson"
)

const (
	keyOnline = "online_users"
)

var (
	reqUpgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
)

var (
	// UserCache is the list of users currently with ws connections to this instance
	UserCache = map[dbt.UUID]*User{}
)

// Connect takes a db.User and upgrades it to a ws.User
func Connect(user *db.User, w http.ResponseWriter, r *http.Request) (*User, error) {
	conn, err := reqUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	u := &User{
		conn,
		user,
		sync.Mutex{},
	}
	UserCache[u.User.UUID] = u

	if !u.IsConnected() {
		store.This.ListAdd(keyOnline, u.User.UUID.String())
		db.Props.Increment("count_users_online")
		BroadcastMessage(map[string]interface{}{
			"type": "user-connect",
			"user": u.User.UUID,
		})
	}
	return u, nil
}

// Close disconnect all remaining users
func Close() {
	for _, item := range UserCache {
		item.Disconnect()
	}
}

// BroadcastMessage sends message to all users
func BroadcastMessage(message map[string]interface{}) {
	for _, item := range UserCache {
		item.SendWsMessage(message)
	}
}

func BroadcastMessageRaw(message *fastjson.Value) {
	for _, item := range UserCache {
		item.SendWsMessageRaw(message.MarshalTo([]byte{}))
	}
}

// AllOnlineIDs returns ULID of every online user
func AllOnlineIDs() []string {
	return store.This.ListGet(keyOnline)
}

// OnlineUserCount is the total number of active users
func OnlineUserCount() int64 {
	return int64(store.This.ListLen(keyOnline))
}
