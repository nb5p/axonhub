import { useApiKeysContext } from '../context/apikeys-context';
import { useApiKey, useUpdateApiKeyProfiles } from '../data/apikeys';
import { type UpdateApiKeyProfilesInput } from '../data/schema';
import { ApiKeysArchiveDialog } from './apikeys-archive-dialog';
import { ApiKeysBulkArchiveDialog } from './apikeys-bulk-archive-dialog';
import { ApiKeysBulkDisableDialog } from './apikeys-bulk-disable-dialog';
import { ApiKeysBulkEnableDialog } from './apikeys-bulk-enable-dialog';
import { ApiKeysCreateDialog } from './apikeys-create-dialog';
import { ApiKeysEditDialog } from './apikeys-edit-dialog';
import { ApiKeysProfileTemplatesDialog } from './apikeys-profile-templates-dialog';
import { ApiKeyProfilesDialog } from './apikeys-profiles-dialog';
// import { ApiKeysDeleteDialog } from './apikeys-delete-dialog'
import { ApiKeysRotateDialog } from './apikeys-rotate-dialog';
import { ApiKeysStatusDialog } from './apikeys-status-dialog';
import { ApiKeysViewDialog } from './apikeys-view-dialog';

export function ApiKeysDialogs() {
  return (
    <>
      <ApiKeysCreateDialog />
      <ApiKeysEditDialog />
      {/* <ApiKeysDeleteDialog /> */}
      <ApiKeysStatusDialog />
      <ApiKeysViewDialog />
      <ApiKeysArchiveDialog />
      <ApiKeysProfilesDialogWrapper />
      <ApiKeysProfileTemplatesDialogWrapper />
      <ApiKeysBulkDisableDialog />
      <ApiKeysBulkArchiveDialog />
      <ApiKeysBulkEnableDialog />
      <ApiKeysRotateDialog />
    </>
  );
}

function ApiKeysProfilesDialogWrapper() {
  const { isDialogOpen, closeDialog, selectedApiKey } = useApiKeysContext();
  const updateProfilesMutation = useUpdateApiKeyProfiles();
  const { data: apiKeyDetail, isLoading: apiKeyDetailLoading } = useApiKey(selectedApiKey?.id || '');

  const handleSubmit = (data: UpdateApiKeyProfilesInput) => {
    if (!selectedApiKey?.id) return;

    updateProfilesMutation.mutate(
      { id: selectedApiKey.id, input: data },
      {
        onSuccess: () => {
          closeDialog('profiles');
        },
      }
    );
  };

  const handleClear = () => {
    handleSubmit({ activeProfile: '', profiles: [] });
  };

  return (
    <ApiKeyProfilesDialog
      open={isDialogOpen.profiles}
      onOpenChange={(open) => !open && closeDialog('profiles')}
      onSubmit={handleSubmit}
      onClear={handleClear}
      canClear={(apiKeyDetail?.profiles?.profiles?.length ?? 0) > 0}
      loading={updateProfilesMutation.isPending}
      initialDataLoading={apiKeyDetailLoading}
      initialData={
        apiKeyDetail?.profiles
          ? {
              activeProfile: apiKeyDetail.profiles.activeProfile || apiKeyDetail.profiles.profiles?.[0]?.name || 'Default',
              profiles: apiKeyDetail.profiles.profiles || [],
            }
          : undefined
      }
    />
  );
}

function ApiKeysProfileTemplatesDialogWrapper() {
  const { isDialogOpen, closeDialog } = useApiKeysContext();

  return (
    <ApiKeysProfileTemplatesDialog open={isDialogOpen.profileTemplates} onOpenChange={(open) => !open && closeDialog('profileTemplates')} />
  );
}
