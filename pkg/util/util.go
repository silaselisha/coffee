package util

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Config struct {
	DBPassword      string `mapstructure:"DB_PASSWORD"`
	DBUri           string `mapstructure:"DB_URI"`
	ServerAddrs     string `mapstructure:"SERVER_ADDRESS"`
	SecretAccessKey string `mapstructure:"SECRET_ACCESS_KEY"`
	JwtExpiresAt    string `mapstructure:"JWT_EXPIRES_AT"`
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

func ResponseHandler(w http.ResponseWriter, message any, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(message)
}

func CreateNewProduct() store.Item {
	product := store.Item{
		Name:        "Caffe Latte",
		Price:       4.50,
		Description: "A cafe latte is a popular coffee drink that consists of espresso and steamed milk, topped with a thin layer of foam. It is perfect for those who enjoy a smooth and creamy coffee with a balanced flavor. At our coffee shop, we use high-quality beans and fresh milk to make our cafe lattes, and we can customize them with different syrups, spices, or whipped cream. ☕",
		Summary:     "A cafe latte is a coffee drink made with espresso and steamed milk, with a thin layer of foam on top. It has a smooth and creamy taste, and can be customized with different flavors. Our coffee shop offers high-quality and fresh cafe lattes for any occasion.🍵",
		Category:    "beverages",
		Ingridients: []string{"Espresso", "Milk", "Falvored syrup"},
	}

	return product
}

func CreateNewUser() store.User {
	user := store.User{
		UserName:    "al3xa",
		Email:       "al3xa@aws.ac.ch",
		Password:    "abstarct&87",
		PhoneNumber: "+1(571)360-6677",
	}
	return user
}

func ImageResizeProcessor(ctx context.Context, file multipart.File) ([]byte, string, error) {
	imageBytes, err := io.ReadAll(file)
	if err != nil {
		if err == io.EOF {
			return nil, "", fmt.Errorf("end of file error: %w", err)
		}
		return nil, "", err
	}
	defer file.Close()

	contentType := strings.Split(http.DetectContentType(imageBytes), "/")[0]
	ext := strings.Split(http.DetectContentType(imageBytes), "/")[1]
	if contentType != "image" {
		fmt.Println(contentType)
		return nil, "", fmt.Errorf("wrong file upload, only images required")
	}

	imageId, err := genS3ObjectNames()
	if err != nil {
		return nil, "", fmt.Errorf("error generating imageid")
	}

	imageName := fmt.Sprintf("%s.%s", imageId, ext)
	return imageBytes, imageName, nil
}

func S3awsImageUpload(ctx context.Context, imageByte []byte, bucket string, objectKey string, resource string) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return "", fmt.Errorf("aws session error %w", err)
	}

	objectKey = fmt.Sprintf("%s/%s", resource, objectKey)
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(imageByte),
	})

	if err != nil {
		log.Print(err)
		return "", err
	}

	avatarURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, objectKey)
	return avatarURL, nil
}

func PasswordEncryption(password []byte) string {
	return fmt.Sprintf("%x", crypto.SHA256.New().Sum(password))
}

func ComparePasswordEncryption(password, comparePassword string) bool {
	hash := fmt.Sprintf("%x", crypto.SHA256.New().Sum([]byte(password)))
	return hash == comparePassword
}

func genS3ObjectNames() (string, error) {
	buff := make([]byte, 16)
	_, err := rand.Read(buff)
	if err != nil {
		return "", fmt.Errorf("error generating random bytes %w", err)
	}
	return fmt.Sprintf("%x", buff), nil
}
