'use client';

import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { DEFAULT_TEST_REQUEST_LIST_SETTINGS, useTestRequestListSettings, useUpdateTestRequestListSettings } from '../data/system';

export function TestRequestListSettings() {
  const { t } = useTranslation();
  const { data: settings } = useTestRequestListSettings();
  const updateSettings = useUpdateTestRequestListSettings();
  const [showInRequestList, setShowInRequestList] = useState(DEFAULT_TEST_REQUEST_LIST_SETTINGS.showInRequestList);
  const supported = settings?.supported ?? false;

  useEffect(() => {
    if (settings) {
      setShowInRequestList(settings.showInRequestList);
    }
  }, [settings]);

  const handleChange = async (enabled: boolean) => {
    const previousValue = showInRequestList;
    setShowInRequestList(enabled);

    try {
      await updateSettings.mutateAsync({ showInRequestList: enabled });
    } catch {
      setShowInRequestList(previousValue);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('system.testRequestList.title')}</CardTitle>
        <CardDescription>{t('system.testRequestList.description')}</CardDescription>
        {!supported && <p className='text-destructive text-sm'>{t('system.testRequestList.backendUnavailable')}</p>}
      </CardHeader>
      <CardContent>
        <div className='flex items-center justify-between gap-4'>
          <Label htmlFor='test-request-list' className='cursor-pointer font-normal'>
            {t('system.testRequestList.showInRequestList')}
          </Label>
          <Switch
            id='test-request-list'
            checked={showInRequestList}
            onCheckedChange={handleChange}
            disabled={!supported || updateSettings.isPending}
          />
        </div>
      </CardContent>
    </Card>
  );
}
