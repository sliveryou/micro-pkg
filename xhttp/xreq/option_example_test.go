package xreq_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/sliveryou/micro-pkg/xhttp/xreq"
)

func ExampleApply() {
	request, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "", nil)
	request, _ = xreq.Apply(request,
		xreq.URL("https://www.test.com"),
		xreq.Path("api", "/students"),
		xreq.BearerAuth("abcdefgh"),
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

func ExampleOptionFunc() {
	token := "custom.token.1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Header.Get("Authorization"))
	}))
	defer server.Close()

	client := xreq.NewClientWithHTTPClient(server.Client(),
		xreq.URL(server.URL),
		xreq.OptionFunc(func(request *http.Request) (*http.Request, error) {
			return xreq.BearerAuth(token).Apply(request)
		}),
	)

	client.Get()
	token = "custom.token.2"
	client.Get()

	// Output:
	// Bearer custom.token.1
	// Bearer custom.token.2
}

func ExampleMethod() {
	request, _ := xreq.New("", "https://www.test.com/api", xreq.Method(http.MethodPost))

	fmt.Println(request.Method)
	fmt.Println(request.URL)

	// Output:
	// POST
	// https://www.test.com/api
}

func ExampleRawURL() {
	rawURL, _ := url.Parse("http://www.test.com/api")
	rawURL.Scheme = "https"
	request, _ := xreq.NewGet("", xreq.RawURL(rawURL))

	fmt.Println(request.URL)

	// Output: https://www.test.com/api
}

func ExampleURL() {
	request, _ := xreq.NewGet("", xreq.URL("https://www.test.com/api"))

	fmt.Println(request.URL)

	// Output: https://www.test.com/api
}

func ExampleScheme() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.Scheme("http"),
	)

	fmt.Println(request.URL)

	// Output: http://www.test.com/api
}

func ExampleUser() {
	user := url.UserPassword("SliverYou", "123456")
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.User(user),
	)

	fmt.Println(request.URL)

	// Output: https://SliverYou:123456@www.test.com/api
}

func ExampleUsername() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.Username("SliverYou"),
	)

	fmt.Println(request.URL)

	// Output: https://SliverYou@www.test.com/api
}

func ExampleUserPassword() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.UserPassword("SliverYou", "123456"),
	)

	fmt.Println(request.URL)

	// Output: https://SliverYou:123456@www.test.com/api
}

func ExampleHost() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.Host("api.test.com"),
	)

	fmt.Println(request.Host)
	fmt.Println(request.URL)

	// Output:
	// api.test.com
	// https://api.test.com/api
}

func ExamplePath() {
	request, _ := xreq.NewGet("https://www.test.com",
		xreq.Path("api", "/tests", "1234/", "cases"),
	)

	fmt.Println(request.URL.Path)

	// Output: /api/tests/1234/cases
}

func ExampleAddPath() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.AddPath("/tests", "1234/", "cases"),
	)

	fmt.Println(request.URL.Path)

	// Output: /api/tests/1234/cases
}

func ExampleQuery() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.Query("page", 1),
		xreq.Query("page", 2),
		xreq.Query("page", 3),
	)

	fmt.Println(request.URL.Query().Encode())

	// Output: page=3
}

func ExampleAddQuery() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.AddQuery("page", 1),
		xreq.AddQuery("page", 2),
		xreq.AddQuery("page", 3),
	)

	fmt.Println(request.URL.Query().Encode())

	// Output: page=1&page=2&page=3
}

func ExampleQueries() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.Queries(url.Values{
			"page":      {"10"},
			"page_size": {"50"},
		}),
		xreq.Queries(url.Values{
			"page":      {"20"},
			"page_size": {"50"},
		}),
	)

	fmt.Println(request.URL.Query().Encode())

	// Output: page=20&page_size=50
}

func ExampleAddQueries() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.AddQueries(url.Values{
			"page":      {"10"},
			"page_size": {"50"},
		}),
		xreq.AddQueries(url.Values{
			"page": {"20", "30"},
		}),
	)

	fmt.Println(request.URL.Query().Encode())

	// Output: page=10&page=20&page=30&page_size=50
}

func ExampleQueryMap() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.QueryMap(map[string]any{
			"page":      10,
			"page_size": 50,
		}),
		xreq.QueryMap(map[string]any{
			"page":      20,
			"page_size": 50,
		}),
	)

	fmt.Println(request.URL.Query().Encode())

	// Output: page=20&page_size=50
}

func ExampleAddQueryMap() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.AddQueryMap(map[string]any{
			"page":      10,
			"page_size": 50,
		}),
		xreq.AddQueryMap(map[string]any{
			"page": 20,
		}),
	)

	fmt.Println(request.URL.Query().Encode())

	// Output: page=10&page=20&page_size=50
}

func ExampleHeader() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.Header("key", "value1"),
		xreq.Header("key", "value2", "value3"),
	)

	fmt.Println(request.Header.Get("key"))

	// Output: value2, value3
}

func ExampleHeaders() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.Headers(http.Header{
			"a": {"1"},
			"b": {"2", "3"},
		}),
		xreq.Headers(http.Header{
			"a": {"2", "3"},
			"c": {"4"},
		}),
	)

	fmt.Println(request.Header)

	// Output: map[A:[2, 3] B:[2, 3] C:[4]]
}

