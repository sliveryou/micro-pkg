package xreq_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/sliveryou/micro-pkg/xhttp/xreq"
)

func ExampleNewClient() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Println(r.Method)
		fmt.Println(r.URL)
		fmt.Println(string(body))
		fmt.Println(r.ContentLength)
		fmt.Println(r.Header.Get("Content-Type"))
		fmt.Println(r.Header.Get("Authorization"))
	}))
	defer server.Close()

	client := xreq.NewClient(xreq.DefaultHTTPClient,
		xreq.URL(server.URL),
		xreq.Path("api", "/students"),
		xreq.BearerAuth("abcdefgh"),
	)

	client.Post(xreq.BodyJSON(map[string]string{"name": "SliverYou", "language": "go"}))

	// Output:
	// POST
	// /api/students
	// {"language":"go","name":"SliverYou"}
	// 36
	// application/json
	// Bearer abcdefgh
}

func ExampleClient_Do() {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Println(r.Method)
		fmt.Println(r.URL)
		fmt.Println(string(body))
		fmt.Println(r.ContentLength)
		fmt.Println(r.Header.Get("Content-Type"))
		fmt.Println(r.Header.Get("Authorization"))
	}))
	defer server.Close()

	client := xreq.NewClient(xreq.DefaultHTTPClient,
		xreq.URL(server.URL),
		xreq.Path("api", "/students"),
		xreq.BearerAuth("abcdefgh"),
	)

	client.Do(http.MethodPost, xreq.BodyJSON(map[string]string{"name": "SliverYou", "language": "go"}))

	// Output:
	// POST
	// /api/students
	// {"language":"go","name":"SliverYou"}
	// 36
	// application/json
	// Bearer abcdefgh
}
