import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { DEFAULT_LIST_PAGINATION_SETTINGS, useListPaginationSettings, useUpdateListPaginationSettings } from '../data/system';
import type { ListPaginationSettings as ListPaginationSettingsValue } from '../data/system';

type ListKey = keyof ListPaginationSettingsValue;

const LIST_KEYS: ListKey[] = ['channels', 'apiKeys', 'requests'];

export function ListPaginationSettings() {
  const { t } = useTranslation();
  const { data: settings } = useListPaginationSettings();
  const updateSettings = useUpdateListPaginationSettings();
  const [values, setValues] = useState<ListPaginationSettingsValue>(DEFAULT_LIST_PAGINATION_SETTINGS);
  const supported = settings?.supported ?? false;

  useEffect(() => {
    if (settings) {
      setValues(settings);
    }
  }, [settings]);

  const handleChange = async (key: ListKey, enabled: boolean) => {
    const previousValues = values;
    const nextValues = { ...values, [key]: enabled };
    setValues(nextValues);

    try {
      await updateSettings.mutateAsync(nextValues);
    } catch {
      setValues(previousValues);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('system.listPagination.title')}</CardTitle>
        <CardDescription>{t('system.listPagination.description')}</CardDescription>
        {!supported && <p className='text-destructive text-sm'>{t('system.listPagination.backendUnavailable')}</p>}
      </CardHeader>
      <CardContent className='space-y-4'>
        {LIST_KEYS.map((key) => {
          const switchID = `list-pagination-${key}`;
          return (
            <div key={key} className='flex items-center justify-between gap-4'>
              <Label htmlFor={switchID} className='cursor-pointer font-normal'>
                {t(`system.listPagination.${key}`)}
              </Label>
              <Switch
                id={switchID}
                checked={values[key]}
                onCheckedChange={(checked) => handleChange(key, checked)}
                disabled={!supported || updateSettings.isPending}
              />
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
}
