package db

// $PGHOST, $PGUSER, $PGPORT etc. must be set to run this generate script.
//go:generate env GO111MODULE=on go run schemadoc/main.go frontend schema.md
//go:generate env GO111MODULE=on go run schemadoc/main.go codeintel schema.codeintel.md
