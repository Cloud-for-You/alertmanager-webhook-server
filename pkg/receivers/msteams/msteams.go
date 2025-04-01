package msteams

import (
	"github.com/cloud-for-you/alertmanager-webhook-server/internal/logger"
)

type MsteamsClient struct{}

func Client() *MsteamsClient {
    return &MsteamsClient{}
}

func (r *MsteamsClient) SendMessage(data []byte) error {
	logger.Log.Infof("%s", string(data))
  return nil
}