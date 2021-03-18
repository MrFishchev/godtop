package consoleui

import (
	"fmt"
	"godtop/consoleui/config"
	"godtop/consoleui/layout"
	"godtop/consoleui/logging"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	ui "github.com/gizak/termui/v3"
)

func Run() error {
	config := config.NewConfig()

	logfile, err := logging.New(config)
	if err != nil {
		fmt.Println("failed to configure logger", err.Error())
		return err
	}
	defer logfile.Close()

	if err := ui.Init(); err != nil {
		return err
	}

	layoutStream, err := getLayout(config)
	if err != nil {
		log.Fatalf("Cannot parse layout: %v", err)
		return err
	}

	layoutRaw := layout.ParseLayout(layoutStream)

	defer ui.Close()

	setDefaultUiColors(config)

	grid, err := layout.GenerateGrid(layoutRaw, config)
	if err != nil {
		return err
	}

	terminalWidth, terminalHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, terminalWidth, terminalHeight)

	ui.Render(grid)

	eventLoop(config, grid)
	return nil
}

func eventLoop(config config.Config, grid *layout.Grid) {
	drawTicker := time.NewTicker(config.UpdateInterval).C

	//handles kill signal
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	uiEvents := ui.PollEvents()
	previousKey := ""

	for {
		select {
		case <-sigTerm:
			return
		case <-drawTicker:
			ui.Render(grid)
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				terminalW, terminalH := payload.Width, payload.Height
				grid.SetRect(0, 0, terminalW, terminalH)
				ui.Clear()

				if previousKey == e.ID {
					previousKey = ""
				} else {
					previousKey = e.ID
				}
			}
		}
	}
}

func setDefaultUiColors(config config.Config) {
	ui.Theme.Default = ui.NewStyle(ui.Color(config.Colorscheme.MainFg), ui.Color(config.Colorscheme.MainBg))
	ui.Theme.Block.Title = ui.NewStyle(ui.Color(config.Colorscheme.BorderFg), ui.Color(config.Colorscheme.MainBg))
	ui.Theme.Block.Border = ui.NewStyle(ui.Color(config.Colorscheme.BorderFg), ui.Color(config.Colorscheme.MainBg))
}

func getLayout(config config.Config) (io.Reader, error) {
	// TODO: parse config to get layout
	return strings.NewReader("cpu\nvolumes network\n"), nil
}
