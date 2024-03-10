package gapi

import (
	"context"
	"fmt"
	"time"

	"github.com/silaselisha/coffee-api/pkg/pb"
	"github.com/silaselisha/coffee-api/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *Server) CreateOrder(ctx context.Context, req *pb.OrderRequest) (*pb.OrderResponse, error) {
	ordersCollection := srv.mongo.Collection(ctx, "coffeeshop", "orders")
	productsCollection := srv.mongo.Collection(ctx, "coffeeshop", "products")

	userId, err := primitive.ObjectIDFromHex(req.GetUser())
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "user forbiden")
	}

	products := req.GetProducts()

	var items []store.ItemOrder
	var totalAmount float64
	var totalDiscount float64

	for _, product := range products {
		fmt.Println(product.GetProductId())
		id, err := primitive.ObjectIDFromHex(product.GetProductId())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid product in the bucket")
		}
		fmt.Println(id)
		var item store.Item
		curr := productsCollection.FindOne(ctx, bson.D{{Key: "_id", Value: id}})
		err = curr.Decode(&item)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, status.Errorf(codes.InvalidArgument, "product not found")
			}
			return nil, status.Errorf(codes.InvalidArgument, "invalid product in the bucket")
		}

		amount := item.Price * float64(product.GetQuantity())
		discount := ((product.GetDiscount() * amount) / 100)
		totalDiscount += discount
		totalAmount += amount - discount

		items = append(items, store.ItemOrder{
			Item:     id,
			Quantity: product.GetQuantity(),
			Discount: product.GetDiscount(),
		})
	}

	order := store.Order{
		Id:            primitive.NewObjectID(),
		User:          userId,
		Items:         items,
		TotalDiscount: totalDiscount,
		TotalAmount:   totalAmount,
		Status:        store.PENDING,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	result, err := ordersCollection.InsertOne(ctx, order)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "order not created")
	}

	var productOrder store.Order
	id := result.InsertedID.(primitive.ObjectID)
	curr := ordersCollection.FindOne(ctx, bson.D{{Key: "_id", Value: id}})
	err = curr.Decode(&productOrder)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	var resultProducts []*pb.Order
	for _, product := range productOrder.Items {
		item := pb.Order{
			ProductId: product.Item.Hex(),
			Quantity:  product.Quantity,
			Discount:  product.Discount,
		}
		resultProducts = append(resultProducts, &item)
	}

	return &pb.OrderResponse{
		Bucket: &pb.Bucket{
			XId:           productOrder.Id.Hex(),
			User:          productOrder.User.Hex(),
			Products:      resultProducts,
			TotalAmount:   productOrder.TotalAmount,
			TotalDiscount: productOrder.TotalDiscount,
			Status:        int32(productOrder.Status),
			CreatedAt:     timestamppb.New(productOrder.CreatedAt),
			UpdatedAt:     timestamppb.New(productOrder.UpdatedAt),
		},
	}, nil
}
