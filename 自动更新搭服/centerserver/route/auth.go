package route

import (
	"net/http"
)

func auth(r *http.Request) bool {
	return true
}
