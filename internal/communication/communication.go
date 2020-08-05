package communication

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path"

	"github.com/BurntSushi/xgb"
	"github.com/janbina/swm/internal/log"
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
	defer func() { _ = listener.Close() }()

	for {
		conn, err := listener.Accept()
		if err == nil {
			go handleClient(conn)
		}
	}
}

func handleClient(conn net.Conn) {
	for {
		msg, err := bufio.NewReader(conn).ReadString(0)
		if err != nil {
			break
		}
		msg = msg[:len(msg)-1]

		out := processCommand(msg)

		if _, err := fmt.Fprintf(conn, "%s%c", out, 0); err != nil {
			log.Infof("Error sending response to swmctl: %s", err)
		}
	}
	_ = conn.Close()
}
