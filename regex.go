package soda

import "regexp"

var (
	regexOperationID = regexp.MustCompile("[^a-zA-Z0-9]+")
	regexFiberPath   = regexp.MustCompile("/:([0-9a-zA-Z_]+)")
	regexSchemaName  = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
)
