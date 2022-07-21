package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {
	// chan はチャネルの宣言 その次に受け取るデータの型を記述する
	forward chan []byte  // 他のクライアントに転送するためのメッセージを保持するチャネル
	join    chan *client // チャットルームに参加しようとしているクライアントのためのチャネル
	leave   chan *client // チャットルームから退室しようとしているクライアントのためのチャネル
	// map[keyType]valueTypeで {key(型はkeyTypeに合うもの): value(型はvalueTypeに合うもの)} というPythonで言う辞書型のものが定義されている。
	clients map[*client]bool // 在室しているすべてのクライアントが保持される
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

// for, select, caseを使うことでチャネルに渡ってきたものを判定して実行コマンドを振り分けできる
// defaultは、caseに当てはまらない場合の挙動
func (r *room) run() {
	// goroutineとして実行する場合、無限ループのforは問題にならない
	for {
		select {
		case client := <-r.join:
			// 参加
			r.clients[client] = true
		case client := <-r.leave:
			// 退室
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			// すべてのクライアントにメッセージを転送
			// チャネルの値を「for value := range chan」で取り出すことができる
			for client := range r.clients {
				select {
				case client.send <- msg:
					// メッセージを送信
				default:
					// 送信に失敗
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
