package main

import (
	"github.com/gorilla/websocket"
)

// clientはチャットを行う1人のユーザーを表す
type client struct {
	socket *websocket.Conn
	send   chan []byte // send: メッセージが送られるチャネル
	room   *room       // room: クライアントが参加しているチャットルーム
}

// クライアントがWebSocketからReadMessageを使ってデータを読み込む
// 受け取ったデータはroomのforwardチャネルに送信される
func (c *client) read() {
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

// 継続的にsendチャネルからメッセージを受け取って書き出しを行う
func (c *client) write() {
	// for value range slice でsliceを1つずつ取り出して処理を行う
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
