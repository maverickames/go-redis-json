package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	redis "github.com/redis/go-redis/v9"
)

// Product struct
type Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Key method
func (p Product) Key() (string, error) {
	if p.ID == "" {
		return "", errors.New("ID is empty")
	}
	return "Product:" + p.ID, nil
}

// Customer struct
type Customer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Key method
func (c Customer) Key() (string, error) {
	if c.ID == "" {
		return "", errors.New("ID is empty")
	}
	return "Customer:" + c.ID, nil
}

// StorageTypes interface
type StorageTypes interface {
	Customer | Product
	Key() (string, error)
}

// Model struct
type Model[T StorageTypes] struct {
	Data T
}

var rdb *redis.Client

func main() {
	// Create a new Redis client
	rdb = redis.NewClient(&redis.Options{})

	// Create a new context
	var ctx = context.Background()

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	// Create a new Product & Push it to storage
	productModel := Model[Product]{Data: Product{
		ID:   "1",
		Name: "Product 1",
	}}
	if err := productModel.Push(ctx); err != nil {
		panic(err)
	}

	// Create a new customer & Push it to storage
	customerModel := Model[Customer]{Data: Customer{
		ID:   "1",
		Name: "Customer 1",
	}}
	if err := customerModel.Push(ctx); err != nil {
		panic(err)
	}

	// Create a new customer 2 & Push it to storage
	newCustomer := Customer{
		ID:   "2",
		Name: "Customer 2",
	}
	if err := Push(ctx, newCustomer); err != nil {
		panic(err)
	}

	// Delete customer
	if err := customerModel.Delete(ctx); err != nil {
		panic(err)
	}

	// Get Product
	product, err := Get[Product](ctx, "Product:1")
	if err != nil {
		panic(err)
	}
	if product == nil {
		fmt.Println("Product not found")
		return
	}
	fmt.Println(product)

	// Delete Product
	if err := Delete(ctx, "Product:1"); err != nil {
		panic(err)
	}

	// New Customer 3 & Push it to storage
	if err := NewCustomer("", "Customer 3").Push(ctx); err != nil {
		fmt.Println(err)
	}
}

// NewCustomer function
func NewCustomer[T Customer](id, name string) *Model[Customer] {
	return &Model[Customer]{Data: Customer{
		ID:   id,
		Name: name,
	}}
}

// Push function
func Push[T StorageTypes](ctx context.Context, data T) error {
	key, err := data.Key()
	if err != nil {
		return err
	}
	return rdb.JSONSet(ctx, key, ".", data).Err()
}

// Get function
func Get[T StorageTypes](ctx context.Context, key string) (*T, error) {
	new := &T{}
	res, err := rdb.JSONGet(ctx, key, ".").Result()
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(res), &new); err != nil {
		return nil, err
	}
	return new, err
}

// Delete function
func Delete(ctx context.Context, key string) error {
	if key != "" {
		return rdb.Del(ctx, key).Err()
	}
	return rdb.Del(ctx, key).Err()
}

// Push method
func (m *Model[T]) Push(ctx context.Context) error {
	key, err := m.Data.Key()
	if err != nil {
		return err
	}
	return rdb.JSONSet(ctx, key, ".", m.Data).Err()
}

// Get method
// This is less usful than the Get function because m is nothing so the key is not known
// This may be useful if you have a key in the struct. Not sure I like it.
func (m *Model[T]) Get(ctx context.Context) (*T, error) {
	new := &T{}
	key, err := m.Data.Key()
	if err != nil {
		return nil, err
	}
	res, err := rdb.JSONGet(ctx, key, ".").Result()
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(res), new); err != nil {
		return nil, err
	}
	return new, nil
}

// Delete method
func (m *Model[T]) Delete(ctx context.Context) error {
	key, err := m.Data.Key()
	if err != nil {
		return err
	}
	return rdb.Del(ctx, key).Err()
}
