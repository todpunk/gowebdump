package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"net"
	"net/http"
	"net/http/httputil"

	"github.com/urfave/cli/v2"
)

type ConnServer struct {
	port string
	host string
}

func NewServer(host, port string) *ConnServer {
	s := ConnServer{host: host, port: port}
	if port == "" {
		s.port = "7780"
	}
	return &s
}

// Attempt to start the server, and if we can't, exit because what else is there?
func (s *ConnServer) StartServer() {
	listen, err := net.Listen("tcp", s.host+":"+s.port)
	if err != nil {
		log.Fatal("Listening has failed: ", err)
	}

	// We can return immediately and do other stuff if we want.

	func() {
		for {
			conn, err := listen.Accept()

			if err != nil {
				log.Println("Failed to accept new connection: ", err)
				defer conn.Close()
				continue
			}
			go DumpRequest(conn)
		}
	}()
}

func DumpRequest(conn net.Conn) {
	t := time.Now()
	tfmt := t.Format("20060102150405")
	thttpfmt := t.Format(http.TimeFormat)
	defer conn.Close()

	f, err := os.CreateTemp("", "gowebdump-"+tfmt+"-") // in Go version older than 1.17 you can use ioutil.TempFile
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	reader := bufio.NewReader(conn)

	response := []byte("HTTP/1.1 200 OK\nContent-Type: text/html; charset=UTF-8\nServer: gowebdump 0.1\nDate: " + thttpfmt + "Content-Length: 2\n\nOK")
	bodystart := false

	for {
		bytes, _, err := reader.ReadLine()
		if err != nil && err.Error() == "EOF" {
			// We have seen the end of the request!
			break
		}
		if err != nil && err.Error() != "EOF" {
			// We don't know what happened
			log.Println("failed to read line, err:")
			log.Println(err)
			return
		}
		// Write whatever every time, as we want the raw request as is
		line := string(bytes)
		if _, err := f.Write([]byte(line + "\n")); err != nil {
			log.Println("failed to write line, err:")
			log.Println(err)
			return
		}
		if line == "" && bodystart {
			// We have an empty line after the body, request over
			break
		}
		if line == "" && !bodystart {
			// We have an empty line before the body, continue
			bodystart = true
			continue
		}
	}

	fmt.Printf("%s\n", f.Name())
	conn.Write(response)
}

func DmpRoot(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err == nil {
		t := time.Now()
		tfmt := t.Format("20060102150405")
		f, err := os.CreateTemp("", "gowebdump-"+tfmt+"-") // in Go version older than 1.17 you can use ioutil.TempFile
		fmt.Printf("%s\n", f.Name())
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		data := []byte(dump)
		if _, err := f.Write(data); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("%s\n", err)
	}
	return
}

func Serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", DmpRoot)

	err := http.ListenAndServe(":7780", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "serve stuff",
				Action: func(cCtx *cli.Context) error {
					server := NewServer("0.0.0.0", "7780")
					server.StartServer()
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
