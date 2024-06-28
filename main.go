package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got / request\n")
	io.WriteString(w, "This is my website!\n")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var connections = make(map[string]*websocket.Conn)

func RemoveIndex(s []*websocket.Conn, index int) []*websocket.Conn {
	return append(s[:index], s[index+1:]...)
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("New Client", conn.RemoteAddr().String())

	connections[conn.RemoteAddr().String()] = conn

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			delete(connections, conn.RemoteAddr().String())
			fmt.Println("Client Disconnected", conn.RemoteAddr().String())
			return
		}

		for _, connection := range connections {
			if connection.RemoteAddr().String() != conn.RemoteAddr().String() {
				if err := connection.WriteMessage(messageType, p); err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/test", getRoot)
	http.HandleFunc("/socket", handler)

	err := http.ListenAndServe(":3333", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
