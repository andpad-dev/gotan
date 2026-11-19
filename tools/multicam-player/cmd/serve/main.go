// Command serve は、確認用の静的ファイルサーバー.
//
// net/http の FileServer は Range リクエストに応えるので、動画のシークを試せる.
// (python3 -m http.server は Range に対応せず、シークすると先頭に戻る.)
//
// usage: serve <dir> <port>
//
// 例:
//
//	go run ./cmd/serve . 8765
package main

import (
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("usage: serve <dir> <port>")
	}
	addr := net.JoinHostPort("127.0.0.1", os.Args[2])
	log.Printf("http://%s/", addr)
	log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir(os.Args[1]))))
}
