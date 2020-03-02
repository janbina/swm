package main

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/cmd"
	"log"
	"net"
	"os"
	"strings"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("No arguments provided")
	}

	conn, err := net.Dial("unix", socketFilePath())
	if err != nil {
		log.Fatalf("Cannot connect to socket: %s", err)
	}
	defer conn.Close()

	command := strings.Join(os.Args[1:], " ")
	if _, err = fmt.Fprintf(conn, "%s%c", command, 0); err != nil {
		log.Fatalf("Error writing command: %s", err)
	}

	reader := bufio.NewReader(conn)
	reply, err := reader.ReadString(0)
	if err != nil {
		log.Fatalf("Cannot read response: %s", err)
	}
	reply = reply[:len(reply)-1]

	if len(reply) > 0 {
		fmt.Println(reply)
	}
}

func socketFilePath() string {
	// TODO: Use swm executable from path
	c := cmd.New("./swm", "--show-socket")
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(c.BufStdout.String())
}
