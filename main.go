package main

import (
	"log"

	"final_project/pkg/db"
	"final_project/pkg/server"
)

func main() {
	if err := db.Init("scheduler.db"); err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
