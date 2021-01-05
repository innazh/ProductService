package handlers

import (
	"BetterProductService/data"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gorilla/mux"
)

//a http.Handler for all /products routes
type Products struct {
	l  *log.Logger
	db *mongo.Client
}

//creates a new http.Handler - Products
func NewProducts(l *log.Logger, db *mongo.Client) *Products {
	return &Products{l, db}
}

//Handles /products GET request - that gets all products from db as []*Product(Products),
// serializes it to JSON & writes it out to the response writer
func (p *Products) GetProducts(w http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle GET Products")

	//get products from the db
	prodList, err := data.GetProducts(p.db)
	if err != nil {
		p.l.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//convert product list to JSON
	w.Header().Set("Content-Type", "application/json")
	err = prodList.ToJSON(w)
	if err != nil {
		p.l.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	return
}

//Handles /products POST request - adds a new product into the database and returns id back to the user
func (p *Products) AddProduct(w http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle POST Product")
	//getProduct object
	prod := r.Context().Value(KeyProduct{}).(data.Product)
	if prod.ID != primitive.NilObjectID {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Product id must be null or empty."))
		return
	}

	insertedId, err := data.AddProduct(&prod, p.db)
	if err != nil {
		p.l.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Product's id:" + insertedId))
	return
}

//PRODUCT

//Handles /products/{id} GET request - gets a product from the database by id
func (p *Products) GetProduct(w http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle GET Product")
	vars := mux.Vars(r)
	id := vars["id"]
	objId, err := strToObjectID(id)
	if err != nil {
		http.Error(w, "Invalid id.", http.StatusBadRequest)
		return
	}

	prod, err := data.GetProduct(objId, p.db)
	if err != nil {
		p.l.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//convert prod to json
	var jprod []byte
	if jprod, err = json.Marshal(prod); err != nil {
		p.l.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jprod)
	return
}

//Handles /products/{id} PUT request - modifies product with Id
func (p *Products) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle PUT Product")
	vars := mux.Vars(r)
	id := vars["id"]
	//get product object from request's context
	prod := r.Context().Value(KeyProduct{}).(data.Product)
	objId, err := strToObjectID(id)
	if err != nil {
		http.Error(w, "Invalid id.", http.StatusBadRequest)
		return
	}

	err = data.UpdateProduct(objId, &prod, p.db)
	if err != nil {
		p.l.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

//Handles /products/{id} DELETE request - deletes a product from the database
//mb returns the number of records deleted
func (p *Products) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle DELETE Product")
	vars := mux.Vars(r)
	id := vars["id"]

	//moving the conversion of str id to primitiveObjID here because it'll allow for a more detailed error response:
	objId, err := strToObjectID(id)
	if err != nil {
		http.Error(w, "Invalid id.", http.StatusBadRequest)
		return
	}

	delNum, err := data.DeleteProduct(objId, p.db)
	if err != nil {
		p.l.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if delNum == 0 {
		http.Error(w, "The requested resource was not found, impossible to delete.", http.StatusNotFound)
		return
	}
	//case: the resource was deleted, everything is well
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Objects deleted: %d", delNum)))
	return
}

//a preffered approach to use for contexts
type KeyProduct struct{}

//Validates the product passed in a request, passes a valid product object with the context of the request and calls next
func (p *Products) MiddlewareProductValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var prod data.Product
		err := prod.FromJSON(r.Body)
		if err != nil {
			http.Error(w, "Product is invalid.", http.StatusBadRequest)
			return
		}
		//pass product with context of the request
		ctx := context.WithValue(r.Context(), KeyProduct{}, prod)
		req := r.WithContext(ctx)

		next.ServeHTTP(w, req)
	})
}

//helper methods
func strToObjectID(id string) (primitive.ObjectID, error) {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		println(err.Error())
		return primitive.ObjectID{}, err
	}
	return objId, nil
}
