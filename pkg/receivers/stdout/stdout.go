package stdout

import (
	"github.com/cloud-for-you/alertmanager-webhook-server/internal/logger"
)

type StdoutClient struct{}

func Client() *StdoutClient {
    return &StdoutClient{}
}

func (r *StdoutClient) SendMessage(data []byte) error {
	logger.Log.Infof("%s", string(data))
  return nil
}
