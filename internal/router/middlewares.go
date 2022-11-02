package router

import (
	"compress/gzip"
	l "github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	"net/http"
)

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gzipWrappedBody, err := gzip.NewReader(r.Body)
			if err != nil {
				l.Logger.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer func() {
				if err := gzipWrappedBody.Close(); err != nil {
					l.Logger.Warn(err.Error())
				}
			}()
			r.Body = gzipWrappedBody
		}
		next.ServeHTTP(w, r)
	})
}
