package communication

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/xgb"
	"log"
	"net"
	"os"
	"path"
)

func GetSocketFilePath(x *xgb.Conn) string {
	name := fmt.Sprintf(":%d.%d", x.DisplayNumber, x.DefaultScreen)

	var runtimeDir string
	xdgRuntime := os.Getenv("XDG_RUNTIME_DIR")
	if len(xdgRuntime) > 0 {
		runtimeDir = path.Join(xdgRuntime, "swm")
	} else {
		runtimeDir = path.Join(os.TempDir(), "swm")
	}

	if err := os.MkdirAll(runtimeDir, 0777); err != nil {
		log.Fatalf("Cannot create dir: %s", err)
	}

	return path.Join(runtimeDir, name)
}

func Listen(x *xgb.Conn) {
	addr := GetSocketFilePath(x)

	_ = os.Remove(addr)

	listener, err := net.Listen("unix", addr)
	if err != nil {
		log.Fatalf("Cannot start listener")
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err == nil {
			go handleClient(conn)
		}
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	for {
		msg, err := bufio.NewReader(conn).ReadString(0)
		if err != nil {
			return
		}
		msg = msg[:len(msg)-1]

		log.Printf("Got command from swmctl: %s", msg)

		out := processCommand(msg)

		fmt.Fprintf(conn, "%s%c", out, 0)
	}
}
