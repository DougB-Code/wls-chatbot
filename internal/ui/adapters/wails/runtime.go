// runtime.go launches the Wails application with bridge bindings.
// internal/ui/adapters/wails/runtime.go
package wails

import (
	"fmt"
	"io/fs"

	wailslogger "github.com/MadeByDoug/wls-chatbot/internal/ui/adapters/wails/logger"
	"github.com/rs/zerolog"
	wailsruntime "github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// Run launches the Wails UI with configured lifecycle hooks and bindings.
func Run(log zerolog.Logger, assets fs.FS, bridgeService *Bridge, logBridge *wailslogger.Logger) error {

	if bridgeService == nil {
		return fmt.Errorf("wails bridge not configured")
	}
	if logBridge == nil {
		return fmt.Errorf("wails log bridge not configured")
	}

	return wailsruntime.Run(&options.App{
		Title:  "Wails Lit Starter ChatBot",
		Width:  1024,
		Height: 768,
		Logger: wailslogger.NewWailsLogger(log),
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 255},
		OnStartup:        bridgeService.Startup,
		OnShutdown:       bridgeService.Shutdown,
		Bind: []interface{}{
			bridgeService,
			logBridge,
		},
	})
}
