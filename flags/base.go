package flags

import (
	"context"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
)

const (
	NoFlag = iota
	RemoveTag
	AddTag
	AddTags
	CheckTag
)

type FlagHandler func(ctx context.Context, m *tg.Message, u *database.User, answer *message.RequestBuilder) error
