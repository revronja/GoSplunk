package GoSplunk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type SplunkClient struct {
	Username, Password, BaseUrl string
	token                       Token
	hc                          *http.Client
}

type SearchResponse struct {
	XMLName xml.Name `xml:"response"`
	Sid     string   `xml:"sid"`
}

type Token struct {
	Value string `json:"sessionKey"`
}

func NewAuthClient(hc *http.Client, username, password, baseurl string) *SplunkClient {
	return &SplunkClient{
		hc:       hc,
		Username: username,
		Password: password,
		BaseUrl:  baseurl,
	}
}

func (sc *SplunkClient) Logon() (Token, error) {
	data := make(url.Values)
	data.Add("username", sc.Username)
	data.Add("password", sc.Password)
	data.Add("output_mode", "json")
	var payload io.Reader
	payload = bytes.NewBufferString(data.Encode())
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/services/auth/login", sc.BaseUrl), payload)
	resp, err := sc.hc.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()
	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	var token Token
	unmarshallErr := json.Unmarshal(respBodyBytes, &token)
	sc.token = token
	return token, unmarshallErr
}

func (sc *SplunkClient) NewSearchJob() (SearchResponse, error) {
	data := make(url.Values)
	data.Add("search", "search*")
	var payload io.Reader
	payload = bytes.NewBufferString(data.Encode())
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/services/search/jobs", sc.BaseUrl), payload)
	req.Header.Add("Authorization", fmt.Sprintf("Splunk %s", sc.token.Value))
	resp, err := sc.hc.Do(req)
	if err != nil {
		return SearchResponse{}, err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return SearchResponse{}, err
	}
	var sr SearchResponse
	unmarshalErr := json.Unmarshal(bytes, &sr)
	return sr, unmarshalErr
}

func (sc *SplunkClient) GetSearches() (string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/services/search/jobs", sc.BaseUrl), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Splunk %s", sc.token.Value))
	resp, err := sc.hc.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	stringbody := string(bytes)
	return stringbody, nil
}

func NewHttpClient() (*http.Client, error) {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}

	return http.DefaultClient, nil
}

func SplunkErrCodes(code int) error {
	switch code {
	case 200:
		return nil
	case 400:
		return fmt.Errorf("400	Request error. See response body for details.")
	case 401:
		return fmt.Errorf("401	Authentication failure, invalid access credentials. Check headers!")
	case 404:
		return fmt.Errorf("404, Requested endpoint does not exist.")
	case 409:
		return fmt.Errorf("409	Invalid operation for this endpoint. See response body for details.")
	case 500:
		return fmt.Errorf("500	Unspecified internal server error. See response body for details")
	case 503:
		return fmt.Errorf("503 Feature is disabled in configuration file.")
	default:
		return fmt.Errorf("status code (%d)", code)
	}
}