func ExampleHeaderMap() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.HeaderMap(map[string]string{
			"a": "1",
			"b": "2",
		}),
	)

	fmt.Println(request.Header)

	// Output: map[A:[1] B:[2]]
}

func ExampleAccept() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.Accept("application/json"),
	)

	fmt.Println(request.Header.Get("Accept"))

	// Output: application/json
}

func ExampleAuthorization() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.Authorization("abcdefgh"),
	)

	fmt.Println(request.Header.Get("Authorization"))

	// Output: abcdefgh
}

func ExampleCacheControl() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.CacheControl("no-cache"),
	)

	fmt.Println(request.Header.Get("Cache-Control"))

	// Output: no-cache
}

func ExampleContentLength() {
	s := "example body"
	request, _ := xreq.NewPost("https://www.test.com/api",
		xreq.Body(io.NopCloser(strings.NewReader(s))),
		xreq.ContentLength(len(s)),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)

	// Output:
	// example body
	// 12
}

func ExampleContentType() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.ContentType("application/json"),
	)

	fmt.Println(request.Header.Get("Content-Type"))

	// Output: application/json
}

func ExampleReferer() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.Referer("https://test.com"),
	)

	fmt.Println(request.Header.Get("Referer"))

	// Output: https://test.com
}

func ExampleUserAgent() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.UserAgent("xreq"),
	)

	fmt.Println(request.Header.Get("User-Agent"))

	// Output: xreq
}

func ExampleAddCookies() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.AddCookies(
			&http.Cookie{
				Name: "cookie-1", Value: "v$1",
				Expires: time.Now().Add(time.Hour),
			},
			&http.Cookie{
				Name: "cookie-2", Value: "v$2",
				Expires: time.Now().Add(time.Hour),
			},
		),
	)

	fmt.Println(request.Header.Get("Cookie"))

	// Output: cookie-1=v$1; cookie-2=v$2
}

func ExampleBasicAuth() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.BasicAuth("SliverYou", "123456"),
	)

	fmt.Println(request.Header.Get("Authorization"))

	// Output: Basic U2xpdmVyWW91OjEyMzQ1Ng==
}

func ExampleBearerAuth() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.BearerAuth("abcdefgh"),
	)

	fmt.Println(request.Header.Get("Authorization"))

	// Output: Bearer abcdefgh
}

func ExampleTokenAuth() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.TokenAuth("abcdefgh"),
	)

	fmt.Println(request.Header.Get("Authorization"))

	// Output: Token abcdefgh
}

func ExampleContext() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.Context(context.Background()),
	)

	fmt.Println(request.Context())

	// Output: context.Background
}

func ExampleContextValue() {
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.ContextValue("key", "value"),
	)

	fmt.Println(request.Context().Value("key"))

	// Output: value
}

func ExampleBody() {
	request, _ := xreq.NewPost("https://www.test.com/api",
		xreq.Body(io.NopCloser(strings.NewReader("example body"))),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(string(body))

	// Output: example body
}

func ExampleBodyReader() {
	request, _ := xreq.NewPost("https://www.test.com/api",
		xreq.BodyReader(strings.NewReader("example body")),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)

	// Output:
	// example body
	// 12
}

func ExampleBodyBytes() {
	request, _ := xreq.NewPost("https://www.test.com/api",
		xreq.BodyBytes([]byte("example body")),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)

	// Output:
	// example body
	// 12
}

func ExampleBodyString() {
	request, _ := xreq.NewPost("https://www.test.com/api",
		xreq.BodyString("example body"),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)

	// Output:
	// example body
	// 12
}

func ExampleBodyForm() {
	request, _ := xreq.NewPost("https://www.test.com/api",
		xreq.BodyForm(url.Values{
			"page":      {"10"},
			"page_size": {"50"},
		}),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)
	fmt.Println(request.Header.Get("Content-Type"))

	// Output:
	// page=10&page_size=50
	// 20
	// application/x-www-form-urlencoded
}

func ExampleBodyFormMap() {
	request, _ := xreq.NewPost("https://www.test.com/api",
		xreq.BodyFormMap(map[string]any{
			"page":      10,
			"page_size": 50,
		}),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)
	fmt.Println(request.Header.Get("Content-Type"))

	// Output:
	// page=10&page_size=50
	// 20
	// application/x-www-form-urlencoded
}

func ExampleBodyJSON() {
	request, _ := xreq.NewPost("https://www.test.com/api",
		xreq.BodyJSON(map[string]string{"name": "SliverYou", "language": "go"}),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)
	fmt.Println(request.Header.Get("Content-Type"))

	// Output:
	// {"language":"go","name":"SliverYou"}
	// 36
	// application/json
}

func ExampleBodyXML() {
	request, _ := xreq.NewPost("https://www.test.com/api",
		xreq.BodyXML("SliverYou"),
	)

	body, _ := io.ReadAll(request.Body)
	fmt.Println(string(body))
	fmt.Println(request.ContentLength)
	fmt.Println(request.Header.Get("Content-Type"))

	// Output:
	// <string>SliverYou</string>
	// 26
	// application/xml
}

func ExampleDump() {
	var buffer bytes.Buffer
	request, _ := xreq.NewGet("https://www.test.com/api",
		xreq.BodyString("example body"),
		xreq.Dump(&buffer),
	)

	fmt.Println(request.ContentLength)
	fmt.Println(buffer.String() == "GET /api HTTP/1.1\r\nHost: www.test.com\r\n\r\nexample body")

	// Output:
	// 12
	// true
}
