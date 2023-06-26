package soda

import "context"

type customizeValidateCtx interface {
	Validate(ctx context.Context) error
}

type customizeValidate interface {
	Validate() error
}
