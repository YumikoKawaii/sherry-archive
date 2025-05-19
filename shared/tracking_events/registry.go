package tracking_events

import (
	"google.golang.org/protobuf/proto"
	"sherry.archive.com/pb/tracking/events"
)

var Registry = map[string]func() proto.Message{
	"sherry.archive.tracking.events.v1.DocumentView": func() proto.Message {
		return &events.DocumentView{}
	},
	"sherry.archive.tracking.events.v1.ChapterView": func() proto.Message {
		return &events.ChapterView{}
	},
	"sherry.archive.tracking.events.v1.ChapterCompleted": func() proto.Message {
		return &events.ChapterCompleted{}
	},
	"sherry.archive.tracking.events.v1.DocumentFavorited": func() proto.Message {
		return &events.DocumentFavorited{}
	},
}
