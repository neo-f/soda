package soda

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func BindHeader(c *gin.Context, obj any) error {
	return binding.Header.Bind(c.Request, obj)
}

func BindQuery(c *gin.Context, obj any) error {
	if err := binding.MapFormWithTag(obj, c.Request.URL.Query(), TagQuery); err != nil {
		return err
	}
	if binding.Validator != nil {
		return binding.Validator.ValidateStruct(obj)
	}
	return nil
}

func BindPath(c *gin.Context, obj any) error {
	m := make(map[string][]string, len(c.Params))
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	if err := binding.MapFormWithTag(obj, m, TagPath); err != nil {
		return err
	}
	if binding.Validator != nil {
		return binding.Validator.ValidateStruct(obj)
	}
	return nil
}

func BindCookie(c *gin.Context, obj any) error {
	m := make(map[string][]string, len(c.Request.Cookies()))
	for _, cookie := range c.Request.Cookies() {
		if _, ok := m[cookie.Name]; !ok {
			m[cookie.Name] = []string{}
		}
		m[cookie.Name] = append(m[cookie.Name], cookie.Value)
	}
	if err := binding.MapFormWithTag(obj, m, TagCookie); err != nil {
		return err
	}
	if binding.Validator != nil {
		return binding.Validator.ValidateStruct(obj)
	}
	return nil
}
