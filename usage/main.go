package main

import (
	"fmt"
	"splunk"
)

func main() {
	hc, err := splunk.NewHttpClient()
	if err != nil {
		fmt.Printf("Couldn't login to splunk: %v\n", err)
	}
	sc := splunk.NewAuthClient(hc, "admin", "changeme, "https://10.0.0.0:8089")

	token, err := sc.NewLogon()
	if err != nil {
		fmt.Printf("Couldn't login to splunk: %v\n", err)
	}
	fmt.Println("Session key: ", token.Value)

}
