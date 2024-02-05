package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ObjetoRecebido struct {
	Chave1 string `json:"chave1"`
	Chave2 string `json:"chave2"`
}

func main() {
	router := gin.Default()

	// Configurar CORS
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

	// Defina suas rotas aqui
	router.POST("/api/enviar-objeto", receiveData)
	router.GET("/api/receber-dados", sendData)

	// Inicie o servidor
	router.Run(":8080")
}

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

func tratarDados(objeto ObjetoRecebido) interface{} {
	// Exemplo de lógica para processar os dados recebidos
	if objeto.Chave1 == "valor-correto" {
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

func sendData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"mensagem": "Olá do backend!"})
}
