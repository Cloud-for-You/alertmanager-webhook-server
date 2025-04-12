package msteams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloud-for-you/alertmanager-webhook-server/internal/logger"
	"github.com/prometheus/alertmanager/template"
)

type MsteamsPayload struct {
	Team     string `json:"team,omitempty"`
	Channel  string `json:"channel,omitempty"`
    Receiver string `json:"receiver"`
    Status   string `json:"status"`
    Alerts   []template.Alert `json:"alerts"`
}

type MsteamsClient struct{
	WebhookURL string
	HTTPClient *http.Client
}

func Client(webhookURL string, httpClient *http.Client) *MsteamsClient {
  if httpClient == nil {
    httpClient = &http.Client{}
  }

  return &MsteamsClient{
    WebhookURL: webhookURL,
    HTTPClient: httpClient,
  }
}

func (r *MsteamsClient) SendMessage(data []byte) error {
    var payload MsteamsPayload
    err := json.Unmarshal(data, &payload)
    if err != nil {
        logger.Log.Errorf("Error while parsing payload: %v", err)
        return err
    }

    // Get MS Teams and Channel DisplayName
    // Dodelat na dynamicke ziskavani z Kubernetes pokud bezi system v K8S nebo z ENV pokud je zadane
    payload.Team = "UNI - Prometheus | DBOS"
    payload.Channel = "LIVE"

    // Prepare payload for sending using http.Post
    formattedPayload, err := json.Marshal(payload)
    if err != nil {
        logger.Log.Errorf("Error while serializing payload: %v", err)
        return err
    }

    // Send the request to Microsoft Teams
    resp, err := http.Post(r.WebhookURL, "application/json", bytes.NewBuffer(formattedPayload))
    if err != nil {
        logger.Log.Errorf("Failed to send message to Microsoft Teams: %v", err)
        return err
    }
    defer resp.Body.Close()

    // Extrakce `runID` z odpovědi
    runID := resp.Header.Get("x-ms-workflow-run-id")
    if runID == "" {
        logger.Log.Errorf("Error: Unable to retrieve workflow-run-id from the response")
        return nil
    }

    // Handle response status
    if resp.StatusCode == http.StatusAccepted {
        logger.Log.Infof("Message accepted (202). Starting async workflow with runID: %s", runID)
    } else if resp.StatusCode != http.StatusOK {
        logger.Log.Errorf("Received non-200 response from Teams: %d", resp.StatusCode)
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    return nil
}