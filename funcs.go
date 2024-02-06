package test

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

type ObjetoRecebido struct {
	Chave1 string `json:"chave1"`
	Chave2 string `json:"chave2"`
}

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

	router.POST("/api/enviar-objeto", receiveData)
	router.GET("/api/receber-dados", sendData)
	router.GET("api/savedata", testeFirebase)
	router.GET("api/buscar-dados-firestore", getDataFirestore)

	// Iniciando servidor
	router.Run(":8080")
}

var ctx = context.Background()

// INICIAR FIREBASE
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

// ENVIAR DADOS AO FIREBASE
func testeFirebase(c *gin.Context) {
	client := initFirebase()
	defer client.Close()

	data := map[string]interface{}{
		"campo1": "valor1",
		"campo2": "valor2",
	}

	// Adicionar dados ao Firestore
	_, _, err := client.Collection("nova").Add(ctx, data)
	if err != nil {
		log.Fatalf("Erro ao salvar dados no Firestore: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Erro ao salvar dados"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dados salvos com sucesso"})
}

// RECEBER ALGO DO FRONT
func receiveData(c *gin.Context) {
	var objeto ObjetoRecebido
	if err := c.BindJSON(&objeto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Erro ao decodificar o objeto JSON"})
		return
	}

	// Chame a função para tratar os dados
	resultado := tratarDados(objeto)
	fmt.Printf("Dados recebidos do frontend: %+v\n", objeto)

	// Responda ao cliente
	c.JSON(http.StatusOK, resultado)
}

func tratarDados(product any) interface{} {
	// Exemplo de lógica para processar os dados recebidos
	if product == "valor-correto" {
		return map[string]interface{}{
			"mensagem": "Dados processados com sucesso!",
			"detalhes": "Chave1 possui o valor esperado.",
		}
	} else {
		return map[string]interface{}{
			"mensagem": "Erro ao processar os dados.",
			"detalhes": "Chave1 não possui o valor esperado.",
		}
	}
}

// ENVIAR ALGO AO FRONT
func sendData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"mensagem": "Olá do backend!"})
}

// BUSCAR DADOS NO FIRESTORE
func getDataFirestore(c *gin.Context) {
	client := initFirebase()
	defer client.Close()
	// Buscar dados do Firestore
	iter := client.Collection("nova").Documents(ctx)
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

	// c.JSON(http.StatusOK, gin.H{"message": "Dados salvos com sucesso"})

	fmt.Printf("Dados recebidos do frontend: %+v\n", product)

	// Resposta ao cliente
	c.JSON(http.StatusOK, "Produto salvo na DB")
}

// func saveProductinFB(produto Product) gin.H {
// 	// Aqui você pode adicionar sua lógica de tratamento de dados
// 	// Por exemplo, salvar os dados em um banco de dados, processar os dados, etc.
// 	// Neste exemplo, retornamos apenas uma mensagem de sucesso com os dados recebidos
// 	return gin.H{"mensagem": "Dados recebidos com sucesso", "produto": produto}
// }
