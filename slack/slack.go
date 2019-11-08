package slack

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

func PostAsFile(param FileParam, webhookURL string) ([]byte, error) {
	b, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	resp, err := http.PostForm(
		webhookURL,
		url.Values{"payload": {string(b)}},
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
