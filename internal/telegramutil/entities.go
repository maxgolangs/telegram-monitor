package telegramutil

import (
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
)

func BuildEntities(users []tg.UserClass, chats []tg.ChatClass) tg.Entities {
	e := tg.Entities{
		Users:    make(map[int64]*tg.User, len(users)),
		Chats:    make(map[int64]*tg.Chat, len(chats)),
		Channels: make(map[int64]*tg.Channel, len(chats)),
	}

	for _, u := range users {
		if uu, ok := u.(*tg.User); ok && uu != nil {
			e.Users[uu.ID] = uu
		}
	}
	for _, c := range chats {
		switch v := c.(type) {
		case *tg.Chat:
			e.Chats[v.ID] = v
		case *tg.Channel:
			e.Channels[v.ID] = v
		}
	}
	return e
}

func PeerTitle(peer tg.PeerClass, e tg.Entities) string {
	switch p := peer.(type) {
	case *tg.PeerUser:
		if u, ok := e.Users[p.UserID]; ok && u != nil {
			if u.Username != "" {
				return "@" + u.Username
			}
			name := strings.TrimSpace(strings.TrimSpace(u.FirstName + " " + u.LastName))
			if name != "" {
				return name
			}
		}
		return fmt.Sprintf("User %d", p.UserID)
	case *tg.PeerChat:
		if c, ok := e.Chats[p.ChatID]; ok && c != nil {
			return c.Title
		}
		return fmt.Sprintf("Chat %d", p.ChatID)
	case *tg.PeerChannel:
		if c, ok := e.Channels[p.ChannelID]; ok && c != nil {
			return c.Title
		}
		return fmt.Sprintf("Channel %d", p.ChannelID)
	default:
		return "Unknown"
	}
}

func PeerKey(peer tg.PeerClass) string {
	switch p := peer.(type) {
	case *tg.PeerUser:
		return fmt.Sprintf("u:%d", p.UserID)
	case *tg.PeerChat:
		return fmt.Sprintf("c:%d", p.ChatID)
	case *tg.PeerChannel:
		return fmt.Sprintf("ch:%d", p.ChannelID)
	default:
		return "unknown"
	}
}

type SenderInfo struct {
	ID       int64
	Username string
	Name     string
}

func Sender(from tg.PeerClass, e tg.Entities) SenderInfo {
	switch p := from.(type) {
	case *tg.PeerUser:
		if u, ok := e.Users[p.UserID]; ok && u != nil {
			name := strings.TrimSpace(strings.TrimSpace(u.FirstName + " " + u.LastName))
			return SenderInfo{ID: u.ID, Username: u.Username, Name: name}
		}
		return SenderInfo{ID: p.UserID}
	case *tg.PeerChannel:
		if c, ok := e.Channels[p.ChannelID]; ok && c != nil {
			return SenderInfo{ID: c.ID, Username: c.Username, Name: c.Title}
		}
		return SenderInfo{ID: p.ChannelID}
	default:
		return SenderInfo{}
	}
}

func MessageLink(peer tg.PeerClass, msgID int, e tg.Entities) string {
	switch p := peer.(type) {
	case *tg.PeerChannel:
		if ch, ok := e.Channels[p.ChannelID]; ok && ch != nil {
			if ch.Username != "" {
				return fmt.Sprintf("https://t.me/%s/%d", ch.Username, msgID)
			}
			return fmt.Sprintf("https://t.me/c/%d/%d", ch.ID, msgID)
		}
		return fmt.Sprintf("https://t.me/c/%d/%d", p.ChannelID, msgID)
	default:
		return ""
	}
}


