package apis

import (
	"context"
	"net/http"
	"sherry.archive.com/applications/archive/adapters/multimedia"
	"sherry.archive.com/applications/archive/pkg/repository"
	pb "sherry.archive.com/pb/archive"
	messages "sherry.archive.com/pb/messages"
	"sherry.archive.com/shared/constants"
	"sherry.archive.com/shared/logger"
	"sherry.archive.com/shared/topics"
)

type Service struct {
	pb.ArchiveServiceServer
	querier           repository.Querier
	multimediaStorage multimedia.StorageClient
	publisher         topics.Publisher
}

func NewService(querier repository.Querier, multimediaClient multimedia.StorageClient, publisher topics.Publisher) *Service {
	return &Service{
		querier:           querier,
		multimediaStorage: multimediaClient,
		publisher:         publisher,
	}
}

func (s *Service) GetBooks(ctx context.Context, request *pb.GetBookRequest) (*pb.GetBookResponse, error) {
	return &pb.GetBookResponse{
		Code:    int32(http.StatusOK),
		Message: "Success",
		Data:    nil,
	}, nil
}

func (s *Service) UpsertBook(ctx context.Context, request *pb.UpsertBookRequest) (*pb.UpsertBookResponse, error) {
	var imageUrl *string
	if request.Image != nil {
		url, err := s.multimediaStorage.Save(ctx, request.Image.Value)
		if err != nil {
			logger.Errorf("error uploading image: %s", err.Error())
			return nil, err
		}

		imageUrl = &url
	}
	record := enrichBookRecordFromUpsertBookRequest(request, imageUrl)
	err := s.querier.UpsertBook(ctx, record)
	if err != nil {
		return nil, err
	}
	return &pb.UpsertBookResponse{
		Code:    int32(http.StatusOK),
		Message: "Success",
		Data: &pb.UpsertBookResponse_Data{
			Book: convertBookRecordToProto(record),
		},
	}, nil
}

func (s *Service) GetPages(ctx context.Context, request *pb.GetPagesRequest) (*pb.GetPagesResponse, error) {
	return &pb.GetPagesResponse{
		Code:    int32(http.StatusOK),
		Message: "Success",
		Data:    nil,
	}, nil
}

func (s *Service) CreatePages(ctx context.Context, request *pb.CreatePagesRequest) (*pb.CreatePagesResponse, error) {
	message := &messages.Pages{
		BookId: request.BookId,
		Data:   request.Pages,
	}
	if err := s.publisher.Publish(ctx, message, constants.MultimediaCompressionTopic, nil); err != nil {
		return nil, err
	}
	return &pb.CreatePagesResponse{
		Code:    int32(http.StatusOK),
		Message: "Success",
	}, nil
}

func (s *Service) GetAuthors(ctx context.Context, request *pb.GetAuthorsRequest) (*pb.GetAuthorsResponse, error) {
	return &pb.GetAuthorsResponse{
		Code:    int32(http.StatusOK),
		Message: "Success",
		Data:    nil,
	}, nil
}

func (s *Service) GetPublishers(ctx context.Context, request *pb.GetPublishersRequest) (*pb.GetPublishersResponse, error) {
	return &pb.GetPublishersResponse{
		Code:    int32(http.StatusOK),
		Message: "Success",
		Data:    nil,
	}, nil
}
