package biz

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/samber/lo"
	"go.uber.org/fx"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/apikey"
	"github.com/looplj/axonhub/internal/ent/apikeyprofiletemplate"
	"github.com/looplj/axonhub/internal/log"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/xerrors"
)

type APIKeyProfileTemplateServiceParams struct {
	fx.In

	Ent *ent.Client
}

type APIKeyProfileTemplateService struct {
	*AbstractService
	apiKeyInvalidator func(context.Context, string)
}

func NewAPIKeyProfileTemplateService(params APIKeyProfileTemplateServiceParams) *APIKeyProfileTemplateService {
	return &APIKeyProfileTemplateService{
		AbstractService: &AbstractService{
			db: params.Ent,
		},
	}
}

func (s *APIKeyProfileTemplateService) SetAPIKeyInvalidator(invalidator func(context.Context, string)) {
	s.apiKeyInvalidator = invalidator
}

func (s *APIKeyProfileTemplateService) invalidateAPIKeys(ctx context.Context, keys []string) {
	if s.apiKeyInvalidator == nil {
		return
	}

	for _, key := range lo.Uniq(keys) {
		s.apiKeyInvalidator(ctx, key)
	}
}

func (s *APIKeyProfileTemplateService) CreateTemplate(ctx context.Context, input ent.CreateAPIKeyProfileTemplateInput, profile *objects.APIKeyProfile) (*ent.APIKeyProfileTemplate, error) {
	client := s.entFromContext(ctx)

	if profile != nil {
		profile.TemplateID = nil
		profile.TemplateName = ""
		profile.Name = input.Name
		if err := normalizeAndValidateProfileRoutingPolicy(profile); err != nil {
			return nil, err
		}
	}

	create := client.APIKeyProfileTemplate.Create().
		SetInput(input).
		SetProfile(profile)

	template, err := create.Save(ctx)
	if err != nil {
		// Name uniqueness is enforced by the (project_id, name, deleted_at) unique
		// index; surface a friendly error instead of a raw constraint violation.
		if ent.IsConstraintError(err) {
			return nil, xerrors.DuplicateNameError("Template", input.Name)
		}

		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

func (s *APIKeyProfileTemplateService) GetTemplate(ctx context.Context, id int) (*ent.APIKeyProfileTemplate, error) {
	client := s.entFromContext(ctx)

	template, err := client.APIKeyProfileTemplate.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return template, nil
}

// GetForRead loads a template by id or name for read-only access. Exactly one
// of id or name must be non-nil.
//
// Like APIKeyService.GetForRead, it goes through the context-bound ent client
// so the APIKeyProfileTemplate privacy policy runs: an API key principal must
// hold read_api_keys and is filtered to templates inside its own project, where
// names are unique (DB index on project_id+name) — so a name identifies at most
// one template and foreign templates surface as NotFound.
func (s *APIKeyProfileTemplateService) GetForRead(ctx context.Context, id *int, name *string) (*ent.APIKeyProfileTemplate, error) {
	if (id == nil) == (name == nil) {
		return nil, fmt.Errorf("exactly one of template id or name must be provided")
	}

	client := s.entFromContext(ctx)
	q := client.APIKeyProfileTemplate.Query()

	switch {
	case id != nil:
		q = q.Where(apikeyprofiletemplate.IDEQ(*id))
	case name != nil:
		q = q.Where(apikeyprofiletemplate.NameEQ(*name))
	}

	template, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	return template, nil
}

func (s *APIKeyProfileTemplateService) ListTemplates(ctx context.Context, projectID int) ([]*ent.APIKeyProfileTemplate, error) {
	client := s.entFromContext(ctx)

	templates, err := client.APIKeyProfileTemplate.Query().
		Where(apikeyprofiletemplate.ProjectIDEQ(projectID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	return templates, nil
}

func (s *APIKeyProfileTemplateService) UpdateTemplate(ctx context.Context, id int, input ent.UpdateAPIKeyProfileTemplateInput, profile *objects.APIKeyProfile) (*ent.APIKeyProfileTemplate, error) {
	var template *ent.APIKeyProfileTemplate
	var synchronizedKeys []string
	err := s.RunInTransaction(ctx, func(ctx context.Context) error {
		client := s.entFromContext(ctx)
		existing, getErr := client.APIKeyProfileTemplate.Get(ctx, id)
		if getErr != nil {
			return fmt.Errorf("failed to get template: %w", getErr)
		}

		update := client.APIKeyProfileTemplate.UpdateOneID(id).
			SetInput(input)

		publishedProfile := profile
		if publishedProfile == nil && input.Name != nil {
			publishedProfile = existing.Profile.Clone()
		}

		if publishedProfile != nil {
			publishedProfile.TemplateID = nil
			publishedProfile.TemplateName = ""
			if existing.Profile != nil && existing.Profile.TemplateSync {
				publishedProfile.TemplateSync = true
			}
			if err := normalizeAndValidateProfileRoutingPolicy(publishedProfile); err != nil {
				return err
			}

			if input.Name != nil {
				publishedProfile.Name = *input.Name
			} else {
				publishedProfile.Name = existing.Name
			}
			update.SetProfile(publishedProfile)
		}

		var saveErr error
		template, saveErr = update.Save(ctx)
		if saveErr != nil {
			// The unique index on (project_id, name, deleted_at) is the source of
			// truth for name uniqueness; map it to a friendly error.
			if ent.IsConstraintError(saveErr) {
				return xerrors.DuplicateNameError("Template", lo.FromPtr(input.Name))
			}

			return fmt.Errorf("failed to update template: %w", saveErr)
		}

		if publishedProfile != nil {
			var syncErr error
			synchronizedKeys, syncErr = s.syncLinkedProfiles(ctx, existing, template, publishedProfile)
			if syncErr != nil {
				return syncErr
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	s.invalidateAPIKeys(ctx, synchronizedKeys)

	return template, nil
}

func (s *APIKeyProfileTemplateService) DeleteTemplate(ctx context.Context, id int) (*ent.APIKeyProfileTemplate, error) {
	var template *ent.APIKeyProfileTemplate
	var detachedKeys []string
	err := s.RunInTransaction(ctx, func(ctx context.Context) error {
		client := s.entFromContext(ctx)

		var getErr error
		template, getErr = client.APIKeyProfileTemplate.Get(ctx, id)
		if getErr != nil {
			return fmt.Errorf("failed to get template for deletion: %w", getErr)
		}

		var detachErr error
		detachedKeys, detachErr = s.detachLinkedProfiles(ctx, template)
		if detachErr != nil {
			return detachErr
		}

		getErr = client.APIKeyProfileTemplate.DeleteOneID(id).Exec(ctx)
		if getErr != nil {
			return fmt.Errorf("failed to delete template: %w", getErr)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	s.invalidateAPIKeys(ctx, detachedKeys)

	return template, nil
}

// ActivateTemplateProfile switches an API key to a profile already linked to
// the requested template without replacing the rest of the profile settings.
func (s *APIKeyProfileTemplateService) ActivateTemplateProfile(ctx context.Context, apiKeyID, templateID int) (*ent.APIKey, error) {
	var updatedKey *ent.APIKey
	previousActiveProfile := ""
	err := s.RunInTransaction(ctx, func(ctx context.Context) error {
		client := s.entFromContext(ctx)

		template, err := client.APIKeyProfileTemplate.Get(ctx, templateID)
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}

		apiKey, err := client.APIKey.Get(ctx, apiKeyID)
		if err != nil {
			return fmt.Errorf("failed to get API key: %w", err)
		}
		if template.ProjectID != apiKey.ProjectID {
			return fmt.Errorf("template and API key must belong to the same project")
		}
		if apiKey.Profiles == nil {
			return fmt.Errorf("API key has no linked profile for template '%s'", template.Name)
		}
		previousActiveProfile = apiKey.Profiles.ActiveProfile

		profileName := ""
		for i := range apiKey.Profiles.Profiles {
			profile := &apiKey.Profiles.Profiles[i]
			if profile.TemplateID != nil && *profile.TemplateID == template.ID {
				profileName = profile.Name
				break
			}
		}
		if profileName == "" {
			return fmt.Errorf("API key has no linked profile for template '%s'", template.Name)
		}

		apiKey.Profiles.ActiveProfile = profileName
		updatedKey, err = client.APIKey.UpdateOneID(apiKey.ID).
			SetProfiles(apiKey.Profiles).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to activate API key template profile: %w", err)
		}

		return nil
	})
	if err != nil {
		log.Warn(ctx, "api key profile activation failed",
			log.Int("api_key_id", apiKeyID),
			log.Int("template_id", templateID),
			log.String("source", "profile_template_quick_switch"),
			log.String("from_profile", previousActiveProfile),
			log.Cause(err))
		return nil, err
	}
	s.invalidateAPIKeys(ctx, []string{updatedKey.Key})
	log.Info(ctx, "api key profile activated",
		log.Int("api_key_id", updatedKey.ID),
		log.Int("project_id", updatedKey.ProjectID),
		log.Int("template_id", templateID),
		log.String("source", "profile_template_quick_switch"),
		log.String("from_profile", previousActiveProfile),
		log.String("to_profile", updatedKey.Profiles.ActiveProfile))

	return updatedKey, nil
}

func (s *APIKeyProfileTemplateService) LoadTemplate(ctx context.Context, templateID, apiKeyID int) (*ent.APIKey, error) {
	var updatedKey *ent.APIKey
	err := s.RunInTransaction(ctx, func(ctx context.Context) error {
		client := s.entFromContext(ctx)

		template, err := client.APIKeyProfileTemplate.Get(ctx, templateID)
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}

		apiKey, getErr := client.APIKey.Get(ctx, apiKeyID)
		if getErr != nil {
			return fmt.Errorf("failed to get API key: %w", getErr)
		}

		if template.ProjectID != apiKey.ProjectID {
			return fmt.Errorf("template and API key must belong to the same project")
		}

		templateProfile := template.Profile.Clone()
		if templateProfile == nil {
			return fmt.Errorf("template has no profile")
		}
		if err := normalizeAndValidateProfileRoutingPolicy(templateProfile); err != nil {
			return err
		}

		existingProfiles := apiKey.Profiles
		if existingProfiles == nil {
			existingProfiles = &objects.APIKeyProfiles{}
		}

		for i := range existingProfiles.Profiles {
			profile := &existingProfiles.Profiles[i]
			if profile.TemplateID != nil && *profile.TemplateID == template.ID {
				return fmt.Errorf("API key already has a profile linked to template '%s'", template.Name)
			}
			if normalizeProfileName(profile.Name) == normalizeProfileName(template.Name) {
				return fmt.Errorf("profile name '%s' conflicts with template name '%s'", profile.Name, template.Name)
			}
		}

		templateProfile.Name = template.Name
		templateProfile.TemplateID = lo.ToPtr(template.ID)
		templateProfile.TemplateName = template.Name

		existingProfiles.Profiles = append(existingProfiles.Profiles, *templateProfile)

		updatedKey, err = client.APIKey.UpdateOneID(apiKeyID).
			SetProfiles(existingProfiles).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update API key profiles: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	s.invalidateAPIKeys(ctx, []string{updatedKey.Key})

	return updatedKey, nil
}

// normalizeLinkedProfileNames makes the template name the single display and
// storage name for linked profiles. It also rejects independent profile names
// that would be indistinguishable from a project template in the UI.
func (s *APIKeyProfileTemplateService) normalizeLinkedProfileNames(
	ctx context.Context,
	existingKey *ent.APIKey,
	nextProfiles *objects.APIKeyProfiles,
) error {
	if existingKey == nil || nextProfiles == nil {
		return nil
	}

	client := s.entFromContext(ctx)
	templates, err := client.APIKeyProfileTemplate.Query().
		Where(apikeyprofiletemplate.ProjectIDEQ(existingKey.ProjectID)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to list API key profile templates: %w", err)
	}

	templatesByID := make(map[int]*ent.APIKeyProfileTemplate, len(templates))
	templatesByName := make(map[string]*ent.APIKeyProfileTemplate, len(templates))
	for _, template := range templates {
		templatesByID[template.ID] = template
		templatesByName[normalizeProfileName(template.Name)] = template
	}

	var activeTemplateID *int
	for i := range nextProfiles.Profiles {
		profile := &nextProfiles.Profiles[i]
		if profile.Name == nextProfiles.ActiveProfile && profile.TemplateID != nil {
			activeTemplateID = lo.ToPtr(*profile.TemplateID)
			break
		}
	}
	if activeTemplateID == nil && existingKey.Profiles != nil && existingKey.Profiles.ActiveProfile == nextProfiles.ActiveProfile {
		for i := range existingKey.Profiles.Profiles {
			profile := &existingKey.Profiles.Profiles[i]
			if profile.Name == existingKey.Profiles.ActiveProfile && profile.TemplateID != nil {
				activeTemplateID = lo.ToPtr(*profile.TemplateID)
				break
			}
		}
	}

	for i := range nextProfiles.Profiles {
		profile := &nextProfiles.Profiles[i]
		if profile.TemplateID == nil {
			profile.TemplateName = ""
			profile.TemplateSync = false
			continue
		}

		template, ok := templatesByID[*profile.TemplateID]
		if !ok {
			profile.TemplateID = nil
			profile.TemplateName = ""
			profile.TemplateSync = false
			continue
		}

		profile.Name = template.Name
		profile.TemplateName = template.Name
		profile.TemplateSync = template.Profile != nil && template.Profile.TemplateSync
	}

	if activeTemplateID != nil {
		if template, ok := templatesByID[*activeTemplateID]; ok {
			nextProfiles.ActiveProfile = template.Name
		}
	}

	for i := range nextProfiles.Profiles {
		profile := &nextProfiles.Profiles[i]
		if profile.TemplateID != nil {
			continue
		}

		if template, ok := templatesByName[normalizeProfileName(profile.Name)]; ok {
			return fmt.Errorf("profile name '%s' conflicts with template name '%s'", profile.Name, template.Name)
		}
	}

	return nil
}

// CountLinkedProfiles returns how many API key profiles currently follow a
// template. API key profiles are embedded JSON, so the count is intentionally
// computed from the project's keys instead of introducing a second source of
// truth.
func (s *APIKeyProfileTemplateService) CountLinkedProfiles(ctx context.Context, template *ent.APIKeyProfileTemplate) (int, error) {
	client := s.entFromContext(ctx)
	keys, err := client.APIKey.Query().
		Where(apikey.ProjectIDEQ(template.ProjectID)).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list API keys linked to template: %w", err)
	}

	count := 0
	for _, key := range keys {
		if key.Profiles == nil {
			continue
		}
		for i := range key.Profiles.Profiles {
			profile := &key.Profiles.Profiles[i]
			if (profile.TemplateID != nil && *profile.TemplateID == template.ID) || isLegacyTemplateMatch(profile, template) {
				count++
			}
		}
	}

	return count, nil
}

// reconcileLinkedProfileChanges applies API key edits according to each
// template's synchronization policy. Ordinary linked templates detach on a
// local edit. Synchronized templates publish that edit back to the template
// and every profile still linked to it in the same transaction.
func (s *APIKeyProfileTemplateService) reconcileLinkedProfileChanges(
	ctx context.Context,
	existingKey *ent.APIKey,
	nextProfiles *objects.APIKeyProfiles,
) ([]string, error) {
	if existingKey == nil || nextProfiles == nil {
		return nil, nil
	}

	type templateUpdate struct {
		previous  *ent.APIKeyProfileTemplate
		published *objects.APIKeyProfile
	}

	client := s.entFromContext(ctx)
	updates := make(map[int]templateUpdate)
	updateOrder := make([]int, 0)

	for i := range nextProfiles.Profiles {
		profile := &nextProfiles.Profiles[i]
		if profile.TemplateID == nil {
			profile.TemplateName = ""
			profile.TemplateSync = false
			continue
		}

		template, err := client.APIKeyProfileTemplate.Get(ctx, *profile.TemplateID)
		if err != nil {
			if ent.IsNotFound(err) {
				profile.TemplateID = nil
				profile.TemplateName = ""
				profile.TemplateSync = false
				continue
			}
			return nil, fmt.Errorf("failed to get linked template %d: %w", *profile.TemplateID, err)
		}
		if template.ProjectID != existingKey.ProjectID {
			return nil, fmt.Errorf("template and API key must belong to the same project")
		}

		syncEnabled := template.Profile != nil && template.Profile.TemplateSync
		profile.TemplateName = template.Name
		profile.TemplateSync = syncEnabled
		linkedProfile := findLinkedProfileByTemplateID(existingKey.Profiles, *profile.TemplateID)
		if linkedProfile == nil {
			if template.Profile != nil && sameTemplateProfileContents(profile, template.Profile) {
				continue
			}
			profile.TemplateID = nil
			profile.TemplateName = ""
			profile.TemplateSync = false
			continue
		}
		if sameLinkedProfileContents(linkedProfile, profile) {
			continue
		}

		if !syncEnabled {
			profile.TemplateID = nil
			profile.TemplateName = ""
			profile.TemplateSync = false
			continue
		}

		published := profile.Clone()
		published.Name = template.Name
		published.TemplateID = nil
		published.TemplateName = ""
		published.TemplateSync = true

		if current, ok := updates[template.ID]; ok {
			if !sameTemplateProfileContents(current.published, published) {
				return nil, fmt.Errorf("linked profiles for template '%s' contain conflicting edits", template.Name)
			}
			continue
		}

		updates[template.ID] = templateUpdate{previous: template, published: published}
		updateOrder = append(updateOrder, template.ID)
	}

	updatedKeys := make([]string, 0)
	for _, templateID := range updateOrder {
		update := updates[templateID]
		template, err := client.APIKeyProfileTemplate.UpdateOneID(templateID).
			SetProfile(update.published).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to update synchronized template %d: %w", templateID, err)
		}

		keys, err := s.syncLinkedProfiles(ctx, update.previous, template, update.published)
		if err != nil {
			return nil, err
		}
		updatedKeys = append(updatedKeys, keys...)

		for i := range nextProfiles.Profiles {
			current := &nextProfiles.Profiles[i]
			if current.TemplateID == nil || *current.TemplateID != templateID {
				continue
			}

			next := update.published.Clone()
			next.Name = template.Name
			next.TemplateID = lo.ToPtr(templateID)
			next.TemplateName = template.Name
			next.TemplateSync = true
			nextProfiles.Profiles[i] = *next
		}
	}

	return updatedKeys, nil
}

func sameTemplateProfileContents(a, b *objects.APIKeyProfile) bool {
	left := normalizeProfileForComparison(a)
	right := normalizeProfileForComparison(b)
	left.Name = ""
	right.Name = ""
	left.TemplateID = nil
	right.TemplateID = nil
	left.TemplateName = ""
	right.TemplateName = ""
	left.TemplateSync = false
	right.TemplateSync = false

	return reflect.DeepEqual(left, right)
}

func (s *APIKeyProfileTemplateService) syncLinkedProfiles(
	ctx context.Context,
	previousTemplate, template *ent.APIKeyProfileTemplate,
	publishedProfile *objects.APIKeyProfile,
) ([]string, error) {
	client := s.entFromContext(ctx)
	keys, err := client.APIKey.Query().
		Where(apikey.ProjectIDEQ(template.ProjectID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys linked to template: %w", err)
	}

	updatedKeys := make([]string, 0)
	for _, key := range keys {
		if key.Profiles == nil {
			continue
		}

		changed := false
		for i := range key.Profiles.Profiles {
			current := &key.Profiles.Profiles[i]
			isLinked := current.TemplateID != nil && *current.TemplateID == template.ID
			if !isLinked && !isLegacyTemplateMatch(current, previousTemplate) {
				continue
			}

			previousProfileName := current.Name
			next := publishedProfile.Clone()
			next.Name = template.Name
			next.TemplateID = lo.ToPtr(template.ID)
			next.TemplateName = template.Name
			key.Profiles.Profiles[i] = *next
			if key.Profiles.ActiveProfile == previousProfileName {
				key.Profiles.ActiveProfile = template.Name
			}
			changed = true
		}

		if changed {
			if err := validateProfileNames(key.Profiles.Profiles); err != nil {
				return nil, fmt.Errorf("cannot publish template '%s' to API key %d: %w", template.Name, key.ID, err)
			}
			if _, err := client.APIKey.UpdateOneID(key.ID).SetProfiles(key.Profiles).Save(ctx); err != nil {
				return nil, fmt.Errorf("failed to publish template to API key %d: %w", key.ID, err)
			}
			updatedKeys = append(updatedKeys, key.Key)
		}
	}

	return updatedKeys, nil
}

// isLegacyTemplateMatch recognizes profiles created before explicit template
// linkage existed. Only an unchanged profile with the template's original name
// is adopted, so previously customized profiles remain independent.
func isLegacyTemplateMatch(profile *objects.APIKeyProfile, template *ent.APIKeyProfileTemplate) bool {
	if profile == nil || profile.TemplateID != nil || template == nil || template.Profile == nil {
		return false
	}
	if profile.Name != template.Name && profile.Name != template.Profile.Name {
		return false
	}

	left := normalizeProfileForComparison(profile)
	right := normalizeProfileForComparison(template.Profile)
	left.Name = ""
	right.Name = ""
	left.TemplateID = nil
	right.TemplateID = nil
	left.TemplateName = ""
	right.TemplateName = ""
	left.TemplateSync = false
	right.TemplateSync = false

	return reflect.DeepEqual(left, right)
}

func (s *APIKeyProfileTemplateService) detachLinkedProfiles(ctx context.Context, template *ent.APIKeyProfileTemplate) ([]string, error) {
	client := s.entFromContext(ctx)
	keys, err := client.APIKey.Query().
		Where(apikey.ProjectIDEQ(template.ProjectID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys linked to template: %w", err)
	}

	updatedKeys := make([]string, 0)
	for _, key := range keys {
		if key.Profiles == nil {
			continue
		}

		changed := false
		for i := range key.Profiles.Profiles {
			profile := &key.Profiles.Profiles[i]
			if profile.TemplateID != nil && *profile.TemplateID == template.ID {
				profile.TemplateID = nil
				profile.TemplateName = ""
				profile.TemplateSync = false
				changed = true
			}
		}

		if changed {
			if _, err := client.APIKey.UpdateOneID(key.ID).SetProfiles(key.Profiles).Save(ctx); err != nil {
				return nil, fmt.Errorf("failed to detach template from API key %d: %w", key.ID, err)
			}
			updatedKeys = append(updatedKeys, key.Key)
		}
	}

	return updatedKeys, nil
}

func normalizeProfileName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
