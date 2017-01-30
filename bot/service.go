package bot

import (
	"io/ioutil"
	"net/http"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/gziphandler"
	"github.com/Sirupsen/logrus"
	"github.com/marzagao/gcp-error-logs-bot/config"
)

type BotService struct {
	config *config.Config
	logger *logrus.Logger
	bot    *Bot
}

func NewBotService(cfg *config.Config, logger *logrus.Logger) (*BotService, error) {
	service := &BotService{
		config: cfg,
	}
	if cfg == nil {
		return service, nil
	}
	if logger == nil {
		logger = logrus.New()
	}
	service.logger = logger
	bot, err := NewBot(service.logger, cfg.GCP.ProjectID, cfg.GCP.LogName, cfg.Timezone, cfg.Slack.WebhookURL)
	service.bot = bot
	return service, err
}

func (s *BotService) handleStackdriverPayload(r *http.Request) (int, interface{}, error) {
	input, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.logger.Errorf("error while reading stackdriver payload: %s", err.Error())
		return http.StatusInternalServerError, nil, err
	}
	s.logger.Debugf("got stackdriver payload: %s", string(input))
	s.bot.Run(string(input))
	return http.StatusOK, "Payload retrieved successfully", nil
}

func (s *BotService) Prefix() string {
	return ""
}

func (s *BotService) Middleware(h http.Handler) http.Handler {
	return gziphandler.GzipHandler(h)
}

func (s *BotService) JSONMiddleware(j server.JSONEndpoint) server.JSONEndpoint {
	return func(r *http.Request) (int, interface{}, error) {
		status, res, err := j(r)
		if err != nil {
			server.LogWithFields(r).WithFields(logrus.Fields{
				"error": err,
			}).Error("problems with serving request")
			return http.StatusServiceUnavailable, nil, &jsonErr{"sorry, this service is unavailable"}
		}
		server.LogWithFields(r).Info("success!")
		return status, res, nil
	}
}

func (s *BotService) JSONEndpoints() map[string]map[string]server.JSONEndpoint {
	return map[string]map[string]server.JSONEndpoint{
		"/webhook": {
			"POST": s.handleStackdriverPayload,
		},
	}
}

type jsonErr struct {
	Err string `json:"error"`
}

func (e *jsonErr) Error() string {
	return e.Err
}
