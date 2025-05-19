package tracking_events

import (
	"github.com/gogo/protobuf/types"
	"sherry.archive.com/pb/tracking/events"
)

type TrackingEvent interface {
	GetTrackingId() string
	GetUserId() *types.StringValue
	GetTimestamp() *types.Timestamp
	Reset()
	String() string
	ProtoMessage()
}

var Registry = map[string]func() TrackingEvent{
	"sherry.archive.tracking.events.v1.DocumentView": func() TrackingEvent {
		return &events.DocumentView{}
	},
	"sherry.archive.tracking.events.v1.ChapterView": func() TrackingEvent {
		return &events.ChapterView{}
	},
	"sherry.archive.tracking.events.v1.ChapterCompleted": func() TrackingEvent {
		return &events.ChapterCompleted{}
	},
	"sherry.archive.tracking.events.v1.DocumentFavorited": func() TrackingEvent {
		return &events.DocumentFavorited{}
	},
}
