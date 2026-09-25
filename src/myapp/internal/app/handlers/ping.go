package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/angelomira/myapp/utils"
)

type pingResp struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

func Ping(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest(r)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(pingResp{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	})
}