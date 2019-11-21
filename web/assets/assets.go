package assets

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const StaticPrefix = "/static/"

var MissingDigest = [sha256.Size]byte{}

// ContentTypes contains a mapping of filename suffix to the content type that
// should be used for that suffix.
var ContentTypes = map[string]string{
	".css": "text/css",
	".ico": "image/x-icon",
	".js":  "application/javascript",
	".map": "application/octet-stream",
}

// ServeStatic returns a middleware handler that will serve static files from
// the web/static directory under the /static URL prefix.
func ServeStatic() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, StaticPrefix) {
			// This URL doesn't fall under our path, return and allow another handler
			// to handle this request.
			return
		}

		// Look for the asset in our bindata, if it's not found then return a 404.
		filename := strings.TrimLeft(c.Request.URL.Path, "/")
		digest, err := AssetDigest(filename)
		if err != nil {
			_ = c.AbortWithError(http.StatusNotFound, err)
			return
		}

		// The bytes of the asset, we may need to load these ourselves in order to
		// compute the etag of the asset, so keep track of them so that we don't
		// load them twice.
		var bs []byte

		// When running in development mode, bindata doesn't provide a digest so
		// make sure that we detect this case and compute it ourselves.
		if digest == MissingDigest {
			bs = MustAsset(filename)
			digest = sha256.Sum256(bs)
		}

		// Now that we found the asset, check and see if the client has it cached
		// already.
		etag := base64.URLEncoding.EncodeToString(digest[:])
		if etag == c.GetHeader("If-None-Match") {
			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		// If we've gotten here then we know the client doesn't have the latest
		// version of the asset, load it if we haven't already and serve its
		// contents up to the caller.
		if bs != nil {
			bs = MustAsset(filename)
		}

		// Determine the content type that should be used to serve this file.
		var contentType = ContentTypes[filepath.Ext(filename)]
		if contentType == "" {
			contentType = http.DetectContentType(bs)
		}

		c.Header("Etag", etag)
		c.Data(http.StatusOK, contentType, bs)
		c.Abort()
	}
}
