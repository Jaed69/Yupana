package main

import (
	"context"
	"net/http"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	_ = anthropic.NewClient()
	_, _ = pgxpool.ParseConfig("postgres://u:p@localhost/db")
	_ = context.Background()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"ok":true}`)) })
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
