package soda

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type UIRender interface {
	Render(spec *openapi3.T) string
}

type builtinUIRender struct {
	template string
}

func (u builtinUIRender) Render(spec *openapi3.T) string {
	s, _ := spec.MarshalJSON()
	replacer := strings.NewReplacer(
		"{:title}", spec.Info.Title,
		"{:spec}", string(s),
	)
	return replacer.Replace(u.template)
}

const uiSwaggerUI = `
<!DOCTYPE html>
<html charset="UTF-8">
<head>
    <meta http-equiv="Content-Type" content="text/html;charset=utf-8">
    <title>{:title} Document [Swagger UI]</title>
    <link type="text/css" rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui.css">
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
</head>
</html>
<body>
  <div id="ui"></div>
  <script>
    let spec = {:spec};
    let oauth2RedirectUrl;

    let query = window.location.href.indexOf("?");
    if (query > 0) {
        oauth2RedirectUrl = window.location.href.substring(0, query);
    } else {
        oauth2RedirectUrl = window.location.href;
    }

    if (!oauth2RedirectUrl.endsWith("/")) {
        oauth2RedirectUrl += "/";
    }
    oauth2RedirectUrl += "oauth-receiver.html";
    SwaggerUIBundle({
        dom_id: '#ui',
        spec: spec,
        filter: false,
        oauth2RedirectUrl: oauth2RedirectUrl,
    })
  </script>`

const uiRapiDoc = `
<!DOCTYPE html>
<html charset="UTF-8">
  <head>
    <meta http-equiv="Content-Type" content="text/html;charset=utf-8">
    <meta name="viewport" content="width=device-width, minimum-scale=1, initial-scale=1, user-scalable=yes">
    <title>{:title} Document [RapiDoc]</title>
    <script type="module" src="https://cdn.jsdelivr.net/npm/rapidoc/dist/rapidoc-min.min.js"></script>
  </head>
  <style>
    rapi-doc::part(section-navbar) { /* <<< targets navigation bar */
      background: linear-gradient(90deg, #3d4e70, #2e3746);
    }
  </style>
  <body>
    <rapi-doc id="thedoc" 
    theme="dark" 
    primary-color = "#f54c47"
    bg-color = "#2e3746"
    text-color = "#bacdee"
    default-schema-tab="model" 
    allow-search="false"
    allow-advanced-search="true"
    show-info="true" 
    show-header="true" 
    show-components="true" 
    schema-style="table"
    show-method-in-nav-bar="as-colored-block" 
    allow-try="true"
    allow-authentication="true" 
    regular-font="Open Sans" 
    mono-font="Roboto Mono" 
    font-size="large"
    schema-description-expanded="true">
    </rapi-doc>
    <script>
      document.addEventListener('DOMContentLoaded', (event) => {
        let docEl = document.getElementById("thedoc");
        docEl.loadSpec({:spec});
      })
    </script>
  </body>
</html>`

const uiStoplightElement = `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>{:title} Document [Elements]</title>
  
    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
    <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
  </head>
  <body>
    <elements-api id="doc" router="hash" hideSchemas="true" />
  </body>

  <script>
    (async() => {
      let doc = document.getElementById("doc");
      doc.apiDescriptionDocument = {:spec};
    })()
  </script>
</html>`

const uiRedoc = `
<!DOCTYPE html>
<html>
  <head>
    <title>Redoc</title>
    <!-- needed for adaptive design -->
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">

    <style>
      body {
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <redoc id="doc"></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"> </script>
    <script>
      (async()=>{
        Redoc.init({:spec}, {}, document.getElementById('doc'))
      })()
    </script>
  </body>
</html>`

var (
	UISwaggerUI        = builtinUIRender{template: uiSwaggerUI}
	UIRapiDoc          = builtinUIRender{template: uiRapiDoc}
	UIStoplightElement = builtinUIRender{template: uiStoplightElement}
	UIRedoc            = builtinUIRender{template: uiRedoc}
)
