package autotask_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/scjalliance/autotask"
)

func ExampleNewClient() {
	client, err := autotask.NewClient(autotask.Config{
		Username:        "api@example.com",
		Secret:          "your-secret",
		IntegrationCode: "YOUR_CODE",
		BaseURL:         "https://webservices5.autotask.net/ATServicesRest",
	})
	if err != nil {
		panic(err)
	}
	_ = client
	fmt.Println("client created")
	// Output: client created
}

func ExamplePtr() {
	name := autotask.Ptr("ACME Corp")
	fmt.Println(*name)
	// Output: ACME Corp
}

func ExampleReader_Query() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": 1, "companyName": "ACME Corp"},
			},
			"pageDetails": map[string]any{
				"nextPageUrl": "",
			},
		})
	}))
	defer srv.Close()

	client, err := autotask.NewClient(autotask.Config{
		Username:                 "api@example.com",
		Secret:                   "secret",
		IntegrationCode:          "CODE",
		BaseURL:                  srv.URL,
		DisableRateLimitTracking: true,
	})
	if err != nil {
		panic(err)
	}

	companies, err := client.Companies.Query(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println(len(companies), *companies[0].CompanyName)
	// Output: 1 ACME Corp
}
