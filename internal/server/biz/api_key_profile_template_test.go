package biz

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/apikey"
	"github.com/looplj/axonhub/internal/ent/apikeyprofiletemplate"
	"github.com/looplj/axonhub/internal/ent/enttest"
	"github.com/looplj/axonhub/internal/ent/project"
	"github.com/looplj/axonhub/internal/ent/user"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/xcache"
)

func setupTestTemplateService(t *testing.T) (*APIKeyProfileTemplateService, *ent.Client) {
	t.Helper()

	client := enttest.NewEntClient(t, "sqlite3", "file:ent?mode=memory&_fk=1")

	svc := NewAPIKeyProfileTemplateService(APIKeyProfileTemplateServiceParams{
		Ent: client,
	})

	return svc, client
}

func TestAPIKeyProfileTemplate(t *testing.T) {
	client := enttest.NewEntClient(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create test user
	hashedPassword, err := HashPassword("test-password")
	require.NoError(t, err)

	testUser, err := client.User.Create().
		SetEmail(fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())).
		SetPassword(hashedPassword).
		SetFirstName("Test").
		SetLastName("User").
		SetStatus(user.StatusActivated).
		Save(ctx)
	require.NoError(t, err)
	_ = testUser

	// Create test project
	projectName := fmt.Sprintf("test-project-%d", time.Now().UnixNano())
	testProject, err := client.Project.Create().
		SetName(projectName).
		SetDescription(projectName).
		SetStatus(project.StatusActive).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	// Test 1: Create ApiKeyProfileTemplate with name, description, project_id, profile JSON
	t.Run("Create ApiKeyProfileTemplate", func(t *testing.T) {
		profile := &objects.APIKeyProfile{
			Name: "test-profile",
			ModelMappings: []objects.ModelMapping{
				{From: "gpt-4", To: "gpt-4-turbo"},
			},
		}

		template, err := client.APIKeyProfileTemplate.Create().
			SetName("my-template").
			SetDescription("A test template").
			SetProject(testProject).
			SetProfile(profile).
			Save(ctx)
		require.NoError(t, err)
		require.NotNil(t, template)
		require.Equal(t, "my-template", template.Name)
		require.Equal(t, "A test template", template.Description)
		require.Equal(t, testProject.ID, template.ProjectID)

		// Verify profile was stored correctly
		require.NotNil(t, template.Profile)
		require.Equal(t, "test-profile", template.Profile.Name)
	})

	// Test 2: Query back the created template
	t.Run("Query ApiKeyProfileTemplate", func(t *testing.T) {
		template, err := client.APIKeyProfileTemplate.Query().
			Where(apikeyprofiletemplate.Name("my-template")).
			Only(ctx)
		require.NoError(t, err)
		require.NotNil(t, template)
		require.Equal(t, "A test template", template.Description)
	})

	// Test 3: Unique index - duplicate name+project_id should fail
	t.Run("Unique index on name+project_id", func(t *testing.T) {
		// Try to create duplicate
		_, err := client.APIKeyProfileTemplate.Create().
			SetName("my-template").
			SetDescription("Duplicate").
			SetProject(testProject).
			SetProfile(&objects.APIKeyProfile{Name: "dup"}).
			Save(ctx)
		require.Error(t, err)
	})

	// Test 4: Same name in different project should succeed
	t.Run("Same name in different project", func(t *testing.T) {
		// Create another project
		project2, err := client.Project.Create().
			SetName(fmt.Sprintf("test-project-2-%d", time.Now().UnixNano())).
			SetDescription("project 2").
			SetStatus(project.StatusActive).
			Save(ctx)
		require.NoError(t, err)

		template2, err := client.APIKeyProfileTemplate.Create().
			SetName("my-template").
			SetDescription("Different project").
			SetProject(project2).
			SetProfile(&objects.APIKeyProfile{Name: "profile2"}).
			Save(ctx)
		require.NoError(t, err)
		require.NotNil(t, template2)
		require.Equal(t, project2.ID, template2.ProjectID)
	})

	// Test 5: Query templates via project edge
	t.Run("Query via project edge", func(t *testing.T) {
		templates, err := testProject.QueryAPIKeyProfileTemplates().All(ctx)
		require.NoError(t, err)
		require.Len(t, templates, 1)
		require.Equal(t, "my-template", templates[0].Name)
	})
}

