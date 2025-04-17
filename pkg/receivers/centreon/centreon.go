package centreon

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/cloud-for-you/alertmanager-webhook-server/internal/logger"
)

type CentreonClient struct {
	NrdpURL    string
	Token      string
	Hostname   string
	HTTPClient *http.Client
}

// Alertmanager webhook payload
type AlertmanagerPayload struct {
	Alerts []Alert `json:"alerts"`
}

type Alert struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}

func Client() *CentreonClient {
	return &CentreonClient{
		NrdpURL: os.Getenv("CENTREON_NRDP_URL"),
		Token:   os.Getenv("CENTREON_NRDP_TOKEN"),
		Hostname: os.Getenv("CENTREON_MONITORING_HOSTNAME"),
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: os.Getenv("INSECURE_SKIP_VERIFY") == "true",
				},
			},
		},
	}
}

func (r *CentreonClient) SendMessage(data []byte) error {
	var payload AlertmanagerPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		logger.Log.Errorf("invalid Alertmanager payload: %v", err)
		return nil
	}

	if len(payload.Alerts) == 0 {
		logger.Log.Errorf("no alerts to process")
		return nil
	}

	var results []CheckResult
	for _, alert := range payload.Alerts {
		serviceName := alert.Labels["alertname"]

		state := "2" // firing = CRITICAL
		if alert.Status == "resolved" {
			state = "0" // resolved = OK
		}

		output := alert.Annotations["summary"]
		if output == "" {
			output = alert.Annotations["description"]
		}
		if output == "" {
			output = "No description provided"
		}

		result := CheckResult{
			CheckResult: struct {
				Type string `json:"type"`
			}{Type: "service"},
			Hostname:    os.Getenv("CENTREON_MONITORING_HOSTNAME"),
			Servicename: serviceName,
			State:       state,
			Output:      output,
		}
		results = append(results, result)
	}

	jsonPayload, err := json.Marshal(CheckResultsPayload{
		CheckResults: results,
	})
	if err != nil {
		logger.Log.Errorf("failed to marshal JSON payload: %v", err)
		return nil
	}

	form := url.Values{}
	form.Set("token", r.Token)
	form.Set("cmd", "submitcheck")
	form.Set("json", string(jsonPayload))

	req, err := http.NewRequest("POST", r.NrdpURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		logger.Log.Errorf("failed to build NRDP request: %v", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		logger.Log.Errorf("failed to send NRDP request: %v", err)
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || !bytes.Contains(body, []byte("OK")) {
		logger.Log.Errorf("Centreon NRDP response error [%d]: %s", resp.StatusCode, string(body))
		return nil
	}

	logger.Log.Infof("NRDP submission successful: %d alert(s) sent", len(results))
	return nil
}