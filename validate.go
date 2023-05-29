package soda

import "context"

type CustomizeValidate interface {
	Validate(ctx context.Context) error
}