// TestLoadTemplate_HappyPath tests loading a template into an API key with different profile names.
// Profile appended, existing profiles unchanged, active profile unchanged.
func TestLoadTemplate_HappyPath(t *testing.T) {
	svc, client := setupTestTemplateService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create test user
	hashedPassword, err := HashPassword("test-password")
	require.NoError(t, err)

	testUser, err := client.User.Create().
		SetEmail(fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())).
		SetPassword(hashedPassword).
		SetFirstName("Test").
		SetLastName("User").
		SetStatus(user.StatusActivated).
		Save(ctx)
	require.NoError(t, err)

	// Create test project
	testProject, err := client.Project.Create().
		SetName(fmt.Sprintf("test-project-%d", time.Now().UnixNano())).
		SetDescription("test").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	// Create API key with existing profiles
	existingProfiles := &objects.APIKeyProfiles{
		ActiveProfile: "Default",
		Profiles: []objects.APIKeyProfile{
			{
				Name: "Default",
				ModelMappings: []objects.ModelMapping{
					{From: "gpt-4", To: "gpt-4-turbo"},
				},
			},
		},
	}

	apiKey, err := client.APIKey.Create().
		SetName("test-api-key").
		SetKey(fmt.Sprintf("ah-test-%d", time.Now().UnixNano())).
		SetUserID(testUser.ID).
		SetProjectID(testProject.ID).
		SetType(apikey.TypeUser).
		SetProfiles(existingProfiles).
		Save(ctx)
	require.NoError(t, err)

	// Create template with a different profile name
	templateProfile := &objects.APIKeyProfile{
		Name: "Production",
		ModelMappings: []objects.ModelMapping{
			{From: "claude-3", To: "claude-3-opus"},
		},
	}

	template, err := client.APIKeyProfileTemplate.Create().
		SetName("prod-template").
		SetDescription("Production template").
		SetProject(testProject).
		SetProfile(templateProfile).
		Save(ctx)
	require.NoError(t, err)

	// Load template into API key
	updatedKey, err := svc.LoadTemplate(ctx, template.ID, apiKey.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedKey)

	// Verify profiles
	require.NotNil(t, updatedKey.Profiles)
	require.Equal(t, "Default", updatedKey.Profiles.ActiveProfile, "active profile should not change")
	require.Len(t, updatedKey.Profiles.Profiles, 2, "should have 2 profiles (original + loaded)")

	// Verify original profile is unchanged
	require.Equal(t, "Default", updatedKey.Profiles.Profiles[0].Name)
	require.Len(t, updatedKey.Profiles.Profiles[0].ModelMappings, 1)
	require.Equal(t, "gpt-4", updatedKey.Profiles.Profiles[0].ModelMappings[0].From)

	// Verify loaded profile is appended
	require.Equal(t, "Production", updatedKey.Profiles.Profiles[1].Name)
	require.NotNil(t, updatedKey.Profiles.Profiles[1].TemplateID)
	require.Equal(t, template.ID, *updatedKey.Profiles.Profiles[1].TemplateID)
	require.Equal(t, template.Name, updatedKey.Profiles.Profiles[1].TemplateName)
	require.Len(t, updatedKey.Profiles.Profiles[1].ModelMappings, 1)
	require.Equal(t, "claude-3", updatedKey.Profiles.Profiles[1].ModelMappings[0].From)
}

