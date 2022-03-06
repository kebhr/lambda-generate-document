package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kebhr/docgen"
	"os"
)

type Payload struct {
	Title string `json:"title"`
	Date  string `json:"date"`
	Name  string `json:"name"`
	Items []Item `json:"items"`
	Type  string `json:"type"`
}

type Item struct {
	Title     string `json:"title"`
	UnitPrice int    `json:"unit_price"`
	Quantity  uint   `json:"quantity"`
	Unit      string `json:"unit"`
}

func genPdf(payload Payload) ([]byte, error) {
	var doc docgen.Document

	docConf, orgConf, err := docgen.Parse("/opt/.config/config.json")
	if err != nil {
		return []byte{}, err
	}

	var items []docgen.Item
	for _, v := range payload.Items {
		items = append(items, docgen.Item(v))
	}

	var part = docgen.Part{
		Title: payload.Title,
		Date:  payload.Date,
		Name:  payload.Name,
		Items: items,
		Type:  payload.Type,
	}

	docConf.FirstPart = part

	if err := doc.Start(docConf, orgConf); err != nil {
		return []byte{}, err
	}

	var buf bytes.Buffer
	if err := doc.Write(&buf); err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

func HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{}
	if origin := os.Getenv("CORS_ORIGIN"); origin != "" {
		headers["Access-Control-Allow-Origin"] = origin
	}
	var payload Payload
	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       err.Error(),
		}, nil
	}

	pdf, err := genPdf(payload)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       "Internal Server Error",
		}, nil
	}

	headers["Content-Type"] = "application/pdf"
	return events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         headers,
		Body:            base64.StdEncoding.EncodeToString(pdf),
		IsBase64Encoded: true,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
