package handlers

import (
	"net/http"
	"regexp"

	"github.com/cloudfoundry-incubator/notifications/models"
	"github.com/cloudfoundry-incubator/notifications/web/params"
	"github.com/ryanmoran/stack"
)

type UpdateNotifications struct {
	updater     NotificationsUpdaterInterface
	errorWriter ErrorWriterInterface
}

func NewUpdateNotifications(notificationsUpdater NotificationsUpdaterInterface, errorWriter ErrorWriterInterface) UpdateNotifications {
	return UpdateNotifications{
		updater:     notificationsUpdater,
		errorWriter: errorWriter,
	}
}

type NotificationsUpdaterInterface interface {
	Update(models.Kind) error
}

func (handler UpdateNotifications) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
	var updateParams params.NotificationUpdateParams

	updateParams, err := params.NewNotificationParams(req.Body)
	if err != nil {
		handler.errorWriter.Write(w, err)
		return
	}

	regex := regexp.MustCompile("/clients/(.*)/notifications/(.*)")
	matches := regex.FindStringSubmatch(req.URL.Path)
	clientID, notificationID := matches[1], matches[2]

	err = handler.updater.Update(updateParams.ToModel(clientID, notificationID))
	if err != nil {
		handler.errorWriter.Write(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
