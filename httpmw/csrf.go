package httpmw

import (
	"net/http"

	"github.com/authenticvision/util-go/httpp"
)

// NewCrossOriginProtection is the httpp equivalent to http.NewCrossOriginProtection.
func NewCrossOriginProtection() *CrossOriginProtection {
	return &CrossOriginProtection{}
}

// CrossOriginProtection wraps http.CrossOriginProtection as Middleware.
type CrossOriginProtection struct {
	http.CrossOriginProtection
}

func (c *CrossOriginProtection) Middleware(handler httpp.Handler) httpp.Handler {
	next := c.Handler(httpp.CollectErrors(handler))
	return httpp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		errReq, err := httpp.WithErrorCollector(r)
		next.ServeHTTP(w, errReq)
		return *err
	})
}
