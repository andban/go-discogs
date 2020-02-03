package discogs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	discogsAPI = "https://api.discogs.com"
)

// Options is a set of options to use discogs API client
type Options struct {
	URL       string
	Currency  string
	UserAgent string
	Token     string
	Key       string
	Secret    string
}

// Client is a Discogs client for making Discogs API requests.
type Client struct {
	Database *DatabaseService
	Search   *SearchService
}

var header *http.Header

// NewClient returns a new Client.
func NewClient(o *Options) (*Client, error) {
	header = &http.Header{}

	if o == nil || o.UserAgent == "" {
		return nil, ErrUserAgentInvalid
	}

	header.Add("User-Agent", o.UserAgent)

	cur, err := currency(o.Currency)
	if err != nil {
		return nil, err
	}

	// if api key is present, secret must be present too (and vice versa)
	if (o.Key == "") != (o.Secret == "") {
		return nil, ErrCredentialsIncomplete
	}

	// set token, it's required for some queries like search
	if o.Token != "" {
		header.Add("Authorization", "Discogs token="+o.Token)
	} else if o.Key != "" && o.Secret != "" {
		header.Add("Authorization", "Discogs key="+o.Key+", secret="+o.Secret)
	}

	if o.URL == "" {
		o.URL = discogsAPI
	}

	return &Client{
		Database: newDatabaseService(o.URL, cur),
		Search:   newSearchService(o.URL + "/database/search"),
	}, nil
}

// currency validates currency for marketplace data.
// Defaults to the authenticated users currency. Must be one of the following:
// USD GBP EUR CAD AUD JPY CHF MXN BRL NZD SEK ZAR
func currency(c string) (string, error) {
	switch c {
	case "USD", "GBP", "EUR", "CAD", "AUD", "JPY", "CHF", "MXN", "BRL", "NZD", "SEK", "ZAR":
		return c, nil
	case "":
		return "USD", nil
	default:
		return "", ErrCurrencyNotSupported
	}
}

func request(path string, params url.Values, resp interface{}) error {
	r, err := http.NewRequest("GET", path+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	r.Header = *header

	client := &http.Client{}
	response, err := client.Do(r)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		switch response.StatusCode {
		case http.StatusUnauthorized:
			return ErrUnauthorized
		default:
			return fmt.Errorf("unknown error: %s", response.Status)
		}
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &resp)
}
