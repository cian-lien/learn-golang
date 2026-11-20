package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func serveOne(listener net.Listener) error {
	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(conn, "收到", strings.TrimSpace(line))
	return err
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	done := make(chan error, 1)
	go func() { done <- serveOne(listener) }()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	if _, err := fmt.Fprintln(conn, "你好"); err != nil {
		panic(err)
	}
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		panic(err)
	}
	fmt.Print(reply)
	if err := <-done; err != nil {
		panic(err)
	}
}
