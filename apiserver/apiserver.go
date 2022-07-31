package apiserver

import (
	"net/http"

)

type APIServer struct {}

func New() *APIServer {
	return &APIServer{}
}

func (s *APIServer) Start() error {
	s.configureRouter()
	return http.ListenAndServe("0.0.0.0:80", nil)
}

func (s *APIServer) configureRouter() {
	http.HandleFunc("/hello", s.HandleHello())
}


func (s *APIServer) HandleHello() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(("Hello world")))
	}
}
