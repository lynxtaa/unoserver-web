// Package application provides app
package application

import (
	"github.com/lynxtaa/unoserver-web/internal/application/handler"
	"github.com/lynxtaa/unoserver-web/internal/converter"
)

// App contains application
type App struct {
	ConvertFile *handler.ConvertFileHandler
}

// New creates new application
func New(converter converter.Client) *App {
	return &App{
		ConvertFile: handler.NewConvertFileHandler(converter),
	}
}
