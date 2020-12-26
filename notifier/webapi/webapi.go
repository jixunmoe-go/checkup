package webapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sourcegraph/checkup/types"
	"io/ioutil"
	"net/http"
)

// Type should match the package name
const Type = "webapi"

// Notifier consist of all the sub components required to use Slack API
type Notifier struct {
	URL string `json:"url"`
}

// New creates a new Notifier instance based on json config
func New(config json.RawMessage) (Notifier, error) {
	var notifier Notifier
	err := json.Unmarshal(config, &notifier)
	return notifier, err
}

// Type returns the notifier package name
func (Notifier) Type() string {
	return Type
}

// Notify implements notifier interface
func (s Notifier) Notify(results []types.Result) error {
	marshaled, err := json.Marshal(results)
	if err != nil {
		return err
	}

	post := bytes.NewReader(marshaled)
	resp, err := http.Post(s.URL, "application/json", post)
	if err != nil {
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("post did not success: %s", respBytes)
	}

	return nil
}
