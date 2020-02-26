package main

import (
	"github.com/BurntSushi/xgbutil"
	"log"
)

func main() {

	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	defer X.Conn().Close()

	if err := takeWmOwnership(X, true); err != nil {
		log.Fatalf("Cannot take wm ownership: %s", err)
	}
}
