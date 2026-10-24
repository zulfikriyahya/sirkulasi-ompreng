package main

import (
	"context"
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/options/linux"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"mbg/internal/db"
	"mbg/internal/reader"
	"mbg/internal/service"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	dir := db.DataDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("gagal membuat folder data %s: %v", dir, err)
	}

	conn, err := db.Open(filepath.Join(dir, "mbg.db"))
	if err != nil {
		log.Fatalf("gagal membuka database: %v", err)
	}
	defer conn.Close()

	svc := service.New(conn, dir)

	var appCtx context.Context
	mgr := reader.NewManager(
		func() string { p, _ := svc.GetPengaturan(); return p.PortSerial },
		func(uid string) { wruntime.EventsEmit(appCtx, "card:tap", uid) },
		func(st reader.Status) {
			svc.SetStatusReader(st)
			if appCtx != nil {
				wruntime.EventsEmit(appCtx, "reader:status", st)
			}
		},
	)

	err = wails.Run(&options.App{
		Title:       "MBG - Makanan Bergizi Gratis",
		Width:       1200,
		Height:      800,
		AssetServer: &assetserver.Options{Assets: assets},
		Linux:       &linux.Options{Icon: icon, ProgramName: "mbg"},
		OnStartup: func(ctx context.Context) {
			appCtx = ctx
			svc.SetContext(ctx)
			go mgr.Run(ctx)
		},
		Bind: []interface{}{svc},
	})
	if err != nil {
		log.Fatalf("wails: %v", err)
	}
}
