package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// LogFlag 控制日志的前缀
	LogFlag = log.LstdFlags | log.Lmicroseconds | log.Lshortfile
	// Interval 表示重连或者写数据的间隔时间
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
		conn, err := net.Dial("tcp", "echod:8080")
		if err != nil {
			errLogger.Printf("net.Dial() failed, error: %s.", err)
		} else {
			infoLogger.Printf("Accept a connection, RemoteAddr: %s.", conn.RemoteAddr())

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

		infoLogger.Printf("Write %d bytes.", n)

		s, err := buf.ReadString('\n')
		if err != nil {
			errLogger.Printf("buf.ReadString() failed, error: %s.", err)
			running = false
		}

		if s != message {
			errLogger.Printf("Wrong message, want: %s, got: %s.", message, s)
			running = false
		}

		select {
		case signal := <-quit:
			infoLogger.Printf("Receive a signal: %d, and I will shutdown gracefully...", signal)
			running = false
			quit <- signal
		case <-time.Tick(interval):
			infoLogger.Printf("Ready for another message.")
		}
	}
}