func TestUpdateTemplatePublishesToLinkedAndUnchangedLegacyProfiles(t *testing.T) {
	svc, client := setupTestTemplateService(t)
	defer client.Close()

	ctx := authz.WithTestBypass(ent.NewContext(context.Background(), client))
	projectEntity, err := client.Project.Create().
		SetName(fmt.Sprintf("publish-project-%d", time.Now().UnixNano())).
		SetDescription("publish test").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	template, err := client.APIKeyProfileTemplate.Create().
		SetName("Production").
		SetProject(projectEntity).
		SetProfile(&objects.APIKeyProfile{
			Name:          "Production",
			ModelMappings: []objects.ModelMapping{{From: "claude", To: "old-model"}},
		}).
		Save(ctx)
	require.NoError(t, err)

	templateID := template.ID
	createKey := func(name string, profile objects.APIKeyProfile) *ent.APIKey {
		key, createErr := client.APIKey.Create().
			SetName(name).
			SetKey(fmt.Sprintf("ah-%s-%d", name, time.Now().UnixNano())).
			SetProjectID(projectEntity.ID).
			SetType(apikey.TypeUser).
			SetProfiles(&objects.APIKeyProfiles{ActiveProfile: profile.Name, Profiles: []objects.APIKeyProfile{profile}}).
			Save(ctx)
		require.NoError(t, createErr)
		return key
	}

	linkedKey := createKey("linked", objects.APIKeyProfile{
		Name:          "Local alias",
		TemplateID:    &templateID,
		TemplateName:  template.Name,
		ModelMappings: []objects.ModelMapping{{From: "claude", To: "old-model"}},
	})
	legacyKey := createKey("legacy", objects.APIKeyProfile{
		Name:          "Production",
		ModelMappings: []objects.ModelMapping{{From: "claude", To: "old-model"}},
	})
	detachedKey := createKey("detached", objects.APIKeyProfile{
		Name:          "Production",
		ModelMappings: []objects.ModelMapping{{From: "claude", To: "custom-model"}},
	})

	count, err := svc.CountLinkedProfiles(ctx, template)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	_, err = svc.UpdateTemplate(ctx, template.ID, ent.UpdateAPIKeyProfileTemplateInput{}, &objects.APIKeyProfile{
		Name:          "Production",
		ModelMappings: []objects.ModelMapping{{From: "claude", To: "new-model"}},
	})
	require.NoError(t, err)

	for _, key := range []*ent.APIKey{linkedKey, legacyKey} {
		updated, getErr := client.APIKey.Get(ctx, key.ID)
		require.NoError(t, getErr)
		require.Equal(t, "new-model", updated.Profiles.Profiles[0].ModelMappings[0].To)
		require.NotNil(t, updated.Profiles.Profiles[0].TemplateID)
		require.Equal(t, template.ID, *updated.Profiles.Profiles[0].TemplateID)
	}

	updatedLinked, err := client.APIKey.Get(ctx, linkedKey.ID)
	require.NoError(t, err)
	require.Equal(t, "Local alias", updatedLinked.Profiles.Profiles[0].Name)

	updatedDetached, err := client.APIKey.Get(ctx, detachedKey.ID)
	require.NoError(t, err)
	require.Equal(t, "custom-model", updatedDetached.Profiles.Profiles[0].ModelMappings[0].To)
	require.Nil(t, updatedDetached.Profiles.Profiles[0].TemplateID)

	_, err = svc.DeleteTemplate(ctx, template.ID)
	require.NoError(t, err)
	for _, key := range []*ent.APIKey{linkedKey, legacyKey} {
		updated, getErr := client.APIKey.Get(ctx, key.ID)
		require.NoError(t, getErr)
		require.Nil(t, updated.Profiles.Profiles[0].TemplateID)
		require.Empty(t, updated.Profiles.Profiles[0].TemplateName)
		require.Equal(t, "new-model", updated.Profiles.Profiles[0].ModelMappings[0].To)
	}
}

func TestDetachModifiedTemplateProfiles(t *testing.T) {
	templateID := 7
	existing := &objects.APIKeyProfiles{Profiles: []objects.APIKeyProfile{
		{
			Name:          "linked",
			TemplateID:    &templateID,
			TemplateName:  "Production",
			ModelMappings: []objects.ModelMapping{{From: "claude", To: "old-model"}},
		},
	}}

	unchanged := &objects.APIKeyProfiles{Profiles: []objects.APIKeyProfile{
		{
			Name:          "linked",
			TemplateID:    &templateID,
			TemplateName:  "tampered name",
			ModelMappings: []objects.ModelMapping{{From: "claude", To: "old-model"}},
		},
	}}
	detachModifiedTemplateProfiles(existing, unchanged)
	require.NotNil(t, unchanged.Profiles[0].TemplateID)
	require.Equal(t, "Production", unchanged.Profiles[0].TemplateName)

	modified := &objects.APIKeyProfiles{Profiles: []objects.APIKeyProfile{
		{
			Name:          "linked",
			TemplateID:    &templateID,
			TemplateName:  "Production",
			ModelMappings: []objects.ModelMapping{{From: "claude", To: "custom-model"}},
		},
	}}
	detachModifiedTemplateProfiles(existing, modified)
	require.Nil(t, modified.Profiles[0].TemplateID)
	require.Empty(t, modified.Profiles[0].TemplateName)
}

