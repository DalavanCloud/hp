// Copyright 2011 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os/exec"
	"encoding/json"
	"net/http"
	"log"
	"fmt"
	"text/template"
)

func (s *state) JS() []byte {
	js := make(map[string]interface{})
	js["total"] = s.Profile.Header.InuseBytes/1024
	js["sizes"] = s.Graph.NodeSizes
	jsbytes, err := json.Marshal(js)
	check(err)
	return jsbytes
}

func (s *state) ServeHttp(addr string) {
	// This seems pretty suboptimal, but I can't figure out how else
	// to define functions before loading a template.
	tmpl := template.Must(template.New("page").Funcs(template.FuncMap{
		"kb": func(n int) string {
			return fmt.Sprintf("%dkb", n/1024)
		},
		"firstn": func(n int, xs []int) []int {
			return xs[:n]
		},
		"json": func(x interface{}) (string, error) {
			js, err := json.Marshal(x)
			return string(js), err
		},
	}).ParseFiles("page.html")).Lookup("page.html")

	http.HandleFunc("/t.png", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "t.png")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
		} else {
			err := tmpl.Execute(w, s)
			check(err)
			//http.ServeFile(w, req, "page.html")
		}
	})
	go func() {
		url := addr
		if url[0] == ':' {
			url = "http://localhost" + url
		}
		log.Printf("spawning browser on %s", url)
		exec.Command("gnome-open", url).Start()
	}()
	log.Fatal(http.ListenAndServe(addr, nil))
}