package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	socketio "github.com/googollee/go-socket.io"
)

type tvconn struct {
	ID     string
	Conn   socketio.Conn
	TVCode string
}

var tvconnStorage = []tvconn{}

func main() {

	//socket io storage
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")

		fmt.Println("connected:", s.ID())
		tvcode := geneTVCode()
		addSocket(s.ID(), s, tvcode)

		s.Emit("connected", tvcode)
		return nil
	})
	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		fmt.Println("notice:", msg)
		s.Emit("reply", "have "+msg)
	})
	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		return "recv " + msg
	})
	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})
	server.OnError("/", func(e error) {
		fmt.Println("meet error:", e)
	})
	server.OnDisconnect("/", func(s socketio.Conn, msg string) {
		fmt.Println("closed", msg)
		fmt.Println(s.ID())
		removeSocket(s.ID())
	})
	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/control", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		keys, ok := req.URL.Query()["tvcode"]
		ctls, ok := req.URL.Query()["ctls"]

		if !ok || len(keys[0]) < 1 {
			log.Println("Url Param 'tvcode' is missing or 'ctls' is missing ")
			res.WriteHeader(200)
			res.Write([]byte("Url Param 'tvcode' or 'ctls' is missing"))
			return
		}

		// Query()["key"] will return an array of items,
		// we only want the single item.
		key := keys[0]
		ctl := ctls[0]

		err, conn := getSocket(key)
		if err != nil {
			res.WriteHeader(200)
			res.Write([]byte(string(err.Error())))
			return
		}
		conn.Emit("control", ctl)

		res.WriteHeader(200)
		res.Write([]byte("ok"))
	}))

	http.Handle("/exists", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		keys, ok := req.URL.Query()["tvcode"]

		if !ok || len(keys[0]) < 1 {
			log.Println("Url Param 'tvcode' is missing.")
			res.WriteHeader(200)
			res.Write([]byte("Url Param 'tvcode' is missing."))
			return
		}

		// Query()["key"] will return an array of items,
		// we only want the single item.
		key := keys[0]

		err, _ := getSocket(key)
		if err != nil {
			res.WriteHeader(200)
			res.Write([]byte(string(err.Error())))
			return
		}

		res.WriteHeader(200)
		res.Write([]byte("FIND"))
	}))

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/controller", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/controller.html")
	})
	log.Println("Serving at localhost:80...")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func getSocket(tvcode string) (error, socketio.Conn) {
	for _, v := range tvconnStorage {
		fmt.Println(v.ID)
		if tvcode == v.TVCode {
			return nil, v.Conn
		}
	}
	return errors.New("TVCode Not Found"), nil
}

func removeSocket(id string) {
	for i, v := range tvconnStorage {
		if id == v.ID {
			// Remove the element at index i from a.
			tvconnStorage[i] = tvconnStorage[len(tvconnStorage)-1] // Copy last element to index i.
			tvconnStorage[len(tvconnStorage)-1] = tvconn{}         // Erase last element (write zero value).
			tvconnStorage = tvconnStorage[:len(tvconnStorage)-1]   // Truncate slice.

			fmt.Println(tvconnStorage) // [A B E D]
		}
	}
}

func addSocket(id string, s socketio.Conn, tvcode string) {
	tvconnStorage = append(tvconnStorage, tvconn{
		ID:     id,
		Conn:   s,
		TVCode: tvcode,
	})
}

//define control type
//define tv m3u8 source
func sendControl() {

}

func geneTVCode() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789")
	length := 4
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String() // E.g. "ExcbsVQs"
	return str
}
