package assets

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const StaticPrefix = "/static/"

// ServeStatic returns a middleware handler that will serve static files from
// the web/static directory under the /static URL prefix.
func ServeStatic() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, StaticPrefix) {
			// This URL doesn't fall under our path, return and allow another handler
			// to handle this request.
			return
		}

		// Look for the asset in our bindata, if it's not found then return a 404,
		// otherwise return the contents of the asset and infer it's content type.
		// In either case this asset is contained under our path, so we will abort
		// the request and not allow any other handlers to process it.
		filename := strings.TrimLeft(c.Request.URL.Path, "/")
		bs, err := Asset(filename)
		if err != nil {
			_ = c.AbortWithError(http.StatusNotFound, err)
			return
		}

		c.Data(http.StatusOK, http.DetectContentType(bs), bs)
		c.Abort()
	}
}
