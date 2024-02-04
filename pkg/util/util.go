package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Config struct {
	DBPassword  string `mapstructure:"DB_PASSWORD"`
	DBUri       string `mapstructure:"DB_URI"`
	ServerAddrs string `mapstructure:"SERVER_ADDRESS"`
}

func LoadEnvs(path string) (config *Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return
}

func HandleFuncDecorator(handle func(ctx context.Context, w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handle(context.Background(), w, r)
	}
}

func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	return client, nil
}

func ResponseHandler(w http.ResponseWriter, message any, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(message)
}

func CreateNewProduct() store.Item {
	product := store.Item{
		Name:        "Caffe Latte",
		Price:       4.50,
		Description: "A cafe latte is a popular coffee drink that consists of espresso and steamed milk, topped with a thin layer of foam. It is perfect for those who enjoy a smooth and creamy coffee with a balanced flavor. At our coffee shop, we use high-quality beans and fresh milk to make our cafe lattes, and we can customize them with different syrups, spices, or whipped cream. ☕",
		Summary:     "A cafe latte is a coffee drink made with espresso and steamed milk, with a thin layer of foam on top. It has a smooth and creamy taste, and can be customized with different flavors. Our coffee shop offers high-quality and fresh cafe lattes for any occasion.🍵",
		Images:      []string{"caffelatte.jpeg", "lattecafe.jpeg"},
		Thumbnail:   "thumbnail.jpeg",
		Category:    "beverages",
		Ingridients: []string{"Espresso", "Milk", "Falvored syrup"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return product
}

func ImageProcessor(key string, r *http.Request) (fileName string, err error) {
	file, _, err := r.FormFile(key)
	if err != nil {
		log.Print(err)
		return "", err
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	mimeType := strings.Split(http.DetectContentType(imageBytes), "/")[1]
	fileName = fmt.Sprintf("%s.%s", uuid.New().String(), mimeType)
	err = os.WriteFile(fileName, imageBytes, 0666)
	if err != nil {
		return "", err
	}

	return
}
