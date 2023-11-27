package main

import (
	"context"
	"fmt"
	"google.golang.org/api/idtoken"
	"net"
	"net/http"
	"os"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	auth_pb "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc"
)

type AuthServer struct{}

type IAPAuth struct {
	ServiceURL string
	Audience   string
	Token      string
	Created    time.Time
}

var iapAuth = IAPAuth{}

func init() {

	serviceUrl := os.Getenv("IAP_SERVICE_URL")
	if serviceUrl == "" {
		fmt.Println("ERROR: IAP_SERVICE_URL not found. Set environment variable containing the IAP service URL")
	}
	iapAuth.ServiceURL = serviceUrl

	audience := os.Getenv("IAP_SERVICE_AUDIENCE")
	if audience == "" {
		fmt.Println("ERROR: IAP_SERVICE_AUDIENCE not found. Set environment variable containing the IAP service URL")
	}
	iapAuth.Audience = audience
}

func (server *AuthServer) Check(ctx context.Context, request *auth_pb.CheckRequest) (*auth_pb.CheckResponse, error) {
	if (iapAuth.Created == time.Time{}) || iapAuth.Created.Before(time.Now().Add(-time.Minute*55)) { // Tokens have a TTL of an hour
		rq, err := http.NewRequest("GET", iapAuth.ServiceURL, nil)
		audience := iapAuth.Audience

		if err != nil {
			fmt.Println("http.NewReqeust: %w", err)
		}

		result, err := makeIAPRequest(rq, audience)
		if err != nil {
			fmt.Println("makeIAPRequest: %w", err)

			iapAuth.Token = ""
			iapAuth.Created = time.Time{}

		} else {
			iapAuth.Token = result
			iapAuth.Created = time.Now()
		}

	}

	headers := map[string]string{
		"Proxy-Authorization": iapAuth.Token,
	}

	return &auth_pb.CheckResponse{
		HttpResponse: &auth_pb.CheckResponse_OkResponse{
			OkResponse: &auth_pb.OkHttpResponse{
				Headers: SetHeaders(headers),
			},
		},
	}, nil
}

func SetHeaders(headers map[string]string) []*corev3.HeaderValueOption {
	var headerValueOptions []*corev3.HeaderValueOption
	for key, value := range headers {
		headerValueOptions = append(headerValueOptions, &corev3.HeaderValueOption{
			Header: &corev3.HeaderValue{
				Key:   key,
				Value: value,
			},
		})
	}

	return headerValueOptions
}

func makeIAPRequest(request *http.Request, audience string) (string, error) {
	ctx := context.Background()

	// client is a http.Client that automatically adds an "Authorization" header
	// to any requests made.
	client, err := idtoken.NewClient(ctx, audience)
	if err != nil {
		return "", fmt.Errorf("idtoken.NewClient: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("client.Do: %w", err)
	}

	authHeader := response.Request.Header["Authorization"][0]
	if authHeader == "" {
		return "", fmt.Errorf("No authorization header found")
	}

	return authHeader, nil
}

func main() {
	// struct with check method
	endPoint := fmt.Sprintf(":%d", 3001)
	listen, _ := net.Listen("tcp", endPoint)

	grpcServer := grpc.NewServer()
	// register envoy proto server
	server := &AuthServer{}
	auth_pb.RegisterAuthorizationServer(grpcServer, server)

	fmt.Println("Server started at port 3001")
	grpcServer.Serve(listen)
}
