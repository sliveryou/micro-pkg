package xreq_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/sliveryou/micro-pkg/xhttp/xreq"
)

func ExampleNew() {
	ctx := context.TODO()
	request, _ := xreq.New(http.MethodPost, "http://www.test.com/api",
		xreq.Context(ctx),
		xreq.Scheme("https"),
		xreq.Host("my.test.com"),
		xreq.AddPath("/students"),
		xreq.BearerAuth("abcdefgh"),
		xreq.BodyJSON(map[string]string{"name": "SliverYou", "language": "go"}),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(request.Method)
	fmt.Println(request.URL)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)
	fmt.Println(request.Header.Get("Content-Type"))
	fmt.Println(request.Header.Get("Authorization"))

	// Output:
	// POST
	// https://my.test.com/api/students
	// {"language":"go","name":"SliverYou"}
	// 36
	// application/json
	// Bearer abcdefgh
}

func ExampleNewGet() {
	ctx := context.TODO()
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.Context(ctx),
		xreq.AddPath("files"),
		xreq.Queries(url.Values{
			"page":      {"10"},
			"page_size": {"50"},
		}),
	)

	fmt.Println(request.Method)
	fmt.Println(request.URL)

	// Output:
	// GET
	// https://www.test.com/api/files?page=10&page_size=50
}

func ExampleNewPost() {
	ctx := context.TODO()
	request, _ := xreq.NewPost("https://www.test.com/api",
		xreq.Context(ctx),
		xreq.AddPath("/students"),
		xreq.BodyJSON(map[string]string{"name": "SliverYou", "language": "go"}),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(request.Method)
	fmt.Println(request.URL)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)
	fmt.Println(request.Header.Get("Content-Type"))

	// Output:
	// POST
	// https://www.test.com/api/students
	// {"language":"go","name":"SliverYou"}
	// 36
	// application/json
}

func ExampleNewPut() {
	ctx := context.TODO()
	request, _ := xreq.NewPut("https://www.test.com/api",
		xreq.Context(ctx),
		xreq.AddPath("/students"),
		xreq.BodyJSON(map[string]any{"id": 1, "name": "SliverYou", "language": "go"}),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(request.Method)
	fmt.Println(request.URL)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)
	fmt.Println(request.Header.Get("Content-Type"))

	// Output:
	// PUT
	// https://www.test.com/api/students
	// {"id":1,"language":"go","name":"SliverYou"}
	// 43
	// application/json
}

func ExampleNewPatch() {
	ctx := context.TODO()
	request, _ := xreq.NewPatch("https://www.test.com/api",
		xreq.Context(ctx),
		xreq.AddPath("/students"),
		xreq.BodyJSON(map[string]any{"id": 1, "name": "SliverYou"}),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(request.Method)
	fmt.Println(request.URL)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)
	fmt.Println(request.Header.Get("Content-Type"))

	// Output:
	// PATCH
	// https://www.test.com/api/students
	// {"id":1,"name":"SliverYou"}
	// 27
	// application/json
}

func ExampleNewDelete() {
	ctx := context.TODO()
	request, _ := xreq.NewDelete("https://www.test.com/api",
		xreq.Context(ctx),
		xreq.AddPath("/file/1"),
		xreq.Query("file_type", "1"),
	)

	fmt.Println(request.Method)
	fmt.Println(request.URL)

	// Output:
	// DELETE
	// https://www.test.com/api/file/1?file_type=1
}

func ExampleNewHead() {
	ctx := context.TODO()
	request, _ := xreq.NewHead("https://www.test.com/api",
		xreq.Context(ctx),
		xreq.AddPath("files"),
		xreq.Queries(url.Values{
			"page":      {"10"},
			"page_size": {"50"},
		}),
	)

	fmt.Println(request.Method)
	fmt.Println(request.URL)

	// Output:
	// HEAD
	// https://www.test.com/api/files?page=10&page_size=50
}

func ExampleNewOptions() {
	ctx := context.TODO()
	request, _ := xreq.NewOptions("https://www.test.com/api",
		xreq.Context(ctx),
		xreq.AddPath("/students"),
		xreq.Headers(http.Header{
			"Origin":                         {"https://my.test.com"},
			"Access-Control-Request-Method":  {"PUT"},
			"Access-Control-Request-Headers": {"Authorization", "Content-Type"},
		}),
		xreq.BodyJSON(map[string]any{"id": 1, "name": "SliverYou", "language": "go"}),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(request.Method)
	fmt.Println(request.URL)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)
	fmt.Println(request.Header.Get("Content-Type"))

	// Output:
	// OPTIONS
	// https://www.test.com/api/students
	// {"id":1,"language":"go","name":"SliverYou"}
	// 43
	// application/json
}

func ExampleDo() {
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

	xreq.Do(http.MethodPost, server.URL,
		xreq.Path("api", "/students"),
		xreq.BearerAuth("abcdefgh"),
		xreq.BodyJSON(map[string]string{"name": "SliverYou", "language": "go"}),
	)

	// Output:
	// POST
	// /api/students
	// {"language":"go","name":"SliverYou"}
	// 36
	// application/json
	// Bearer abcdefgh
}
