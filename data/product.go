package data

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const DBNAME = "Products"
const COLLNAME = "products"

type Product struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Quantity    int                `json:"quantity,omitempty" bson:"quantity,omitempty"`
	Price       float32            `json:"price,omitempty" bson:"price,omitempty"`
	Rating      float32            `json:"rating,omitempty" bson:"rating,omitempty"`
	//createdOn
	//updatedOn
	//deletedOn
}

// Products = collection of Product
type Products []*Product

//conversion helper methods
func (p *Product) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(p)
}

/*"ToJSON serializes the contents of the collection to JSON
NewEncoder provides better performance than json.Unmarshal() as it does not
have to buffer the output into an in memory slice of bytes
this reduces allocations and the overheads of the service"*/
func (p *Products) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

//Product data manipulation methods
func GetProducts(db *mongo.Client) (Products, error) {
	//get a list of products from db
	prodCol := db.Database(DBNAME).Collection(COLLNAME)
	cur, err := prodCol.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.TODO())

	var prodList Products
	for cur.Next(context.TODO()) {
		var prod *Product
		err = cur.Decode(&prod)
		prodList = append(prodList, prod)
	}

	if err != nil {
		return nil, err
	}
	return prodList, nil
}

//Gets product with id from db, returns product and error objects
func GetProduct(id primitive.ObjectID, db *mongo.Client) (Product, error) {
	prodCol := db.Database(DBNAME).Collection(COLLNAME)
	filter := bson.M{"_id": id}
	result := prodCol.FindOne(context.TODO(), filter)

	var prod Product
	err := result.Decode(&prod)
	if err != nil {
		return Product{}, err
	}
	return prod, nil
}

/*Adds product to db, returns the ID of the inserted object*/
func AddProduct(p *Product, db *mongo.Client) (string, error) {
	//insert product p into db
	prodCol := db.Database(DBNAME).Collection(COLLNAME)
	result, err := prodCol.InsertOne(context.TODO(), p)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

/*Updates product with id in the database by passing object p to the database. Assumption: object's model has omitempty tags*/
func UpdateProduct(id primitive.ObjectID, p *Product, db *mongo.Client) error {
	//find the product with the same id and update it.
	prodCol := db.Database(DBNAME).Collection(COLLNAME)

	filter := bson.M{"_id": id}
	result, err := prodCol.UpdateOne(context.TODO(), filter, bson.D{{"$set", p}}) //TODO: omitempty on the model
	if err != nil {
		return err
	} else if result.ModifiedCount == 0 {
		msg := "Failed to update product with id = " + id.Hex()
		return errors.New(msg)
	}
	return nil
}

/*Deletes product with id from db, returns the number of records deleted and error object. Assumption: the operation does not error if the id doesn't exist*/
func DeleteProduct(id primitive.ObjectID, db *mongo.Client) (int, error) {
	//delete product with id from db
	prodCol := db.Database(DBNAME).Collection(COLLNAME)

	filter := bson.M{"_id": id}
	result, err := prodCol.DeleteOne(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return int(result.DeletedCount), nil
}
