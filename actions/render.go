package actions

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr"
)

var r *render.Engine
var assetsBox = packr.NewBox("../public")

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.html",

		// Box containing all of the templates:
		TemplatesBox: packr.NewBox("../templates"),
		AssetsBox:    assetsBox,

		// Add template helpers here:
		Helpers: render.Helpers{
			// uncomment for non-Bootstrap form helpers:
			// "form":     plush.FormHelper,
			// "form_for": plush.FormForHelper,
		},
	})
}

// SRIHandler adds support for Subresource integrity
func SRIHandler(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		jsonstring := r.AssetsBox.Bytes("assets/manifest.json")
		var m map[string]string
		json.Unmarshal(jsonstring, &m)
		for k, v := range m {
			if strings.Contains(k, ".css") || strings.Contains(k, ".js") {
				sha384 := sha512.New384()
				sha5 := sha512.New()

				sha384.Write(r.AssetsBox.Bytes("assets/" + v))
				sha5.Write(r.AssetsBox.Bytes("assets/" + v))
				hash := sha384.Sum(nil)
				hash2 := sha5.Sum(nil)
				k1 := strings.Replace(k, ".", "_", -1)
				c.Set(k1, "sha384-"+base64.StdEncoding.EncodeToString(hash)+" sha512-"+base64.StdEncoding.EncodeToString(hash2))
			}
		}
		return next(c)
	}
}

// SetSecurityHeaders sets security headers
func SetSecurityHeaders(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		number := 32
		b := make([]byte, number)
		_, err := rand.Read(b)
		if err != nil {
			b = []byte("poidewjdewpnwefpoewi0pfiwüdkß§ik")
		}
		nonce := base64.StdEncoding.EncodeToString(b)
		c.Set("nonce", nonce)
		c.Response().Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'strict-dynamic' 'nonce-"+nonce+"' 'self'; img-src 'self'; style-src 'self' 'nonce-"+nonce+"'; form-action 'self'; frame-ancestors 'none'; object-src 'none'; base-uri 'none';")
		return next(c)
	}
}
