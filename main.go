package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"net/http"
	"net/http/httputil"

	"github.com/urfave/cli/v2"
)

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
					Serve()
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
