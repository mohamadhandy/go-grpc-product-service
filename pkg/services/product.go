package services

import (
	"context"
	"go-grpc-product-svc/pkg/db"
	"go-grpc-product-svc/pkg/models"
	"go-grpc-product-svc/pkg/pb"
	"net/http"
)

type Server struct {
	pb.ProductServiceServer
	H db.Handler
}

func (s *Server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	product := models.Product{
		Name:  req.Name,
		Stock: req.Stock,
		Price: req.Price,
	}
	if result := s.H.DB.Create(&product); result.Error != nil {
		return &pb.CreateProductResponse{
			Status: http.StatusConflict,
			Error:  result.Error.Error(),
		}, nil
	}
	return &pb.CreateProductResponse{
		Status: http.StatusCreated,
		Id:     product.Id,
	}, nil
}

func (s *Server) FindOne(ctx context.Context, req *pb.FindOneRequest) (*pb.FindOneResponse, error) {
	product := models.Product{}
	if result := s.H.DB.First(&product, req.Id); result.Error != nil {
		return &pb.FindOneResponse{
			Status: http.StatusConflict,
			Error:  result.Error.Error(),
		}, nil
	}

	data := &pb.FindOneData{
		Id:    product.Id,
		Name:  product.Name,
		Stock: product.Stock,
		Price: product.Price,
	}

	return &pb.FindOneResponse{
		Status: http.StatusOK,
		Data:   data,
	}, nil
}

func (s *Server) DecreaseStock(ctx context.Context, req *pb.DecreaseStockRequest) (*pb.DecreaseStockResponse, error) {
	product := models.Product{}
	if result := s.H.DB.First(&product, req.Id); result.Error != nil {
		return &pb.DecreaseStockResponse{
			Status: http.StatusNotFound,
			Error:  result.Error.Error(),
		}, nil
	}

	if product.Stock <= 0 {
		return &pb.DecreaseStockResponse{
			Status: http.StatusConflict,
			Error:  "stock too low",
		}, nil
	}

	logStock := models.StockDecreaseLog{}
	if result := s.H.DB.Where(&models.StockDecreaseLog{OrderId: req.OrderId}).First(&logStock); result.Error == nil {
		return &pb.DecreaseStockResponse{
			Status: http.StatusConflict,
			Error:  "stock already decreased",
		}, nil
	}

	product.Stock -= 1
	s.H.DB.Save(&product)

	logStock.Id = req.OrderId
	logStock.ProductRefer = product.Id

	s.H.DB.Create(&logStock)
	return &pb.DecreaseStockResponse{
		Status: http.StatusOK,
	}, nil
}
