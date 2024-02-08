package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Product struct {
	Ref         string `json:"ref"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Sale        string `json:"sale"`
	Supplier    string `json:"supplier"`
	Cost        string `json:"cost"`
}

func main() {
	router := gin.Default()

	// Configuração CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// Rotas
	router.POST("api/add-product", newProduct)
	router.GET("api/get-products-list", getProducts)

	// Iniciando servidor
	router.Run(":8080")
}

var ctx = context.Background()

func initFirebase() *firestore.Client {
	opt := option.WithCredentialsFile("../stock-control-1a0ab-firebase-adminsdk-hkpew-f737899bec.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error initializing Firestore client: %v", err)
	}

	return client
}

func newProduct(c *gin.Context) {
	client := initFirebase()
	defer client.Close()

	var product Product

	if err := c.BindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Erro ao decodificar o objeto JSON"})
		return
	}

	_, _, err := client.Collection("Products").Add(ctx, product)
	if err != nil {
		log.Fatalf("Erro ao salvar dados no Firestore: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Erro ao salvar dados"})
		return
	}

	fmt.Printf("Dados recebidos do frontend: %+v\n", product)

	// Resposta ao cliente
	c.JSON(http.StatusOK, "Produto salvo na DB")
}

func getProducts(c *gin.Context) {
	client := initFirebase()
	defer client.Close()
	// Buscar dados do Firestore
	iter := client.Collection("Products").Documents(ctx)
	defer iter.Stop()

	var result []map[string]interface{}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Erro ao iterar documentos: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Erro ao buscar dados"})
			return
		}

		// Mapear dados do documento para um mapa
		data := doc.Data()
		result = append(result, data)

		fmt.Printf("Dados do Firestore: %v\n", data)
	}

	c.JSON(http.StatusOK, result)
}
