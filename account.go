package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func getAccountDetails(ctx *cli.Context) error {
	// Getting the account details
	var baseURL = getNGBaseURL(ctx)
	accountsEndpoint := "accounts/" + cliCdRequestData.Account
	url := GetUrlWithQueryParams("", baseURL, accountsEndpoint, map[string]string{
		"accountIdentifier": cliCdRequestData.Account,
	})
	resp, err := Get(url, cliCdRequestData.AuthToken)

	if err != nil {
		fmt.Printf("Response status: %s \n", resp.Status)
		fmt.Printf("Response code: %s \n", resp.Code)
		fmt.Printf("Response resource: %s \n", resp.Resource)
		fmt.Printf("Response messages: %s \n", resp.Messages)
		return nil
	}

	return nil
}
