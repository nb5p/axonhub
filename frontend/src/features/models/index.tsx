import { useState, useMemo, useCallback, useEffect } from 'react';
import { SortingState } from '@tanstack/react-table';
import { IconPlus, IconSettings, IconAlertCircle } from '@tabler/icons-react';
import { useTranslation } from 'react-i18next';
import { useDebounce } from '@/hooks/use-debounce';
import { usePersistedFilter } from '@/hooks/use-persisted-filter';
import { usePermissions } from '@/hooks/usePermissions';
import { useHorizontalScroll } from '@/hooks/use-horizontal-scroll';
import { Button } from '@/components/ui/button';
import { Header } from '@/components/layout/header';
import { Main } from '@/components/layout/main';
import { PermissionGuard } from '@/components/permission-guard';
import { useOnboardingInfo } from '@/features/system/data/system';
import { createColumns } from './components/models-columns';
import { ModelsDialogs } from './components/models-dialogs';
import { ModelsOnboardingFlow } from './components/models-onboarding-flow';
import { ModelsTable } from './components/models-table';
import ModelsProvider, { useModels } from './context/models-context';
import { useQueryAllModels } from './data/models';
import { useDevelopersData } from './data/providers';

function ModelsContent() {
  useDevelopersData();
  const { t } = useTranslation();
  const { modelPermissions } = usePermissions();
  const [nameFilter, setNameFilter] = usePersistedFilter<string>('models', 'name', '');
  const [sorting, setSorting] = useState<SortingState>(() => {
    const stored = localStorage.getItem('models-table-sorting');
    if (stored) {
      try {
        return JSON.parse(stored);
      } catch {
        return [{ id: 'name', desc: false }];
      }
    }
    return [{ id: 'name', desc: false }];
  });

  useEffect(() => {
    localStorage.setItem('models-table-sorting', JSON.stringify(sorting));
  }, [sorting]);

  const debouncedNameFilter = useDebounce(nameFilter, 300);

  const whereClause = (() => {
    if (debouncedNameFilter) {
      return {
        or: [{ nameContainsFold: debouncedNameFilter }, { modelIDContainsFold: debouncedNameFilter }],
      };
    }
    return undefined;
  })();

  const { data, isLoading } = useQueryAllModels({
    where: whereClause,
  });

  const handleNameFilterChange = useCallback(
    (filter: string) => {
      setNameFilter(filter);
    },
    [setNameFilter]
  );

  const columns = useMemo(() => createColumns(t, modelPermissions.canWrite), [t, modelPermissions.canWrite]);

  return (
    <div className='flex flex-1 flex-col overflow-hidden'>
      <ModelsTable
        data={data?.edges?.map((edge) => edge.node) || []}
        columns={columns}
        loading={isLoading}
        totalCount={data?.totalCount}
        nameFilter={nameFilter}
        sorting={sorting}
        onSortingChange={setSorting}
        onNameFilterChange={handleNameFilterChange}
        canWrite={modelPermissions.canWrite}
      />
    </div>
  );
}

function CreateButton() {
  const { t } = useTranslation();
  const { setOpen } = useModels();

  return (
    <Button onClick={() => setOpen('create')} className='h-8 sm:h-9'>
      <IconPlus className='mr-2 h-4 w-4' />
      {t('models.actions.create')}
    </Button>
  );
}

function BulkAddButton() {
  const { t } = useTranslation();
  const { setOpen } = useModels();

  return (
    <Button variant='outline' onClick={() => setOpen('batchCreate')} className='h-8 sm:h-9'>
      <IconPlus className='mr-2 h-4 w-4' />
      {t('models.actions.bulkAdd')}
    </Button>
  );
}

function SettingsButton() {
  const { t } = useTranslation();
  const { setOpen } = useModels();

  return (
    <Button variant='outline' onClick={() => setOpen('settings')} data-settings-button className='h-8 sm:h-9'>
      <IconSettings className='mr-2 h-4 w-4' />
      {t('models.actions.settings')}
    </Button>
  );
}

function DetectUnassociatedButton() {
  const { t } = useTranslation();
  const { setOpen } = useModels();

  return (
    <Button variant='outline' onClick={() => setOpen('unassociated')} className='h-8 sm:h-9'>
      <IconAlertCircle className='mr-2 h-4 w-4' />
      {t('models.actions.detectUnassociated')}
    </Button>
  );
}

function ActionButtons() {
  const scrollRef = useHorizontalScroll<HTMLDivElement>();
  return (
    <div ref={scrollRef} className='flex min-w-0 gap-2 overflow-x-auto md:overflow-x-visible'>
      <PermissionGuard requiredScope='write_channels'>
        <>
          <DetectUnassociatedButton />
          <SettingsButton />
          <BulkAddButton />
          <CreateButton />
        </>
      </PermissionGuard>
    </div>
  );
}

export default function ModelsManagement() {
  const { t } = useTranslation();
  const { data: onboardingInfo } = useOnboardingInfo();
  const [showOnboarding, setShowOnboarding] = useState(false);

  const shouldShowOnboarding = onboardingInfo && !onboardingInfo.systemModelSetting?.onboarded;

  useEffect(() => {
    if (shouldShowOnboarding) {
      setShowOnboarding(true);
    }
  }, [shouldShowOnboarding]);

  const handleOnboardingComplete = useCallback(() => {
    setShowOnboarding(false);
  }, []);

  return (
    <ModelsProvider>
      <Header fixed className='h-10 py-1 sm:h-16 sm:p-4'>
        <div className='flex min-w-0 w-full flex-1 items-center sm:justify-between'>
          <div className='hidden min-w-0 sm:block'>
            <h2 className='text-xl font-bold tracking-tight'>{t('models.title')}</h2>
            <p className='text-muted-foreground hidden text-sm sm:block'>{t('models.description')}</p>
          </div>
          <ActionButtons />
        </div>
      </Header>

      <Main fixed className='mt-10! py-0 sm:mt-16! sm:py-6'>
        <ModelsContent />
      </Main>
      <ModelsDialogs />
      {showOnboarding && <ModelsOnboardingFlow onComplete={handleOnboardingComplete} />}
    </ModelsProvider>
  );
}
