package main

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"cloud.google.com/go/errorreporting"
)

func (b *Bookshelf) sendLog(w http.ResponseWriter, r *http.Request) *appError {
	fmt.Fprintln(b.logWriter, "log sent")

	fmt.Fprintln(w, `<html>Log sent! Check the <a href="http://console.cloud.google.com/logs">logging section of the Cloud Console</a>.</html>`)

	return nil
}

func (b *Bookshelf) sendError(w http.ResponseWriter, r *http.Request) *appError {
	msg := `<html>Logging an error. Check <a href="http://console.cloud.google.com/errors">Error Reporting</a> (it may take a minute or two for the error to appear).</html>`
	err := errors.New("uh oh! an error occurred")
	return b.appErrorf(r, err, msg)
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	err     error
	message string
	code    int
	req     *http.Request
	b       *Bookshelf
	stack   []byte
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil {
		fmt.Fprintf(e.b.logWriter, "Handler error (reported to Error Reporting): status code: %d, message: %s, underlying err: %+v\n", e.code, e.message, e.err)
		w.WriteHeader(e.code)
		// fmt.Fprint(w, e.message)

		e.b.errorClient.Report(errorreporting.Entry{
			Error: e.err,
			Req:   r,
			Stack: e.stack,
		})
		e.b.errorClient.Flush()
	}
}

func (b *Bookshelf) appErrorf(r *http.Request, err error, format string, v ...interface{}) *appError {
	return &appError{
		err:     err,
		message: fmt.Sprintf(format, v...),
		code:    500,
		req:     r,
		b:       b,
		stack:   debug.Stack(),
	}
}

func (b *Bookshelf) addMessage(message string) {
	b.Messages = append(b.Messages, message)
}
