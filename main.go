package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/cloudDatastoreExport", exportHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Hello, World!")
}

func exportHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ts, err := google.DefaultTokenSource(ctx,
		"https://www.googleapis.com/auth/datastore")
	if err != nil {
		log.Printf("Get token failed. err: %v", err)
	}
	client := oauth2.NewClient(ctx, ts)

	q := r.URL.Query()

	type M map[string]interface{}
	type I []interface{}
	kindI := I{}
	for _, k := range q["kind"] {
		kindI = append(kindI, k)
	}
	log.Println(q["outputUrlPrefix"])
	reqBody := M{
		"outputUrlPrefix": q["outputUrlPrefix"][0],
		"entityFilter": M{
			"kinds": kindI,
		}}
	b, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Marshal request body failed. err: %v", err)
		return
	}

	req, err := http.NewRequest(
		"POST",
		"https://datastore.googleapis.com/v1/projects/{YOUR_PROJECT_ID}:export",
		bytes.NewBuffer(b))

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed. err: %#v", err)
		return
	}

	if res.StatusCode != http.StatusOK {
		respb, _ := ioutil.ReadAll(res.Body)
		log.Printf("Request is not OK. ReponseBody: %s", string(respb))
		return
	}

	log.Printf("Request success. Response: %v", res)
	return
}