func TestActivateTemplateProfile(t *testing.T) {
	svc, client := setupTestTemplateService(t)
	defer client.Close()

	ctx := authz.WithTestBypass(ent.NewContext(context.Background(), client))
	projectEntity, err := client.Project.Create().
		SetName(fmt.Sprintf("activate-template-project-%d", time.Now().UnixNano())).
		SetDescription("activate template profile test").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	privacyTemplate, err := client.APIKeyProfileTemplate.Create().
		SetName("Privacy").
		SetProject(projectEntity).
		SetProfile(&objects.APIKeyProfile{Name: "Privacy", TemplateSync: true}).
		Save(ctx)
	require.NoError(t, err)
	codeTemplate, err := client.APIKeyProfileTemplate.Create().
		SetName("Code").
		SetProject(projectEntity).
		SetProfile(&objects.APIKeyProfile{Name: "Code", TemplateSync: true}).
		Save(ctx)
	require.NoError(t, err)

	privacyTemplateID := privacyTemplate.ID
	codeTemplateID := codeTemplate.ID
	key, err := client.APIKey.Create().
		SetName("switchable-key").
		SetKey(fmt.Sprintf("ah-switchable-%d", time.Now().UnixNano())).
		SetProjectID(projectEntity.ID).
		SetType(apikey.TypeUser).
		SetProfiles(&objects.APIKeyProfiles{
			ActiveProfile: "Privacy package",
			Profiles: []objects.APIKeyProfile{
				{Name: "Privacy package", TemplateID: &privacyTemplateID, TemplateName: privacyTemplate.Name, TemplateSync: true},
				{Name: "Code package", TemplateID: &codeTemplateID, TemplateName: codeTemplate.Name, TemplateSync: true},
			},
		}).
		Save(ctx)
	require.NoError(t, err)

	updated, err := svc.ActivateTemplateProfile(ctx, key.ID, codeTemplate.ID)
	require.NoError(t, err)
	require.Equal(t, "Code package", updated.Profiles.ActiveProfile)
	require.Len(t, updated.Profiles.Profiles, 2)

	unlinkedTemplate, err := client.APIKeyProfileTemplate.Create().
		SetName("Unlinked").
		SetProject(projectEntity).
		SetProfile(&objects.APIKeyProfile{Name: "Unlinked", TemplateSync: true}).
		Save(ctx)
	require.NoError(t, err)
	_, err = svc.ActivateTemplateProfile(ctx, key.ID, unlinkedTemplate.ID)
	require.ErrorContains(t, err, "has no linked profile")
}

