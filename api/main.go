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
	router.GET("api/get-sales-list", getSales)
	router.POST("api/delete-product", deleteProduct)

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
		data["ID"] = doc.Ref.ID
		result = append(result, data)

		fmt.Printf("Dados do Firestore: %v\n", data)
	}

	c.JSON(http.StatusOK, result)
}

func deleteProduct(c *gin.Context) {
	client := initFirebase()
	defer client.Close()

	var ID string

	if err := c.BindJSON(&ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Erro ao decodificar o objeto JSON"})
		return
	}

	// Obter referência para o documento na coleção "Products"
	docRef := client.Collection("Products").Doc(ID)

	// Obter os dados do documento
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao obter o documento"})
		return
	}

	// Salvar os dados na coleção "Sales"
	if _, err := client.Collection("Sales").Doc(ID).Set(ctx, docSnapshot.Data()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao mover o documento para a coleção Sales"})
		return
	}

	// Remover o documento da coleção "Products"
	if _, err := docRef.Delete(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao remover o documento da coleção Products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Produto movido com sucesso para a coleção Sales"})
}

func getSales(c *gin.Context) {
	client := initFirebase()
	defer client.Close()
	// Buscar dados do Firestore
	iter := client.Collection("Sales").Documents(ctx)
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
		data["ID"] = doc.Ref.ID
		result = append(result, data)

		fmt.Printf("Dados do Firestore: %v\n", data)
	}

	c.JSON(http.StatusOK, result)
}
