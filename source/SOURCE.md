# Source fonts

These are the Go fonts by Bigelow & Holmes, copied unchanged from
`golang.org/x/image` **v0.46.0**, folder `font/gofont/ttfs`. Their licence is
in `README-gofont` (BSD-3-Clause, the Go licence).

| File | SHA-256 |
|---|---|
| Go-Regular.ttf | `197d9f3703b4c00af609178876a8d73e396f64fe438b2c871778566632374be3` |
| Go-Bold.ttf | `c18494baa7ea35b8dbfac3861787d360b9c0ab91232562371a09e6fbf09aaa40` |
| Go-Mono.ttf | `8bc66a0154bbf69cd24e5bde41a12ec9495c8a242c5def2255e4a164900f1ed7` |
| Go-Mono-Bold.ttf | `b165f1a142dc88d3c388de2c89d3ab62dd04620ef21ad07f43fb51f77857edc3` |

To update them, copy the same files from a newer `golang.org/x/image`, update
the version and checksums here, and run `make`.
