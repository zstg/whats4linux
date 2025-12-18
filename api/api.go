package api

import (
	"context"
	"fmt"

	"github.com/lugvitc/whats4linux/internal/wa"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.mau.fi/whatsmeow"
)

// Api struct
type Api struct {
	ctx      context.Context
	waClient *whatsmeow.Client
}

// NewApi creates a new Api application struct
func New() *Api {
	return &Api{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *Api) Startup(ctx context.Context) {
	a.ctx = ctx
	a.waClient = wa.NewClient(ctx)
}

func (a *Api) Login() error {
	var err error
	if a.waClient.Store.ID == nil {
		qrChan, _ := a.waClient.GetQRChannel(a.ctx)
		err = a.waClient.Connect()
		if err != nil {
			return err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				runtime.EventsEmit(a.ctx, "wa:qr", evt.Code)
			} else {
				runtime.EventsEmit(a.ctx, "wa:status", evt.Event)
			}
		}
	} else {
		runtime.EventsEmit(a.ctx, "wa:status", "logged_in")
		fmt.Println("Already logged in, connecting...")
		// Already logged in, just connect
		err = a.waClient.Connect()
		if err != nil {
			return err
		}
	}
	return nil
}
