package xreq_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/sliveryou/micro-pkg/xhttp/xreq"
)

func ExampleResponse() {
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

	response, _ := client.Post(xreq.BodyJSON(map[string]string{"name": "SliverYou", "language": "go"}))

	result := make(map[string]string)
	response.Unmarshal(&result)

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
