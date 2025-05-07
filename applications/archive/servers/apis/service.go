package apis

import (
	"context"
	"net/http"
	"sherry.archive.com/applications/archive/adapters/multimedia"
	"sherry.archive.com/applications/archive/pkg/repository"
	pb "sherry.archive.com/pb/archive"
	"sherry.archive.com/pb/messages"
	"sherry.archive.com/shared/constants"
	"sherry.archive.com/shared/logger"
	"sherry.archive.com/shared/proto_values"
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
		Code:    uint32(http.StatusOK),
		Message: "Success",
		Data:    nil,
	}, nil
}

func (s *Service) UpsertBook(ctx context.Context, request *pb.UpsertBookRequest) (*pb.UpsertBookResponse, error) {
	imageUrl := &request.ImageUrl.Value
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
		Code:    uint32(http.StatusOK),
		Message: "Success",
		Data: &pb.UpsertBookResponse_Data{
			Book: convertBookRecordToProto(record),
		},
	}, nil
}

func (s *Service) GetPages(ctx context.Context, request *pb.GetPagesRequest) (*pb.GetPagesResponse, error) {
	pages, err := s.querier.GetPages(ctx, &repository.GetPagesFilter{
		IDs:    request.Ids,
		BookId: proto_values.UInt32ValueToPointer(request.BookId),
		Pagination: &repository.Pagination{
			Page:     request.Page,
			PageSize: request.PageSize,
		},
	})
	if err != nil {
		logger.Errorf("error fetching pages: %s", err.Error())
		return nil, err
	}
	pagesProto := make([]*pb.Page, 0)
	for _, page := range pages {
		pagesProto = append(pagesProto, convertPageRecordToProto(&page))
	}

	return &pb.GetPagesResponse{
		Code:    uint32(http.StatusOK),
		Message: "Success",
		Data: &pb.GetPagesResponse_Data{
			Pages: pagesProto,
			Pagination: &pb.Pagination{
				Page:     request.Page,
				PageSize: request.PageSize,
			},
		},
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
		Code:    uint32(http.StatusOK),
		Message: "Success",
	}, nil
}

func (s *Service) UpdatePage(ctx context.Context, request *pb.UpdatePageRequest) (*pb.UpdatePageResponse, error) {
	imageUrl := request.ImageUrl
	if request.Image != nil {
		url, err := s.multimediaStorage.Save(ctx, request.Image.Value)
		if err != nil {
			logger.Errorf("error uploading image: %s", err.Error())
			return nil, err
		}

		imageUrl = url
	}

	// TODO: check permission with book
	page := &repository.Page{
		ID:       request.Id,
		BookID:   request.BookId,
		ImageUrl: imageUrl,
		Index:    request.Index,
	}
	if err := s.querier.UpsertPage(ctx, page); err != nil {
		return nil, err
	}

	return &pb.UpdatePageResponse{
		Code:    uint32(http.StatusOK),
		Message: "Success",
		Data: &pb.UpdatePageResponse_Data{
			Page: convertPageRecordToProto(page),
		},
	}, nil
}

func (s *Service) GetAuthors(ctx context.Context, request *pb.GetAuthorsRequest) (*pb.GetAuthorsResponse, error) {
	authors, err := s.querier.GetAuthors(ctx, &repository.GetAuthorsFilter{
		IDs: request.AuthorIds,
		Pagination: &repository.Pagination{
			Page:     request.Page,
			PageSize: request.PageSize,
		},
	})
	if err != nil {
		return nil, err
	}
	authorsProto := make([]*pb.Author, 0)
	for _, author := range authors {
		authorsProto = append(authorsProto, convertAuthorRecordToProto(&author))
	}

	return &pb.GetAuthorsResponse{
		Code:    uint32(http.StatusOK),
		Message: "Success",
		Data: &pb.GetAuthorsResponse_Data{
			Authors: authorsProto,
			Pagination: &pb.Pagination{
				Page:     request.Page,
				PageSize: request.PageSize,
			},
		},
	}, nil
}

func (s *Service) GetPublishers(ctx context.Context, request *pb.GetPublishersRequest) (*pb.GetPublishersResponse, error) {
	publishers, err := s.querier.GetPublishers(ctx, &repository.GetPublishersFilter{
		IDs: request.PublisherIds,
		Pagination: &repository.Pagination{
			Page:     request.Page,
			PageSize: request.PageSize,
		},
	})
	if err != nil {
		return nil, err
	}

	publishersProto := make([]*pb.Publisher, 0)
	for _, publisher := range publishers {
		publishersProto = append(publishersProto, convertPublisherRecordToProto(&publisher))
	}

	return &pb.GetPublishersResponse{
		Code:    uint32(http.StatusOK),
		Message: "Success",
		Data: &pb.GetPublishersResponse_Data{
			Publishers: publishersProto,
			Pagination: &pb.Pagination{
				Page:     request.Page,
				PageSize: request.PageSize,
			},
		},
	}, nil
}
