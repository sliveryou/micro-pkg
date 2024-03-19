package xreq_test

import (
	"context"
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

	client := xreq.NewClient(
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

	client := xreq.NewClient(
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

func ExampleClient_Call() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer server.Close()

	client := xreq.NewClient(
		xreq.URL(server.URL),
		xreq.Path("api", "/students"),
		xreq.BearerAuth("abcdefgh"),
	)

	result := make(map[string]string)
	response, _ := client.Call(http.MethodPost, &result, xreq.BodyJSON(map[string]string{"name": "SliverYou", "language": "go"}))

	fmt.Println(response.IsSuccess())
	fmt.Println(response.IsError())
	fmt.Println(response.ContentType())
	fmt.Println(response.Size())
	fmt.Println(response.String())
	fmt.Println(result)

	// Output:
	// true
	// false
	// application/json; charset=utf-8
	// 36
	// {"language":"go","name":"SliverYou"}
	// map[language:go name:SliverYou]
}

func ExampleClient_DoWithRequest() {
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

	client := xreq.NewClient(
		xreq.BearerAuth("abcdefgh"),
	)

	request, _ := xreq.NewPost(server.URL,
		xreq.Context(context.TODO()),
		xreq.Path("api", "/students"),
		xreq.BodyJSON(map[string]string{"name": "SliverYou", "language": "go"}),
	)

	client.DoWithRequest(request)

	// Output:
	// POST
	// /api/students
	// {"language":"go","name":"SliverYou"}
	// 36
	// application/json
	// Bearer abcdefgh
}

func ExampleClient_CallWithRequest() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer server.Close()

	client := xreq.NewClient(
		xreq.BearerAuth("abcdefgh"),
	)

	request, _ := xreq.NewPost(server.URL,
		xreq.Context(context.TODO()),
		xreq.Path("api", "/students"),
		xreq.BodyJSON(map[string]string{"name": "SliverYou", "language": "go"}),
	)

	result := make(map[string]string)
	response, _ := client.CallWithRequest(request, &result)

	fmt.Println(response.IsSuccess())
	fmt.Println(response.IsError())
	fmt.Println(response.ContentType())
	fmt.Println(response.Size())
	fmt.Println(response.String())
	fmt.Println(result)

	// Output:
	// true
	// false
	// application/json; charset=utf-8
	// 36
	// {"language":"go","name":"SliverYou"}
	// map[language:go name:SliverYou]
}
