package slack

import (
	"encoding/json"
	"fmt"
	"github.com/sourcegraph/checkup/storage/util"
	"strconv"
	"strings"

	slack "github.com/ashwanthkumar/slack-go-webhook"

	"github.com/sourcegraph/checkup/types"
)

// Type should match the package name
const Type = "slack"

// Notifier consist of all the sub components required to use Slack API
type Notifier struct {
	Username string `json:"username"`
	Channel  string `json:"channel"`
	Webhook  string `json:"webhook"`
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
	errs := make(types.Errors, 0)
	for _, result := range results {
		if !result.Healthy {
			if err := s.Send(&result); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

func (s Notifier) NotifyUnhealthy(result *types.Result) error {
	if !result.Healthy {
		return s.Send(result)
	}
	return nil
}

// Notify implements DownTimeNotifier interface
func (s Notifier) NotifyDowntime(currResults, prevResults []types.Result, reader types.HistoryReader) error {
	errs := make(types.Errors, 0)
	for _, checkResult := range currResults {
		var prevResult = util.FindResultOfSameType(&checkResult, prevResults)
		err := s.NotifyOneResult(&checkResult, prevResult, reader)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// Send request via Slack API to create incident
func (s Notifier) Send(result *types.Result) error {
	return s.SendOnline(result, 0)
}

func (s Notifier) SendOnline(result *types.Result, lastOnlineTime int64) error {
	color := "good"
	if !result.Healthy {
		color = "danger"
	}
	attach := slack.Attachment{}
	attach.AddField(slack.Field{Title: result.Title, Value: result.Endpoint})
	attach.AddField(slack.Field{Title: "Status", Value: strings.ToUpper(fmt.Sprint(result.Status()))})
	if lastOnlineTime > 0 {
		downtime := (result.Timestamp - lastOnlineTime) / (1e+6)
		attach.AddField(slack.Field{Title: "Downtime", Value: strconv.FormatInt(downtime, 10)})
	}
	attach.Color = &color
	payload := slack.Payload{
		Text:        result.Title,
		Username:    s.Username,
		Channel:     s.Channel,
		Attachments: []slack.Attachment{attach},
	}

	errs := slack.Send(s.Webhook, "", payload)
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (s Notifier) NotifyOneResult(checkResult *types.Result, prev *types.Result, reader types.HistoryReader) error {
	if prev == nil {
		return s.NotifyUnhealthy(checkResult)
	}

	// We found a history entry
	if prev.Healthy == checkResult.Healthy {
		return nil
	}

	// Previously OK, now dead
	if prev.Healthy {
		return s.NotifyUnhealthy(checkResult)
	}

	return s.SendOnline(checkResult, reader.FindLastWorkingTime(checkResult))
}
