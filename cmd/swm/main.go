package main

import (
	"flag"
	"github.com/BurntSushi/xgbutil"
	"log"
)

func main() {

	replace := flag.Bool("replace", false, "whether swm should replace current wm")
	flag.Parse()

	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	defer X.Conn().Close()

	if err := takeWmOwnership(X, *replace); err != nil {
		log.Fatalf("Cannot take wm ownership: %s", err)
	}
}
