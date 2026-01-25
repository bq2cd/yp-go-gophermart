// Binary gophermart provides a CLI entry point to launch GopherMart HTTP API and
// corresponding processing sub-systems, e.g. order processing, etc.
package main

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/app"

func main() {
	app.Run()
}
