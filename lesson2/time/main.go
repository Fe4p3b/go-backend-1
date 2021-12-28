package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go message(conn)
		go handleConn(conn)
	}
}

func handleConn(c net.Conn) {
	defer c.Close()
	for {
		_, err := io.WriteString(c, time.Now().Format("15:04:05\n\r"))
		if err != nil {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func message(c net.Conn) {
	defer c.Close()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		_, err := io.WriteString(c, scanner.Text()+"\n\r")
		if err != nil {
			return
		}
	}
}
