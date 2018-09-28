package dump

import (
	"Envoy-Pilot/cmd/server/mapper"
	"Envoy-Pilot/cmd/server/storage"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func SetUpHttpServer() {
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/dump/cds/", configDumpCDS)
	http.HandleFunc("/dump/lds/", configDumpLDS)
	log.Println("Starting http server on :9090..")
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", "pong")
}

func configDumpCDS(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	keyPath := strings.Replace(path, "dump/cds/", "", -1)
	log.Printf("Getting cds dump for %s\n", keyPath)
	m := &mapper.ClusterMapper{}
	cwrapper := storage.GetConsulWrapper()
	jsonStr := cwrapper.GetString(keyPath)
	fmt.Printf("json dump %s\n", jsonStr)
	val, err := m.GetClusters(jsonStr)

	if err != nil {
		fmt.Fprintf(w, "Error creating obj %s", err)
		return
	}
	resJson, err := json.Marshal(val)
	if err != nil {
		fmt.Fprintf(w, "Error parsing json %s", err)
		return
	}

	fmt.Fprintf(w, "%s", resJson)
}

func configDumpLDS(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	keyPath := strings.Replace(path, "dump/lds/", "", -1)
	log.Printf("Getting lds dump for %s\n", keyPath)
	m := &mapper.ListenerMapper{}
	cwrapper := storage.GetConsulWrapper()
	jsonStr := cwrapper.GetString(keyPath)
	fmt.Printf("json dump %s\n", jsonStr)
	val, err := m.GetListeners(jsonStr)

	if err != nil {
		fmt.Fprintf(w, "Error creating obj %s", err)
		return
	}
	resJson, err := json.Marshal(val)
	if err != nil {
		fmt.Fprintf(w, "Error parsing json %s", err)
		return
	}

	fmt.Fprintf(w, "%s", resJson)
}
