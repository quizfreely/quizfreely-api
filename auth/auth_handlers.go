package auth

import (
	"net/http"
)

type SignUpReqBody struct {
	Username string `json:"username"`
	NewPassword string `json:"password"`
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody SignUpReqBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", 400)
		return
	}
	w.WriteHeader(201)
	w.W
}
