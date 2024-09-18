package soda_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neo-f/soda/v3"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBinding(t *testing.T) {
	engine := soda.New()

	type embedQuery struct {
		EmbedQuery string `query:"eq"`
	}
	type embedCookie struct {
		EmbedCookie string `cookie:"ec"`
	}
	type embedHeader struct {
		EmbedHeader string `header:"eh"`
	}
	type embedPath struct {
		EmbedPath string `path:"ep"`
	}

	type schema struct {
		Query []string `query:"query" json:"query,omitempty"`
		embedQuery
		Cookie []string `cookie:"cookie" json:"cookie,omitempty"`
		embedCookie
		Path string `path:"path" json:"path,omitempty"`
		embedPath
		Header []string `header:"header" json:"header,omitempty"`
		embedHeader
	}

	handler := func(c *gin.Context) {
		in := soda.GetInput[schema](c)
		c.JSON(200, in)
	}
	engine.Get("test", handler).SetInput(&schema{}).OK()
	engine.Get("test/:path", handler).SetInput(&schema{}).OK()
	engine.Get("test/embed/:ep", handler).SetInput(&schema{}).OK()

	Convey("Given a soda engine", t, func() {
		Convey("Bind Query", func() {
			request := httptest.NewRequest("GET", "/test?query=1&query=2&query=3,4&[]query=5&eq=abc", nil)
			request.Header.Add("Content-Type", "application/json")
			response := httptest.NewRecorder()
			engine.App().ServeHTTP(response, request)
			var actual schema
			_ = json.NewDecoder(response.Body).Decode(&actual)
			expected := schema{
				Query:      []string{"1", "2", "3", "4", "5"},
				embedQuery: embedQuery{EmbedQuery: "abc"},
			}
			slices.Sort(actual.Query)
			So(actual, ShouldResemble, expected)
		})

		Convey("Bind Query Failed", func() {
			request := httptest.NewRequest("GET", "/test?query[=1&query=2", nil)
			request.Header.Add("Content-Type", "application/json")
			response := httptest.NewRecorder()
			engine.App().ServeHTTP(response, request)
			So(response.Code, ShouldEqual, 400)
		})

		Convey("Bind Cookie", func() {
			request := httptest.NewRequest("GET", "/test", nil)
			request.AddCookie(&http.Cookie{Name: "cookie", Value: "1"})
			request.AddCookie(&http.Cookie{Name: "cookie", Value: "2"})
			request.AddCookie(&http.Cookie{Name: "cookie", Value: "3,4"})
			request.AddCookie(&http.Cookie{Name: "ec", Value: "5"})
			response := httptest.NewRecorder()
			engine.App().ServeHTTP(response, request)
			body, _ := io.ReadAll(response.Body)
			expect, _ := json.Marshal(schema{
				Cookie:      []string{"1", "2", "3", "4"},
				embedCookie: embedCookie{EmbedCookie: "5"},
			})
			So(string(body), ShouldEqual, string(expect))
		})

		Convey("Bind Header", func() {
			request := httptest.NewRequest("GET", "/test", nil)
			request.Header.Add("header", "1")
			request.Header.Add("header", "2")
			request.Header.Add("header", "3,4")
			request.Header.Add("eh", "5")
			response := httptest.NewRecorder()
			engine.App().ServeHTTP(response, request)
			body, _ := io.ReadAll(response.Body)
			expect, _ := json.Marshal(schema{
				Header:      []string{"1", "2", "3", "4"},
				embedHeader: embedHeader{EmbedHeader: "5"},
			})
			So(string(body), ShouldEqual, string(expect))
		})

		Convey("Bind Path", func() {
			request := httptest.NewRequest("GET", "/test/1", nil)
			response := httptest.NewRecorder()
			engine.App().ServeHTTP(response, request)
			body, _ := io.ReadAll(response.Body)
			expect, _ := json.Marshal(schema{
				Path: "1",
			})
			So(string(body), ShouldEqual, string(expect))
		})

		Convey("Bind Path Embed", func() {
			request := httptest.NewRequest("GET", "/test/embed/1", nil)
			response := httptest.NewRecorder()
			engine.App().ServeHTTP(response, request)
			body, _ := io.ReadAll(response.Body)
			expect, _ := json.Marshal(schema{
				embedPath: embedPath{EmbedPath: "1"},
			})
			So(string(body), ShouldEqual, string(expect))
		})
	})
}
