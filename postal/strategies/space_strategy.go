package strategies

import (
	"github.com/cloudfoundry-incubator/notifications/models"
	"github.com/cloudfoundry-incubator/notifications/postal"
	"github.com/cloudfoundry-incubator/notifications/postal/utilities"
)

const SpaceEndorsement = "You received this message because you belong to the {{.Space}} space in the {{.Organization}} organization."

type SpaceStrategy struct {
	tokenLoader        utilities.TokenLoaderInterface
	userLoader         utilities.UserLoaderInterface
	spaceLoader        utilities.SpaceLoaderInterface
	organizationLoader utilities.OrganizationLoaderInterface
	findsUserGUIDs     utilities.FindsUserGUIDsInterface
	templatesLoader    utilities.TemplatesLoaderInterface
	mailer             MailerInterface
	receiptsRepo       models.ReceiptsRepoInterface
}

func NewSpaceStrategy(tokenLoader utilities.TokenLoaderInterface, userLoader utilities.UserLoaderInterface, spaceLoader utilities.SpaceLoaderInterface,
	organizationLoader utilities.OrganizationLoaderInterface, findsUserGUIDs utilities.FindsUserGUIDsInterface, templatesLoader utilities.TemplatesLoaderInterface,
	mailer MailerInterface, receiptsRepo models.ReceiptsRepoInterface) SpaceStrategy {

	return SpaceStrategy{
		tokenLoader:        tokenLoader,
		userLoader:         userLoader,
		spaceLoader:        spaceLoader,
		organizationLoader: organizationLoader,
		findsUserGUIDs:     findsUserGUIDs,
		templatesLoader:    templatesLoader,
		mailer:             mailer,
		receiptsRepo:       receiptsRepo,
	}
}

func (strategy SpaceStrategy) Dispatch(clientID, guid string, options postal.Options, conn models.ConnectionInterface) ([]Response, error) {
	responses := []Response{}

	token, err := strategy.tokenLoader.Load()
	if err != nil {
		return responses, err
	}

	space, err := strategy.spaceLoader.Load(guid, token)
	if err != nil {
		return responses, err
	}

	organization, err := strategy.organizationLoader.Load(space.OrganizationGUID, token)
	if err != nil {
		return responses, err
	}

	userGUIDs, err := strategy.findsUserGUIDs.UserGUIDsBelongingToSpace(guid, token)
	if err != nil {
		return responses, err
	}

	users, err := strategy.userLoader.Load(userGUIDs, token)
	if err != nil {
		return responses, err
	}

	templates, err := strategy.templatesLoader.LoadTemplates(clientID, options.KindID)
	if err != nil {
		return responses, postal.TemplateLoadError("An email template could not be loaded")
	}

	err = strategy.receiptsRepo.CreateReceipts(conn, userGUIDs, clientID, options.KindID)
	if err != nil {
		return responses, err
	}

	options.Endorsement = SpaceEndorsement

	responses = strategy.mailer.Deliver(conn, templates, users, options, space, organization, clientID, "")

	return responses, nil
}
