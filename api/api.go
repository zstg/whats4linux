package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/lugvitc/whats4linux/internal/mstore"
	"github.com/lugvitc/whats4linux/internal/wa"
	"github.com/nyaruka/phonenumbers"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Contact struct {
	JID        string `json:"jid"`
	Short      string `json:"short"`
	FullName   string `json:"full_name"`
	PushName   string `json:"push_name"`
	IsBusiness bool   `json:"is_business"`
}

type ChatElement struct {
	LatestMessage string `json:"latest_message"`
	Contact
}

// Api struct
type Api struct {
	ctx          context.Context
	cw           *wa.ContainerWrapper
	waClient     *whatsmeow.Client
	messageStore *mstore.MessageStore
}

// NewApi creates a new Api application struct
func New() *Api {
	return &Api{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *Api) Startup(ctx context.Context) {
	a.ctx = ctx
	dbLog := waLog.Stdout("Database", "ERROR", true)
	var err error
	a.cw, err = wa.NewContainerWrapper(ctx, "sqlite3", "file:wa.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	a.waClient = wa.NewClient(ctx, a.cw.GetContainer())
	a.messageStore = mstore.NewMessageStore()
}

func (a *Api) Login() error {
	var err error
	a.waClient.AddEventHandler(a.mainEventHandler)
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
		// load only once
		// TODO: add a global flag system for such things
		// if the initialised is 1 => don't load again else do this
		a.cw.Initialise(a.ctx, a.waClient)
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

func (a *Api) FetchGroups() ([]wa.Group, error) {
	groups, err := a.waClient.GetJoinedGroups(a.ctx)
	if err != nil {
		return nil, err
	}

	var result []wa.Group
	for _, g := range groups {
		result = append(result, wa.Group{
			JID:              g.JID.String(),
			Name:             g.Name,
			Topic:            g.Topic,
			OwnerJID:         g.OwnerJID.String(),
			ParticipantCount: len(g.Participants),
		})
	}
	return result, nil
}

func (a *Api) FetchContacts() ([]Contact, error) {
	rawContacts, err := a.waClient.Store.Contacts.GetAllContacts(a.ctx)
	if err != nil {
		return nil, err
	}
	contacts := make([]Contact, 0, len(rawContacts))

	var result []Contact
	for jid, c := range rawContacts {
		rawNum := "+" + jid.User
		// Parse phone number to use as International Format
		num, err := phonenumbers.Parse(rawNum, "")
		if err != nil && !phonenumbers.IsValidNumber(num) {
			continue
		}

		contacts = append(contacts, Contact{
			JID:        phonenumbers.Format(num, phonenumbers.INTERNATIONAL),
			FullName:   c.FullName,
			Short:      c.FirstName,
			PushName:   c.PushName,
			IsBusiness: c.BusinessName != "",
		})
	}
	return result, nil
}

func (a *Api) FetchMessages(jid string) ([]mstore.Message, error) {
	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return nil, err
	}
	messages := a.messageStore.GetMessages(parsedJID)

	return messages, nil
}

func (a *Api) DownloadMedia(chatJID string, messageID string) (string, error) {
	fmt.Printf("DownloadMedia called with chatJID: %s, messageID: %s\n", chatJID, messageID)
	parsedJID, err := types.ParseJID(chatJID)
	if err != nil {
		fmt.Printf("Error parsing JID: %v\n", err)
		return "", err
	}
	msg := a.messageStore.GetMessage(parsedJID, messageID)
	if msg == nil {
		fmt.Printf("Message not found in store for JID: %s, MsgID: %s\n", parsedJID, messageID)
		return "", fmt.Errorf("message not found")
	}

	var data []byte
	var downloadErr error

	if msg.Content.ImageMessage != nil {
		data, downloadErr = a.waClient.Download(a.ctx, msg.Content.ImageMessage)
	} else if msg.Content.VideoMessage != nil {
		data, downloadErr = a.waClient.Download(a.ctx, msg.Content.VideoMessage)
	} else if msg.Content.AudioMessage != nil {
		data, downloadErr = a.waClient.Download(a.ctx, msg.Content.AudioMessage)
	} else if msg.Content.DocumentMessage != nil {
		data, downloadErr = a.waClient.Download(a.ctx, msg.Content.DocumentMessage)
	} else if msg.Content.StickerMessage != nil {
		data, downloadErr = a.waClient.Download(a.ctx, msg.Content.StickerMessage)
	} else {
		return "", fmt.Errorf("no media content found")
	}

	if downloadErr != nil {
		return "", downloadErr
	}

	// Save to Downloads if image, video, document or audio
	if msg.Content.ImageMessage != nil || msg.Content.VideoMessage != nil || msg.Content.DocumentMessage != nil || msg.Content.AudioMessage != nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		downloadsDir := filepath.Join(homeDir, "Downloads")

		var fileName string
		if msg.Content.ImageMessage != nil {
			fileName = messageID + ".jpg" // Default to jpg
		} else if msg.Content.VideoMessage != nil {
			fileName = messageID + ".mp4" // Default to mp4
		} else if msg.Content.AudioMessage != nil {
			fileName = messageID + ".ogg" // Default to ogg
		} else if msg.Content.DocumentMessage != nil {
			if msg.Content.DocumentMessage.FileName != nil && *msg.Content.DocumentMessage.FileName != "" {
				fileName = *msg.Content.DocumentMessage.FileName
			} else {
				fileName = messageID + ".bin"
			}
		}

		filePath := filepath.Join(downloadsDir, fileName)

		// Check if file exists
		if _, err := os.Stat(filePath); err == nil {
			// File exists, prompt user to save as
			newPath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
				DefaultDirectory: downloadsDir,
				DefaultFilename:  fileName,
				Title:           "File already exists. Save as...",
				Filters: []runtime.FileFilter{
					{
						DisplayName: "Media Files",
						Pattern:     "*" + filepath.Ext(fileName),
					},
				},
			})
			if err != nil {
				return "", err
			}
			if newPath == "" {
				return "", fmt.Errorf("download cancelled")
			}
			filePath = newPath
			fileName = filepath.Base(filePath)
		}

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return "", fmt.Errorf("failed to save file: %v", err)
		}

		// Try sending a Linux desktop notification with notify-send (best-effort)
		if _, lookErr := exec.LookPath("notify-send"); lookErr == nil {
			// include full path so the user knows where it was saved
			_ = exec.Command("notify-send", "whats4linux", fmt.Sprintf("Downloaded: %s", filePath)).Run()
		}

		// Emit event for frontend listeners as well
		runtime.EventsEmit(a.ctx, "download:complete", fileName)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func (a *Api) GetChatList() ([]ChatElement, error) {
	cmList := a.messageStore.GetChatList()
	ce := make([]ChatElement, len(cmList))
	for i, cm := range cmList {
		var fc Contact
		if cm.JID.Server == types.GroupServer {
			groupInfo, err := a.cw.FetchGroup(cm.JID.String())
			if err != nil {
				return nil, err
			}
			fc = Contact{
				JID:      cm.JID.String(),
				FullName: groupInfo.Name,
			}
		} else {
			contact, err := a.waClient.Store.Contacts.GetContact(a.ctx, cm.JID)
			if err != nil {
				return nil, err
			}
			fc = Contact{
				JID:        cm.JID.String(),
				Short:      contact.FirstName,
				FullName:   contact.FullName,
				PushName:   contact.PushName,
				IsBusiness: contact.BusinessName != "",
			}
		}
		ce[i] = ChatElement{
			LatestMessage: cm.MessageText,
			Contact:       fc,
		}
	}
	return ce, nil
}

func (a *Api) SendMessage(chatJID string, message string) error {
	if a.waClient.Store.ID == nil {
		return fmt.Errorf("client not logged in")
	}

	parsedJID, err := types.ParseJID(chatJID)
	if err != nil {
		return err
	}

	msgContent := &waE2E.Message{
		Conversation: &message,
	}

	resp, err := a.waClient.SendMessage(a.ctx, parsedJID, msgContent)
	if err != nil {
		return err
	}

	// Manually add to store and emit event so UI updates immediately
	a.messageStore.ProcessMessageEvent(&events.Message{
		Info: types.MessageInfo{
			ID:        resp.ID,
			Timestamp: resp.Timestamp,
			MessageSource: types.MessageSource{
				Chat:     parsedJID,
				IsFromMe: true,
				Sender:   *a.waClient.Store.ID,
			},
		},
		Message: msgContent,
	})
	runtime.EventsEmit(a.ctx, "wa:new_message")

	return nil
}

func (a *Api) mainEventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		a.messageStore.ProcessMessageEvent(v)
		runtime.EventsEmit(a.ctx, "wa:new_message")
	default:
		// Ignore other events for now
	}
}
