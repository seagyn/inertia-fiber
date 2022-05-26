package inertia

import (
	"context"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// Inertia type.
type Inertia struct {
	url           string
	rootTemplate  string
	version       string
	session       *session.Session
	sharedProps   map[string]interface{}
	sharedFuncMap template.FuncMap
	templateFS    fs.FS
}

// New function.
func New(url, rootTemplate, version string, session *session.Session) *Inertia {
	i := new(Inertia)
	i.url = url
	i.rootTemplate = rootTemplate
	i.version = version
	i.session = session
	i.sharedProps = make(map[string]interface{})
	i.sharedFuncMap = template.FuncMap{"marshal": marshal}

	return i
}

// NewWithFS function.
func NewWithFS(url, rootTemplate, version string, session *session.Session, templateFS fs.FS) *Inertia {
	i := New(url, rootTemplate, version, session)
	i.templateFS = templateFS

	return i
}

// Share function.
func (i *Inertia) Share(key string, value interface{}) {
	i.sharedProps[key] = value
}

// ShareFunc function.
func (i *Inertia) ShareFunc(key string, value interface{}) {
	i.sharedFuncMap[key] = value
}

// WithProp function.
func (i *Inertia) WithProp(ctx context.Context, key string, value interface{}) context.Context {
	contextProps := ctx.Value(ContextKeyProps)

	if contextProps != nil {
		contextProps, ok := contextProps.(map[string]interface{})
		if ok {
			contextProps[key] = value

			return context.WithValue(ctx, ContextKeyProps, contextProps)
		}
	}

	return context.WithValue(ctx, ContextKeyProps, map[string]interface{}{
		key: value,
	})
}

// WithViewData function.
func (i *Inertia) WithViewData(ctx context.Context, key string, value interface{}) context.Context {
	contextViewData := ctx.Value(ContextKeyViewData)

	if contextViewData != nil {
		contextViewData, ok := contextViewData.(map[string]interface{})
		if ok {
			contextViewData[key] = value

			return context.WithValue(ctx, ContextKeyViewData, contextViewData)
		}
	}

	return context.WithValue(ctx, ContextKeyViewData, map[string]interface{}{
		key: value,
	})
}

// Render function.
func (i *Inertia) Render(c *fiber.Ctx, component string, props map[string]interface{}) error {
	only := make(map[string]string)
	partial := c.Get("X-Inertia-Partial-Data")

	if partial != "" && c.Get("X-Inertia-Partial-Component") == component {
		for _, value := range strings.Split(partial, ",") {
			only[value] = value
		}
	}

	page := &Page{
		Component: component,
		Props:     make(map[string]interface{}),
		Routes:    flattenRoutes(c.App().Stack()),
		URL:       c.OriginalURL(),
		Version:   i.version,
	}

	for key, value := range i.sharedProps {
		if _, ok := only[key]; len(only) == 0 || ok {
			page.Props[key] = value
		}
	}

	contextProps := c.Context().Value(ContextKeyProps)

	if contextProps != nil {
		contextProps, ok := contextProps.(map[string]interface{})
		if !ok {
			return ErrInvalidContextProps
		}

		for key, value := range contextProps {
			if _, ok := only[key]; len(only) == 0 || ok {
				page.Props[key] = value
			}
		}
	}

	for key, value := range props {
		if _, ok := only[key]; len(only) == 0 || ok {
			page.Props[key] = value
		}
	}

	page.Props["params"] = c.AllParams()

	if c.Get("X-Inertia") != "" {
		js, err := json.Marshal(page)
		if err != nil {
			return err
		}

		c.Set("Vary", "Accept")
		c.Set("X-Inertia", "true")
		c.Set("Content-Type", "application/json")

		return c.Send(js)
	}

	viewData := make(map[string]interface{})
	contextViewData := c.Context().Value(ContextKeyViewData)

	if contextViewData != nil {
		contextViewData, ok := contextViewData.(map[string]interface{})
		if !ok {
			return ErrInvalidContextViewData
		}

		for key, value := range contextViewData {
			viewData[key] = value
		}
	}

	viewData["page"] = page

	ts, err := i.createRootTemplate()
	if err != nil {
		return err
	}

	c.Set("Content-Type", "text/html")

	err = ts.Execute(c.Response().BodyWriter(), viewData)
	if err != nil {
		return err
	}

	return nil
}

// Location function.
func (i *Inertia) Location(w http.ResponseWriter, location string) {
	w.Header().Set("X-Inertia-Location", location)
	w.WriteHeader(http.StatusConflict)
}

func (i *Inertia) createRootTemplate() (*template.Template, error) {
	ts := template.New(filepath.Base(i.rootTemplate)).Funcs(i.sharedFuncMap)

	if i.templateFS != nil {
		return ts.ParseFS(i.templateFS, i.rootTemplate)
	}

	return ts.ParseFiles(i.rootTemplate)
}

func flattenRoutes(r [][]*fiber.Route) []*fiber.Route {
	routes := []*fiber.Route{}

	for _, routeGroup := range r {
		for _, route := range routeGroup {
			routes = append(routes, route)
		}
	}

	return routes
}
