package repositories

import "errors"

// ErrSignatureNotFound is returned when a document signature lookup finds no
// matching row. Sentinel so the HTTP layer can errors.Is it and map to 404.
//
// Issue: #140
var ErrSignatureNotFound = errors.New("document: signature not found")
