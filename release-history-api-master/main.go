package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/applift/release-history-api/db"
	"github.com/applift/release-history-api/server"
)

func main() {
	connStr := os.Getenv("POSTGRES_CONNECTION_STRING")
	if connStr == "" {
		connStr = "postgres://releasehistory:releasehistorylocal@localhost:5432/releasehistory?sslmode=disable"
	}
	user := os.Getenv("USERNAME")
	if user == "" {
		user = "local"
	}
	password := os.Getenv("PASSWORD")
	if password == "" {
		password = "local123"
	}

	sqlDB := db.GetDB(connStr)
	defer sqlDB.Close()
	handler := server.NewHandler(sqlDB, user, password)
	http.HandleFunc("/health", handler.HealthHandler)
	http.HandleFunc("/deployment", handler.BasicAuth(handler.DeploymentHandler))
	http.HandleFunc("/release", handler.BasicAuth(handler.ReleaseHandler))

	fmt.Println("Server started at port 3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
