package middlewares

import "net/http"

// Resource: https://www.youtube.com/watch?v=H7tbjKFSg58&t=33s
type Middleware func(http.Handler) http.Handler
