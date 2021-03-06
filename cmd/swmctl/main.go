package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgbutil"
	"github.com/janbina/swm/internal/buildconfig"
	"github.com/janbina/swm/internal/communication"
)

func main() {

	version := flag.Bool("v", false, "print swmctl version")
	flag.Parse()

	if *version {
		fmt.Println(buildconfig.Version)
		os.Exit(0)
	}

	xgb.Logger = log.New(ioutil.Discard, "", 0)
	x, err := xgbutil.NewConn()
	if err != nil {
		fmt.Println("Cannot initialize X connection")
		os.Exit(1)
	}
	defer x.Conn().Close()

	socket := communication.GetSocketFilePath(x.Conn())

	conn, err := net.Dial("unix", socket)
	if err != nil {
		fmt.Printf("Cannot connect to swm. Is swm running on display %d?", x.Conn().DisplayNumber)
		os.Exit(1)
	}
	defer func() { _ = conn.Close() }()

	args := make([]string, len(os.Args)-1)
	for i, a := range os.Args[1:] {
		args[i] = fmt.Sprintf("\"%s\"", a)
	}

	command := strings.Join(args, " ")
	if _, err = fmt.Fprintf(conn, "%s%c", command, 0); err != nil {
		fmt.Printf("Cannot send command to swm")
		os.Exit(1)
	}

	reader := bufio.NewReader(conn)
	reply, err := reader.ReadString(0)
	if err != nil {
		fmt.Printf("Cannot read swm's reply")
		os.Exit(1)
	}
	reply = reply[:len(reply)-1]

	if len(reply) > 0 {
		fmt.Println(reply)
	}
}
