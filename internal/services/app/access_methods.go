package app

import (
	"errors"
	"strings"

	"github.com/brewbits-co/releasedesk/internal/domains/app"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/teams"
	"github.com/brewbits-co/releasedesk/pkg/tokenhash"
)

func (s *service) IsInstanceAdmin(userID int) bool {
	account, err := s.userRepo.GetByID(userID)
	return err == nil && account.Role == values.Admin
}

func (s *service) GetBySlug(slug values.Slug) (app.App, error) {
	return s.appRepo.GetBySlug(slug)
}

func (s *service) GetByID(id int) (app.App, error) {
	return s.appRepo.GetByID(id)
}

func (s *service) CanView(userID int, slug values.Slug) (bool, error) {
	if s.IsInstanceAdmin(userID) {
		return true, nil
	}
	entity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return false, err
	}
	_, err = s.memberRepo.Get(entity.ID, userID)
	return err == nil, nil
}

// CanEditContent is true for an app admin, or for an instance admin, manager, or developer who can see the app.
func (s *service) CanEditContent(userID int, slug values.Slug) (bool, error) {
	if ok, err := s.CanAdmin(userID, slug); err != nil || ok {
		return ok, err
	}
	account, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, nil
	}
	if account.Role != values.Admin && account.Role != values.Manager && account.Role != values.Developer {
		return false, nil
	}
	return s.CanView(userID, slug)
}

func (s *service) CanAdmin(userID int, slug values.Slug) (bool, error) {
	if s.IsInstanceAdmin(userID) {
		return true, nil
	}
	entity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return false, err
	}
	member, err := s.memberRepo.Get(entity.ID, userID)
	if err != nil {
		return false, nil
	}
	return member.Role == app.MemberAdmin, nil
}

func (s *service) UpdateSettings(userID int, slug values.Slug, description string, private bool, teamsWebhook string) (app.App, error) {
	ok, err := s.CanAdmin(userID, slug)
	if err != nil || !ok {
		return app.App{}, errors.New("not allowed")
	}
	if teamsWebhook != "-" {
		if err = teams.ValidateIncomingWebhook(teamsWebhook); err != nil {
			return app.App{}, err
		}
	}
	entity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return app.App{}, err
	}
	if teamsWebhook == "-" {
		teamsWebhook = ""
	} else if teamsWebhook == "" {
		teamsWebhook = entity.TeamsWebhookURL
	}
	if err = s.appRepo.UpdateSettings(entity.ID, description, private, teamsWebhook); err != nil {
		return app.App{}, err
	}
	return s.appRepo.GetBySlug(slug)
}

func (s *service) ListMembers(userID int, slug values.Slug) ([]app.Member, error) {
	ok, err := s.CanView(userID, slug)
	if err != nil || !ok {
		return nil, errors.New("not allowed")
	}
	entity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	return s.memberRepo.ListByApp(entity.ID)
}

func (s *service) AddMember(actorID int, slug values.Slug, memberUserID int, role app.MemberRole) error {
	ok, err := s.CanAdmin(actorID, slug)
	if err != nil || !ok {
		return errors.New("not allowed")
	}
	if role != app.MemberAdmin && role != app.MemberUser {
		return errors.New("role must be admin or user")
	}
	if _, err = s.userRepo.GetByID(memberUserID); err != nil {
		return errors.New("user not found")
	}
	entity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return err
	}
	return s.memberRepo.Save(&app.Member{AppID: entity.ID, UserID: memberUserID, Role: role})
}

func (s *service) RemoveMember(actorID int, slug values.Slug, memberUserID int) error {
	ok, err := s.CanAdmin(actorID, slug)
	if err != nil || !ok {
		return errors.New("not allowed")
	}
	entity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return err
	}
	member, err := s.memberRepo.Get(entity.ID, memberUserID)
	if err != nil {
		return errors.New("member not found")
	}
	if member.Role == app.MemberAdmin {
		members, err := s.memberRepo.ListByApp(entity.ID)
		if err != nil {
			return err
		}
		admins := 0
		for _, existing := range members {
			if existing.Role == app.MemberAdmin {
				admins++
			}
		}
		if admins <= 1 {
			return errors.New("keep at least one app admin")
		}
	}
	return s.memberRepo.Delete(entity.ID, memberUserID)
}

func (s *service) MintToken(actorID int, slug values.Slug, name, scope string) (string, app.CIToken, error) {
	ok, err := s.CanAdmin(actorID, slug)
	if err != nil || !ok {
		return "", app.CIToken{}, errors.New("not allowed")
	}
	entity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return "", app.CIToken{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", app.CIToken{}, errors.New("name is required")
	}
	if scope != app.TokenScopeRead {
		scope = app.TokenScopeWrite
	}
	raw, prefix, hash, err := tokenhash.Mint()
	if err != nil {
		return "", app.CIToken{}, err
	}
	token := app.CIToken{AppID: entity.ID, Name: name, Prefix: prefix, Hash: hash, Scope: scope}
	if err = s.tokenRepo.Save(&token); err != nil {
		return "", app.CIToken{}, err
	}
	return raw, token, nil
}

func (s *service) ListTokens(actorID int, slug values.Slug) ([]app.CIToken, error) {
	ok, err := s.CanAdmin(actorID, slug)
	if err != nil || !ok {
		return nil, errors.New("not allowed")
	}
	entity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	return s.tokenRepo.ListByApp(entity.ID)
}

func (s *service) RevokeToken(actorID int, slug values.Slug, tokenID int) error {
	ok, err := s.CanAdmin(actorID, slug)
	if err != nil || !ok {
		return errors.New("not allowed")
	}
	entity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return err
	}
	if err = s.tokenRepo.Revoke(entity.ID, tokenID); err != nil {
		return errors.New("token not found")
	}
	return nil
}

func (s *service) LookupToken(raw string) (app.CIToken, error) {
	return s.tokenRepo.GetByHash(tokenhash.Hash(raw))
}

func (s *service) ListGroups(userID int, slug values.Slug) ([]app.DistributionGroup, error) {
	ok, err := s.CanView(userID, slug)
	if err != nil || !ok {
		return nil, errors.New("not allowed")
	}
	entity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	return s.groupRepo.ListByApp(entity.ID)
}

func (s *service) SaveGroup(userID int, slug values.Slug, name string, restrict bool) error {
	ok, err := s.CanAdmin(userID, slug)
	if err != nil || !ok {
		return errors.New("not allowed")
	}
	entity, err := s.appRepo.GetBySlug(slug)
	if err != nil {
		return err
	}
	return s.groupRepo.Save(&app.DistributionGroup{AppID: entity.ID, Name: name, RestrictPortal: restrict})
}
