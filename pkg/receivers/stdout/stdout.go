package receivers

import (
	"github.com/cloud-for-you/alertmanager-webhook-server/internal/logger"
)

type STDOUTReceiver struct{}

func NewSTDOUTReceiver() *STDOUTReceiver {
    return &STDOUTReceiver{}
}

func (r *STDOUTReceiver) Producer(data []byte) error {
	logger.Log.Infof("%s", string(data))
  return nil
}
