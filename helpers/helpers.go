package helpers

import (
	"Newton/query"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var mySigningKey = []byte("foxtrot")

//GetEnvWithKey : get env value
func GetEnvWithKey(key string) string {
	return os.Getenv(key)
}

//LoadEnv : loading the env file
func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		os.Exit(1)
	}
}

// GenerateJWT ...
func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()
	claims["admin"] = true
	claims["userid"] = "chirantan"

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		fmt.Println("Something went wrong")
	}
	fmt.Println(tokenString)
	return tokenString, err
}

// ProductUploadHandler : handles the product upload
func ProductImageHandler(w http.ResponseWriter, r *http.Request, id primitive.ObjectID, sub string, cat string) {

	w.Header().Set("Content-Type", "application/json")

	var imageName [8]string
	var imageURL [8]string
	key := []string{
		"img1",
		"img2",
		"img3",
		"img4",
		"img5",
		"img6",
		"img7",
		"img8",
	}
	var count int

	for i := 0; i < 8; i++ {

		_, _, err := r.FormFile(key[i])
		if err != nil {
			break
		}
		count++
	}

	LoadEnv()
	awsAccessKeyID := GetEnvWithKey("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := GetEnvWithKey("AWS_SECRET_ACCESS_KEY")

	for i := 0; i < count; i++ {

		file, fileHeader, err := r.FormFile(key[i])
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, "Could not get uploaded file")
			return
		}
		imageName[i] = fileHeader.Filename
		newString := cat + "/" + sub + "/" + imageName[i]

		s, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials(
				awsAccessKeyID,     // id
				awsSecretAccessKey, // secret
				""),                // token can be left blank for now
		})
		if err != nil {
			fmt.Fprintf(w, "Could not upload file")
		}
		uploader := s3manager.NewUploader(s)

		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:             aws.String("ckrht"),
			ACL:                aws.String("public-read"),
			Key:                aws.String(newString),
			Body:               file,
			ContentType:        aws.String("image/jpeg"),
			ContentDisposition: aws.String("inline; filename=" + fmt.Sprintf("%s", imageName[i])),
		})
		if err != nil {
			fmt.Printf("failed to upload file, %v", err)
			return
		}
		imageURL[i] = aws.StringValue(&result.Location)
		file.Close()
	}
	for i := 0; i < len(imageURL); i++ {

		filter := bson.M{"_id": id}
		update := bson.M{"$push": bson.M{"img": imageURL[i]}}

		query.UpdateOne("products", filter, update)
	}

}
