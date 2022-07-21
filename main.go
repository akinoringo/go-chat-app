package main

import (
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// Httpリクエストを処理する
// 1. 入力元のファイルの読み込み
// 2. テンプレートをコンパイルして実行
// 3. 結果をhttp.ResponseWriterオブジェクトに出力
// 既存のStructや型に対して、ServeHTTPメソッドを用意することでhttp.Handleに登録できるようになる
// 参考：https://qiita.com/taizo/items/bf1ec35a65ad5f608d45
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// tにはアドレスが入るため、実体にアクセスするためには(*t).once.Do(func)のように書く必要がある。
	// しかし、構造体の場合はtのみでも自動でtの実体を表すため以下のように書ける
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
		t.templ.Execute(w, nil)
	})
}

func main() {
	r := newRoom()
	// ルート
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	// チャットルーム開始
	// goroutineとしてチャット関連の処理はバックグラウンドで実行
	go r.run()

	// WEBサーバーの開始
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
