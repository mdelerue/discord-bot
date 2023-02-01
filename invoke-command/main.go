package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
)

// Bot parameters

var GuildID = "XXXXXXXXXX"
var BotToken = "BOT_TOKEN"
var PUBLIC_KEY = "XXXXXXXX"

var session *discordgo.Session

type CommandOptions struct {
	Name  string `json:"name"`
	Type  int    `json:"type"`
	Value string `json:"value"`
}

type RequestBodyData struct {
	Name   string           `json:"name"`
	Option []CommandOptions `json:"options"`
}

type RequestBody struct {
	Type int             `json:"type"`
	Data RequestBodyData `json:"data"`
}

type ResponseData struct {
	Content string `json:"content"`
}
type ResponseBody struct {
	Type int          `json:"type"`
	Data ResponseData `json:"data,omitempty"`
}

func init() {
	var err error
	session, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func formatValid(valid bool) events.APIGatewayProxyResponse {
	if valid {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusForbidden,
	}
}

func handler(request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {

	var signature = request.Headers["x-signature-ed25519"]
	var timestamp = request.Headers["x-signature-timestamp"]

	pubkey_b, _ := hex.DecodeString(PUBLIC_KEY)
	decodedSignature, err := hex.DecodeString(signature)

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	SignedData := []byte(timestamp + request.Body)
	IsSignatureValid := ed25519.Verify(ed25519.PublicKey(pubkey_b), SignedData, decodedSignature)

	if !IsSignatureValid {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	var body = RequestBody{}
	err = json.Unmarshal([]byte(request.Body), &body)

	fmt.Print(request.Body)

	if err != nil {
		fmt.Print(err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusForbidden,
		}, nil
	}

	if body.Type == 1 {

		var responseBody, marshalError = json.Marshal(ResponseBody{Type: body.Type})

		if marshalError != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusForbidden,
			}, nil
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(responseBody),
		}, nil
	}

	fmt.Println(body)
	responseBody := ResponseBody{
		Type: 4,
		Data: ResponseData{
			Content: fmt.Sprint(body.Data.Option[0].Type) + " - " + body.Data.Option[0].Name + " - " + body.Data.Option[0].Value,
		},
	}

	jsonResponse, _ := json.Marshal(responseBody)

	fmt.Println(string(jsonResponse))
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(jsonResponse),
	}, nil

}

func main() {
	lambda.Start(handler)
}