func TestSynchronizedTemplatePublishesAPIKeyProfileEdits(t *testing.T) {
	apiKeyService, client := setupTestAPIKeyService(t, xcache.Config{Mode: xcache.ModeMemory})
	defer apiKeyService.Stop()
	defer client.Close()

	ctx := authz.WithTestBypass(ent.NewContext(context.Background(), client))
	templateService := NewAPIKeyProfileTemplateService(APIKeyProfileTemplateServiceParams{Ent: client})
	projectEntity, err := client.Project.Create().
		SetName(fmt.Sprintf("sync-project-%d", time.Now().UnixNano())).
		SetDescription("synchronized template test").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	template, err := client.APIKeyProfileTemplate.Create().
		SetName("Synchronized").
		SetProject(projectEntity).
		SetProfile(&objects.APIKeyProfile{
			Name:          "Synchronized",
			TemplateSync:  true,
			ModelMappings: []objects.ModelMapping{{From: "shared", To: "old-model"}},
		}).
		Save(ctx)
	require.NoError(t, err)

	templateID := template.ID
	createLinkedKey := func(name, profileName string) *ent.APIKey {
		key, createErr := client.APIKey.Create().
			SetName(name).
			SetKey(fmt.Sprintf("ah-%s-%d", name, time.Now().UnixNano())).
			SetProjectID(projectEntity.ID).
			SetType(apikey.TypeUser).
			SetProfiles(&objects.APIKeyProfiles{
				ActiveProfile: profileName,
				Profiles: []objects.APIKeyProfile{{
					Name:          profileName,
					TemplateID:    &templateID,
					TemplateName:  template.Name,
					TemplateSync:  true,
					ModelMappings: []objects.ModelMapping{{From: "shared", To: "old-model"}},
				}},
			}).
			Save(ctx)
		require.NoError(t, createErr)
		return key
	}

	firstKey := createLinkedKey("first", "First alias")
	secondKey := createLinkedKey("second", "Second alias")
	newlyLinkedKey, err := client.APIKey.Create().
		SetName("newly-linked").
		SetKey(fmt.Sprintf("ah-newly-linked-%d", time.Now().UnixNano())).
		SetProjectID(projectEntity.ID).
		SetType(apikey.TypeUser).
		SetProfiles(&objects.APIKeyProfiles{
			ActiveProfile: "Local profile",
			Profiles: []objects.APIKeyProfile{{
				Name:          "Local profile",
				ModelMappings: []objects.ModelMapping{{From: "shared", To: "old-model"}},
			}},
		}).
		Save(ctx)
	require.NoError(t, err)

	attachedKey, err := apiKeyService.UpdateAPIKeyProfiles(ctx, newlyLinkedKey.ID, objects.APIKeyProfiles{
		ActiveProfile: "Local profile",
		Profiles: []objects.APIKeyProfile{{
			Name:          "Local profile",
			TemplateID:    &templateID,
			TemplateName:  template.Name,
			TemplateSync:  true,
			ModelMappings: []objects.ModelMapping{{From: "shared", To: "old-model"}},
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, attachedKey.Profiles.Profiles[0].TemplateID)
	require.True(t, attachedKey.Profiles.Profiles[0].TemplateSync)

	cachedSecond, err := apiKeyService.GetAPIKey(ctx, secondKey.Key)
	require.NoError(t, err)
	require.Equal(t, "old-model", cachedSecond.Profiles.Profiles[0].ModelMappings[0].To)

	updatedFirst, err := apiKeyService.UpdateAPIKeyProfiles(ctx, firstKey.ID, objects.APIKeyProfiles{
		ActiveProfile: "First alias",
		Profiles: []objects.APIKeyProfile{{
			Name:          "First alias",
			TemplateID:    &templateID,
			TemplateName:  template.Name,
			TemplateSync:  false,
			ModelMappings: []objects.ModelMapping{{From: "shared", To: "new-model"}},
		}},
	})
	require.NoError(t, err)
	require.True(t, updatedFirst.Profiles.Profiles[0].TemplateSync)
	require.NotNil(t, updatedFirst.Profiles.Profiles[0].TemplateID)

	updatedTemplate, err := client.APIKeyProfileTemplate.Get(ctx, template.ID)
	require.NoError(t, err)
	require.True(t, updatedTemplate.Profile.TemplateSync)
	require.Equal(t, "new-model", updatedTemplate.Profile.ModelMappings[0].To)

	updatedSecond, err := apiKeyService.GetAPIKey(ctx, secondKey.Key)
	require.NoError(t, err)
	require.Equal(t, "Second alias", updatedSecond.Profiles.Profiles[0].Name)
	require.True(t, updatedSecond.Profiles.Profiles[0].TemplateSync)
	require.Equal(t, "new-model", updatedSecond.Profiles.Profiles[0].ModelMappings[0].To)

	updatedNewlyLinked, err := client.APIKey.Get(ctx, newlyLinkedKey.ID)
	require.NoError(t, err)
	require.Equal(t, "new-model", updatedNewlyLinked.Profiles.Profiles[0].ModelMappings[0].To)

	_, err = templateService.UpdateTemplate(ctx, template.ID, ent.UpdateAPIKeyProfileTemplateInput{}, &objects.APIKeyProfile{
		Name:          template.Name,
		TemplateSync:  false,
		ModelMappings: []objects.ModelMapping{{From: "shared", To: "final-model"}},
	})
	require.NoError(t, err)

	updatedTemplate, err = client.APIKeyProfileTemplate.Get(ctx, template.ID)
	require.NoError(t, err)
	require.True(t, updatedTemplate.Profile.TemplateSync, "synchronization cannot be disabled once enabled")
}

// TestLoadTemplate_NameConflict tests loading a template where profile name already exists.
// Auto-rename with suffix " (1)".
func TestLoadTemplate_NameConflict(t *testing.T) {
	svc, client := setupTestTemplateService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create test user
	hashedPassword, err := HashPassword("test-password")
	require.NoError(t, err)

	testUser, err := client.User.Create().
		SetEmail(fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())).
		SetPassword(hashedPassword).
		SetFirstName("Test").
		SetLastName("User").
		SetStatus(user.StatusActivated).
		Save(ctx)
	require.NoError(t, err)

	// Create test project
	testProject, err := client.Project.Create().
		SetName(fmt.Sprintf("test-project-%d", time.Now().UnixNano())).
		SetDescription("test").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	// Create API key with a profile named "Production"
	existingProfiles := &objects.APIKeyProfiles{
		ActiveProfile: "Production",
		Profiles: []objects.APIKeyProfile{
			{
				Name: "Production",
				ModelMappings: []objects.ModelMapping{
					{From: "gpt-4", To: "gpt-4-turbo"},
				},
			},
		},
	}

	apiKey, err := client.APIKey.Create().
		SetName("test-api-key").
		SetKey(fmt.Sprintf("ah-test-%d", time.Now().UnixNano())).
		SetUserID(testUser.ID).
		SetProjectID(testProject.ID).
		SetType(apikey.TypeUser).
		SetProfiles(existingProfiles).
		Save(ctx)
	require.NoError(t, err)

	// Create template also named "Production"
	templateProfile := &objects.APIKeyProfile{
		Name: "Production",
		ModelMappings: []objects.ModelMapping{
			{From: "claude-3", To: "claude-3-opus"},
		},
	}

	template, err := client.APIKeyProfileTemplate.Create().
		SetName("prod-template").
		SetDescription("Production template").
		SetProject(testProject).
		SetProfile(templateProfile).
		Save(ctx)
	require.NoError(t, err)

	// Load template into API key
	updatedKey, err := svc.LoadTemplate(ctx, template.ID, apiKey.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedKey)

	// Verify profiles
	require.NotNil(t, updatedKey.Profiles)
	require.Equal(t, "Production", updatedKey.Profiles.ActiveProfile, "active profile should not change")
	require.Len(t, updatedKey.Profiles.Profiles, 2, "should have 2 profiles")

	// Original profile unchanged
	require.Equal(t, "Production", updatedKey.Profiles.Profiles[0].Name)

	// Loaded profile should be auto-renamed to "Production (1)"
	require.Equal(t, "Production (1)", updatedKey.Profiles.Profiles[1].Name)
	require.Len(t, updatedKey.Profiles.Profiles[1].ModelMappings, 1)
	require.Equal(t, "claude-3", updatedKey.Profiles.Profiles[1].ModelMappings[0].From)
}

// TestLoadTemplate_MultipleConflicts tests loading a template when multiple name conflicts exist.
// Key has "Production", "Production (1)", template is "Production". Should become "Production (2)".
func TestLoadTemplate_MultipleConflicts(t *testing.T) {
	svc, client := setupTestTemplateService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create test user
	hashedPassword, err := HashPassword("test-password")
	require.NoError(t, err)

	testUser, err := client.User.Create().
		SetEmail(fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())).
		SetPassword(hashedPassword).
		SetFirstName("Test").
		SetLastName("User").
		SetStatus(user.StatusActivated).
		Save(ctx)
	require.NoError(t, err)

	// Create test project
	testProject, err := client.Project.Create().
		SetName(fmt.Sprintf("test-project-%d", time.Now().UnixNano())).
		SetDescription("test").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	// Create API key with "Production" and "Production (1)"
	existingProfiles := &objects.APIKeyProfiles{
		ActiveProfile: "Production",
		Profiles: []objects.APIKeyProfile{
			{
				Name: "Production",
				ModelMappings: []objects.ModelMapping{
					{From: "gpt-4", To: "gpt-4-turbo"},
				},
			},
			{
				Name: "Production (1)",
				ModelMappings: []objects.ModelMapping{
					{From: "gpt-3.5", To: "gpt-3.5-turbo"},
				},
			},
		},
	}

	apiKey, err := client.APIKey.Create().
		SetName("test-api-key").
		SetKey(fmt.Sprintf("ah-test-%d", time.Now().UnixNano())).
		SetUserID(testUser.ID).
		SetProjectID(testProject.ID).
		SetType(apikey.TypeUser).
		SetProfiles(existingProfiles).
		Save(ctx)
	require.NoError(t, err)

	// Create template named "Production"
	templateProfile := &objects.APIKeyProfile{
		Name: "Production",
		ModelMappings: []objects.ModelMapping{
			{From: "claude-3", To: "claude-3-opus"},
		},
	}

	template, err := client.APIKeyProfileTemplate.Create().
		SetName("prod-template").
		SetDescription("Production template").
		SetProject(testProject).
		SetProfile(templateProfile).
		Save(ctx)
	require.NoError(t, err)

	// Load template into API key
	updatedKey, err := svc.LoadTemplate(ctx, template.ID, apiKey.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedKey)

	// Verify profiles
	require.NotNil(t, updatedKey.Profiles)
	require.Len(t, updatedKey.Profiles.Profiles, 3, "should have 3 profiles")

	// Original profiles unchanged
	require.Equal(t, "Production", updatedKey.Profiles.Profiles[0].Name)
	require.Equal(t, "Production (1)", updatedKey.Profiles.Profiles[1].Name)

	// Loaded profile should be auto-renamed to "Production (2)"
	require.Equal(t, "Production (2)", updatedKey.Profiles.Profiles[2].Name)
}

// TestLoadTemplate_TemplateNotFound tests that loading a non-existent template returns an error.
func TestLoadTemplate_TemplateNotFound(t *testing.T) {
	svc, client := setupTestTemplateService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create test user
	hashedPassword, err := HashPassword("test-password")
	require.NoError(t, err)

	testUser, err := client.User.Create().
		SetEmail(fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())).
		SetPassword(hashedPassword).
		SetFirstName("Test").
		SetLastName("User").
		SetStatus(user.StatusActivated).
		Save(ctx)
	require.NoError(t, err)

	// Create test project
	testProject, err := client.Project.Create().
		SetName(fmt.Sprintf("test-project-%d", time.Now().UnixNano())).
		SetDescription("test").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	// Create API key
	apiKey, err := client.APIKey.Create().
		SetName("test-api-key").
		SetKey(fmt.Sprintf("ah-test-%d", time.Now().UnixNano())).
		SetUserID(testUser.ID).
		SetProjectID(testProject.ID).
		SetType(apikey.TypeUser).
		Save(ctx)
	require.NoError(t, err)

	// Try to load non-existent template
	_, err = svc.LoadTemplate(ctx, 99999, apiKey.ID)
	require.Error(t, err)
}

