package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	// LogFlag 控制日志的前缀
	LogFlag = log.LstdFlags | log.Lmicroseconds | log.Lshortfile
	server  = "echod:8080"
	// interval 表示重连或者写数据的间隔时间
	interval = 5 * time.Second
	message  = "ping\n"
)

var (
	errLogger  = log.New(os.Stderr, "ERROR ", LogFlag)
	infoLogger = log.New(os.Stdout, "INFO ", LogFlag)
)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM)

	running := true
	for running {
		infoLogger.Printf("Will connect to server: %s...", server)
		conn, err := net.Dial("tcp", server)
		if err != nil {
			errLogger.Printf("net.Dial() failed, error: %s.", err)
		} else {
			infoLogger.Printf("Connected, RemoteAddr: %s.", conn.RemoteAddr())

			handle(conn, quit)
		}

		select {
		case signal := <-quit:
			infoLogger.Printf("Receive a signal: %d, and I will shutdown gracefully...", signal)
			running = false
		case <-time.Tick(interval):
			infoLogger.Printf("Will dial again...")
		}
	}

	infoLogger.Print("Shutdown gracefully.")
}

func handle(conn io.ReadWriteCloser, quit chan os.Signal) {
	defer func() {
		if err := conn.Close(); err != nil {
			errLogger.Printf("conn.Close() failed, error: %s.", err)
		}
	}()

	buf := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	running := true
	for running {
		n, err := buf.WriteString(message)
		if err != nil {
			errLogger.Printf("buf.WriteString() failed, error: %s.", err)
			running = false
		}

		if err = buf.Flush(); err != nil {
			errLogger.Printf("buf.Flush() failed, error: %s.", err)
			running = false
		}

		infoLogger.Printf("Write a message: %s, length: %d bytes.", strings.TrimSpace(message), n)

		s, err := buf.ReadString('\n')
		if err != nil {
			errLogger.Printf("buf.ReadString() failed, error: %s.", err)
			running = false
		}

		infoLogger.Printf("Read a response: %s.", strings.TrimSpace(s))

		if s != message {
			errLogger.Printf("Response is wrong, want: %s, got: %s.",
				strings.TrimSpace(message), strings.TrimSpace(s))
			running = false
		}

		select {
		case signal := <-quit:
			infoLogger.Printf("Receive a signal: %d, and I will shutdown gracefully...", signal)
			running = false
			quit <- signal
		case <-time.Tick(interval):
			infoLogger.Printf("Will write another message...")
		}
	}
}
