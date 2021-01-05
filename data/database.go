package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*Connects to a remote db instance, returns a db client instance*/
//mongodb+srv://huh:<password>@cluster0.enzda.mongodb.net/<dbname>?retryWrites=true&w=majority
func GetNewClient(connStr string) *mongo.Client {
	var err error
	DbConn, err := mongo.NewClient(options.Client().ApplyURI(connStr))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = DbConn.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return DbConn
}

// func GetMongoDbCollection(dbName, collName string) *mongo.Collection {
// 	return DbConn.Database(dbName).Collection(collName)
// }

// func CloseDbConn() {
// 	DbConn.Disconnect(context.TODO())
// }
