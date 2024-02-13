// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import "net/http"

// Binding describes the interface which needs to be implemented for binding the
// data present in the request such as JSON request body, query parameters or
// the form POST.
type Binding interface {
	Name() string
	Bind(*http.Request, any) error
}

// These implement the Binding interface and can be used to bind the data
// present in the request to struct instances.
var (
	Form     = formBinding{}
	Query    = queryBinding{}
	FormPost = formPostBinding{}
)
