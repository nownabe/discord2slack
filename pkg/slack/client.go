package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"google.golang.org/appengine/urlfetch"
)

type Client struct {
	WebhookURL string
}

func New(url string) (*Client, error) {
	return &Client{url}, nil
}

func (c *Client) Post(ctx context.Context, m Message) error {
	js, err := json.Marshal(m)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.WebhookURL, bytes.NewBuffer(js))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := urlfetch.Client(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		serr := errors.Errorf("slack api error: status code %d", resp.StatusCode)

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(serr, err.Error())
		}

		return errors.Wrapf(serr, string(respBody))
	}

	return nil
}
