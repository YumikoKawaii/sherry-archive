package apis

import (
	"github.com/gogo/protobuf/types"
	"sherry.archive.com/applications/archive/pkg/repository"
	pb "sherry.archive.com/pb/archive"
	"sherry.archive.com/shared/proto_values"
	"time"
)

func enrichDocumentRecordFromUpsertDocumentRequest(req *pb.UpsertDocumentRequest, imageUrl *string) *repository.Document {
	return &repository.Document{
		Title:           req.Title,
		Description:     proto_values.StringValueToPointer(req.Description),
		ImageUrl:        imageUrl,
		AuthorId:        proto_values.UInt32ValueToPointer(req.AuthorId),
		PublisherId:     proto_values.UInt32ValueToPointer(req.PublisherId),
		CategoryId:      proto_values.UInt32ValueToPointer(req.CategoryId),
		PublicationDate: uint64ValueToTime(req.PublicationDate),
	}
}

func convertDocumentRecordToProto(document *repository.Document) *pb.Document {
	return &pb.Document{
		Id:              document.Id,
		Title:           document.Title,
		Description:     proto_values.StringPointerToValue(document.Description),
		ImageUrl:        proto_values.StringPointerToValue(document.ImageUrl),
		AuthorId:        proto_values.UInt32PointerToValue(document.AuthorId),
		PublisherId:     proto_values.UInt32PointerToValue(document.PublisherId),
		CategoryId:      proto_values.UInt32PointerToValue(document.CategoryId),
		PublicationDate: proto_values.UInt64PointerToValue(timeToUint64Value(document.PublicationDate)),
	}
}

func convertPageRecordToProto(page *repository.Page) *pb.Page {
	return &pb.Page{
		Id:        page.Id,
		ChapterId: page.ChapterId,
		ImageUrl:  page.ImageUrl,
		Index:     page.Index,
	}
}

func convertAuthorRecordToProto(author *repository.Author) *pb.Author {
	return &pb.Author{
		Id:          author.Id,
		Name:        author.Name,
		Description: author.Description,
		ImageUrl:    author.ImageUrl,
	}
}

func convertPublisherRecordToProto(publisher *repository.Publisher) *pb.Publisher {
	return &pb.Publisher{
		Id:          publisher.Id,
		Name:        publisher.Name,
		Description: publisher.Description,
		ImageUrl:    publisher.ImageUrl,
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
