package main

import (
	"log"
	"os"

	"infograf/internal/app"
)

func main() {
	addr := os.Getenv("INFOGRAF_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	dataDir := os.Getenv("INFOGRAF_DATA_DIR")
	dbPath := os.Getenv("INFOGRAF_DB_PATH")

	srv, err := app.New(app.Options{DataDir: dataDir, DBPath: dbPath})
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Close()

	log.Printf("Infograf listening on %s", addr)
	if err := srv.ListenAndServe(addr); err != nil {
		log.Fatal(err)
	}
}
