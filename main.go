package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"

	"googlemaps.github.io/maps"
)

const pmConnectAddress = "B31 2UQ"

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

// Response represents a response from the aws-lambda function.
type Response struct {
	Distance string `json:"distance"`
	Ok       bool   `json:"ok"`
}

func main() {
	lambda.Start(Handler)
}

// Handler handles requests to the lambda function
func Handler(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	api := os.Getenv("GOOGLE_API")
	client, err := maps.NewClient(maps.WithAPIKey(api))
	if err != nil {
		return serverError(err)
	}

	origin := r.QueryStringParameters["origin"]

	distance, err := getDistance(client, origin)
	if err != nil {
		clientError(http.StatusNotFound)
	}

	resp := Response{
		Distance: string(distance),
		Ok:       true,
	}

	j, err := json.Marshal(resp)
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(j),
	}, nil
}

func getDistance(client *maps.Client, origin string) ([]byte, error) {
	r := &maps.DistanceMatrixRequest{
		Origins:      []string{origin},
		Destinations: []string{pmConnectAddress},
	}

	resp, err := client.DistanceMatrix(context.Background(), r)
	if err != nil {
		return nil, errors.Wrap(err,
			fmt.Sprintf("could not get response from origin %s", origin))
	}

	// resp always contains information, this will not panic due to
	// being out of range
	distance := resp.Rows[0].Elements[0].Distance.HumanReadable
	if len(distance) == 0 {
		return nil, errors.New("could not find a distance, is your origin correct?")
	}

	return []byte(distance), nil
}

// serverError is a helper function for handling error that the AWS API
// Gateway understands.
func serverError(err error) (events.APIGatewayProxyResponse, error) {
	errorLogger.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

// clientError is a helper for send responses relating to client errors.
func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}
