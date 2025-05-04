package apis

import (
	"github.com/gogo/protobuf/types"
	"sherry.archive.com/applications/archive/pkg/repository"
	pb "sherry.archive.com/pb/archive"
	"sherry.archive.com/shared/proto_values"
	"time"
)

func enrichBookRecordFromUpsertBookRequest(req *pb.UpsertBookRequest, imageUrl *string) *repository.Book {
	return &repository.Book{
		Title:           req.Title,
		Description:     proto_values.StringValueToPointer(req.Description),
		ImageUrl:        imageUrl,
		AuthorId:        proto_values.UInt32ValueToPointer(req.AuthorId),
		PublisherId:     proto_values.UInt32ValueToPointer(req.PublisherId),
		CategoryId:      proto_values.UInt32ValueToPointer(req.CategoryId),
		PublicationDate: uint64ValueToTime(req.PublicationDate),
	}
}

func convertBookRecordToProto(book *repository.Book) *pb.Book {
	return &pb.Book{
		Id:              book.ID,
		Title:           book.Title,
		Description:     proto_values.StringPointerToValue(book.Description),
		ImageUrl:        proto_values.StringPointerToValue(book.ImageUrl),
		AuthorId:        proto_values.UInt32PointerToValue(book.AuthorId),
		PublisherId:     proto_values.UInt32PointerToValue(book.PublisherId),
		CategoryId:      proto_values.UInt32PointerToValue(book.CategoryId),
		PublicationDate: proto_values.UInt64PointerToValue(timeToUint64Value(book.PublicationDate)),
	}
}

func uint64ValueToTime(value *types.UInt64Value) *time.Time {
	secs := proto_values.UInt64ValueToPointer(value)
	if secs != nil {
		t := time.Unix(int64(*secs), 0)
		return &t
	}
	return nil
}

func timeToUint64Value(ts *time.Time) *uint64 {
	if ts == nil {
		return nil
	}
	secs := uint64(ts.Unix())
	return &secs
}