// TestLoadTemplate_APIKeyNotFound tests that loading into a non-existent API key returns an error.
func TestLoadTemplate_APIKeyNotFound(t *testing.T) {
	svc, client := setupTestTemplateService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create test project
	testProject, err := client.Project.Create().
		SetName(fmt.Sprintf("test-project-%d", time.Now().UnixNano())).
		SetDescription("test").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	// Create template
	templateProfile := &objects.APIKeyProfile{
		Name: "Production",
	}

	template, err := client.APIKeyProfileTemplate.Create().
		SetName("prod-template").
		SetDescription("Production template").
		SetProject(testProject).
		SetProfile(templateProfile).
		Save(ctx)
	require.NoError(t, err)

	// Try to load into non-existent API key
	_, err = svc.LoadTemplate(ctx, template.ID, 99999)
	require.Error(t, err)
}

// TestLoadTemplate_DifferentProject tests that loading a template from a different project fails.
func TestLoadTemplate_DifferentProject(t *testing.T) {
	svc, client := setupTestTemplateService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create test user
	hashedPassword, err := HashPassword("test-password")
	require.NoError(t, err)

	testUser, err := client.User.Create().
		SetEmail(fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())).
		SetPassword(hashedPassword).
		SetFirstName("Test").
		SetLastName("User").
		SetStatus(user.StatusActivated).
		Save(ctx)
	require.NoError(t, err)

	// Create project 1
	project1, err := client.Project.Create().
		SetName(fmt.Sprintf("project-1-%d", time.Now().UnixNano())).
		SetDescription("project 1").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	// Create project 2
	project2, err := client.Project.Create().
		SetName(fmt.Sprintf("project-2-%d", time.Now().UnixNano())).
		SetDescription("project 2").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	// Create API key in project 1
	apiKey, err := client.APIKey.Create().
		SetName("test-api-key").
		SetKey(fmt.Sprintf("ah-test-%d", time.Now().UnixNano())).
		SetUserID(testUser.ID).
		SetProjectID(project1.ID).
		SetType(apikey.TypeUser).
		Save(ctx)
	require.NoError(t, err)

	// Create template in project 2
	templateProfile := &objects.APIKeyProfile{
		Name: "Production",
	}

	template, err := client.APIKeyProfileTemplate.Create().
		SetName("prod-template").
		SetDescription("Production template").
		SetProject(project2).
		SetProfile(templateProfile).
		Save(ctx)
	require.NoError(t, err)

	// Try to load template from project 2 into API key in project 1
	_, err = svc.LoadTemplate(ctx, template.ID, apiKey.ID)
	require.Error(t, err)
}
