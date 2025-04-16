package centreon

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cloud-for-you/alertmanager-webhook-server/internal/logger"
)

type CentreonClient struct{
	NrdpURL string
	Token string
	Hostname string
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

type NRDPRequest struct {
	XMLName     xml.Name      `xml:"nrdp"`
	CheckResult []CheckResult `xml:"checkresults>checkresult"`
}

type CheckResult struct {
	Type         string `xml:"type,attr"`      // "service"
	CheckType    string `xml:"checktype,attr"` // "1" = passive
	Hostname     string `xml:"hostname"`
	ServiceName  string `xml:"servicename"`
	State        int    `xml:"state"`  // 0 = OK, 2 = CRITICAL
	Output       string `xml:"output"` // Alert summary
}


func Client() *CentreonClient {
	return &CentreonClient{
		NrdpURL: "https://centreon.example.com/nrdp/", // ← os.Getenv("CENTREON_NRDP_URL")
		Token:   "your-nrdp-token",                    // ← os.Getenv("CENTREON_NRDP_TOKEN")
		Hostname: "inf-tocp",                          // ← os.Getenv("CLUSTER_NAME")
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
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
		if serviceName == "" {
			serviceName = "UnnamedAlert"
		}

		state := 2 // firing = CRITICAL
		if alert.Status == "resolved" {
			state = 0 // resolved = OK
		}

		output := alert.Annotations["summary"]
		if output == "" {
			output = alert.Annotations["description"]
		}
		if output == "" {
			output = "No description provided"
		}

		result := CheckResult{
			Type:        "service",
			CheckType:   "1",
			Hostname:    r.Hostname,
			ServiceName: serviceName,
			State:       state,
			Output:      output,
		}
		results = append(results, result)
	}

	xmlPayload, err := xml.MarshalIndent(NRDPRequest{CheckResult: results}, "", "  ")
	if err != nil {
		logger.Log.Errorf("failed to marshal NRDP XML: %v", err)
		return nil
	}

	form := url.Values{}
	form.Set("token", r.Token)
	form.Set("cmd", "submitcheck")
	form.Set("xml", string(xmlPayload))

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