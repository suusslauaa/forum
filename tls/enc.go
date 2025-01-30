package tls

import "embed"

//go:embed *.pem
var Pems embed.FS
