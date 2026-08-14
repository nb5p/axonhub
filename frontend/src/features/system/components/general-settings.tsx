'use client';

import React, { useState, useEffect } from 'react';
import { Check, ChevronsUpDown, Loader2, Save } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { Label } from '@/components/ui/label';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Switch } from '@/components/ui/switch';
import { AutoCompleteSelect } from '@/components/auto-complete-select';
import { useSystemContext } from '../context/system-context';
import { currencyCodes } from '../data/currencies';
import {
  useGeneralSettings,
  useUpdateGeneralSettings,
  useUserAgentPassThroughSettings,
  useUpdateUserAgentPassThroughSettings,
  usePassThroughSettings,
  useUpdatePassThroughSettings,
} from '../data/system';
import { GMTTimeZoneOptions } from '../data/timezones';
import { ListPaginationSettings } from './list-pagination-settings';
import { SidebarNavigationSettings } from './sidebar-navigation-settings';

export function GeneralSettings() {
  const { t } = useTranslation();
  const { data: settings, isLoading: isLoadingSettings } = useGeneralSettings();
  const updateSettings = useUpdateGeneralSettings();
  const { isLoading, setIsLoading } = useSystemContext();

  // User-Agent Pass-Through settings
  const { data: uaSettings, isLoading: isLoadingUASettings } = useUserAgentPassThroughSettings();
  const updateUASettings = useUpdateUserAgentPassThroughSettings();
  const [uaPassThroughEnabled, setUaPassThroughEnabled] = useState(false);

  // Pass-Through (request/response body) settings
  const { data: ptSettings, isLoading: isLoadingPTSettings } = usePassThroughSettings();
  const updatePTSettings = useUpdatePassThroughSettings();
  const [passThroughEnabled, setPassThroughEnabled] = useState(false);
  const [preferPassThroughEnabled, setPreferPassThroughEnabled] = useState(false);
  const [preferPassThroughExceptionsEnabled, setPreferPassThroughExceptionsEnabled] = useState(false);
  const [preferPassThroughExceptionConversions, setPreferPassThroughExceptionConversions] = useState<string[]>([]);

  const [currencyCode, setCurrencyCode] = useState('USD');
  const [timezone, setTimezone] = useState('UTC');

  const currencyItems = React.useMemo(
    () =>
      currencyCodes.map((code) => ({
        value: code,
        label: t(`currencies.${code}`),
      })),
    [t]
  );

  const timezoneItems = React.useMemo(() => GMTTimeZoneOptions, []);

  // Update local state when settings are loaded
  useEffect(() => {
    if (settings) {
      setCurrencyCode(settings.currencyCode || 'USD');
      setTimezone(settings.timezone || 'UTC');
    }
  }, [settings]);

  // Update UA pass-through state when loaded
  useEffect(() => {
    if (uaSettings) {
      setUaPassThroughEnabled(uaSettings.enabled);
    }
  }, [uaSettings]);

  // Update pass-through state when loaded
  useEffect(() => {
    if (ptSettings) {
      setPassThroughEnabled(ptSettings.enabled);
      setPreferPassThroughEnabled(ptSettings.preferPassThrough);
      setPreferPassThroughExceptionsEnabled(ptSettings.preferPassThroughExceptionsEnabled);
      setPreferPassThroughExceptionConversions(ptSettings.preferPassThroughExceptionConversions);
    }
  }, [ptSettings]);

  const handleSave = async () => {
    setIsLoading(true);
    try {
      await updateSettings.mutateAsync({
        currencyCode: currencyCode.trim(),
        timezone: timezone.trim(),
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleUAPassThroughChange = async (enabled: boolean) => {
    const previousValue = uaPassThroughEnabled;
    setUaPassThroughEnabled(enabled);
    try {
      await updateUASettings.mutateAsync({ enabled });
    } catch {
      // Revert state on error
      setUaPassThroughEnabled(previousValue);
    }
  };

  const handlePassThroughChange = async (enabled: boolean) => {
    const previousValue = passThroughEnabled;
    setPassThroughEnabled(enabled);
    try {
      await updatePTSettings.mutateAsync({ enabled });
    } catch {
      // Revert state on error
      setPassThroughEnabled(previousValue);
    }
  };

  const handlePreferPassThroughChange = async (enabled: boolean) => {
    const previousValue = preferPassThroughEnabled;
    setPreferPassThroughEnabled(enabled);
    try {
      await updatePTSettings.mutateAsync({ preferPassThrough: enabled });
    } catch {
      setPreferPassThroughEnabled(previousValue);
    }
  };

  const handlePreferPassThroughExceptionsChange = async (enabled: boolean) => {
    const previousValue = preferPassThroughExceptionsEnabled;
    setPreferPassThroughExceptionsEnabled(enabled);
    try {
      await updatePTSettings.mutateAsync({ preferPassThroughExceptionsEnabled: enabled });
    } catch {
      setPreferPassThroughExceptionsEnabled(previousValue);
    }
  };

  const handleConversionExceptionToggle = (key: string) => {
    setPreferPassThroughExceptionConversions((current) =>
      current.includes(key) ? current.filter((conversion) => conversion !== key) : [...current, key]
    );
  };

  const handleSaveConversionExceptions = async () => {
    const previousValue = ptSettings?.preferPassThroughExceptionConversions ?? [];
    try {
      await updatePTSettings.mutateAsync({
        preferPassThroughExceptionConversions: preferPassThroughExceptionConversions,
      });
    } catch {
      setPreferPassThroughExceptionConversions(previousValue);
    }
  };

  const conversionGroups = React.useMemo(() => {
    const groups = new Map<string, NonNullable<typeof ptSettings>['availableConversions']>();
    for (const conversion of ptSettings?.availableConversions ?? []) {
      const group = groups.get(conversion.requestType) ?? [];
      group.push(conversion);
      groups.set(conversion.requestType, group);
    }
    return Array.from(groups.entries());
  }, [ptSettings]);

  const conversionExceptionsHaveChanges = React.useMemo(() => {
    const saved = [...(ptSettings?.preferPassThroughExceptionConversions ?? [])].sort();
    const selected = [...preferPassThroughExceptionConversions].sort();
    return saved.length !== selected.length || saved.some((value, index) => value !== selected[index]);
  }, [preferPassThroughExceptionConversions, ptSettings]);

  const formatLabel = (format: string) => t(`models.dialogs.association.conditions.formatOptions.${format}`, { defaultValue: format });

  const hasChanges = settings ? settings.currencyCode !== currencyCode || settings.timezone !== timezone : false;

  if (isLoadingSettings) {
    return (
      <div className='flex h-32 items-center justify-center'>
        <Loader2 className='h-6 w-6 animate-spin' />
        <span className='text-muted-foreground ml-2'>{t('common.loading')}</span>
      </div>
    );
  }

  return (
    <div className='space-y-6'>
      <Card>
        <CardHeader>
          <CardTitle>{t('system.general.title')}</CardTitle>
          <CardDescription>{t('system.general.description')}</CardDescription>
        </CardHeader>
        <CardContent className='space-y-6'>
          <div className='space-y-2'>
            <Label htmlFor='currency-code'>{t('system.general.currencyCode.label')}</Label>
            <div className='max-w-md'>
              <AutoCompleteSelect
                selectedValue={currencyCode}
                onSelectedValueChange={setCurrencyCode}
                items={currencyItems}
                placeholder={t('system.general.currencyCode.placeholder')}
                isLoading={isLoadingSettings}
              />
            </div>
            <div className='text-muted-foreground text-sm'>{t('system.general.currencyCode.description')}</div>
          </div>

          <div className='space-y-2'>
            <Label htmlFor='timezone'>{t('system.general.timezone.label')}</Label>
            <div className='max-w-md'>
              <AutoCompleteSelect
                selectedValue={timezone}
                onSelectedValueChange={setTimezone}
                items={timezoneItems}
                placeholder={t('system.general.timezone.placeholder')}
                isLoading={isLoadingSettings}
              />
            </div>
            <div className='text-muted-foreground text-sm'>{t('system.general.timezone.description')}</div>
          </div>
        </CardContent>
      </Card>

      <ListPaginationSettings />

      <SidebarNavigationSettings />

      <Card>
        <CardHeader>
          <CardTitle>{t('system.passThroughGroup.title')}</CardTitle>
          <CardDescription>{t('system.passThroughGroup.description')}</CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='flex items-center justify-between'>
            <div className='space-y-0.5'>
              <Label htmlFor='ua-pass-through'>{t('system.userAgentPassThrough.label')}</Label>
              <div className='text-muted-foreground text-sm'>{t('system.userAgentPassThrough.helpText')}</div>
            </div>
            <Switch
              id='ua-pass-through'
              checked={uaPassThroughEnabled}
              onCheckedChange={handleUAPassThroughChange}
              disabled={isLoadingUASettings || updateUASettings.isPending}
            />
          </div>
          <div className='flex items-center justify-between'>
            <div className='space-y-0.5'>
              <Label htmlFor='pass-through'>{t('system.passThrough.label')}</Label>
              <div className='text-muted-foreground text-sm'>{t('system.passThrough.helpText')}</div>
            </div>
            <Switch
              id='pass-through'
              checked={passThroughEnabled}
              onCheckedChange={handlePassThroughChange}
              disabled={isLoadingPTSettings || updatePTSettings.isPending}
            />
          </div>
          <div className='space-y-3'>
            <div className='flex items-center justify-between gap-4'>
              <div className='space-y-0.5'>
                <Label htmlFor='prefer-pass-through'>{t('system.preferPassThrough.label')}</Label>
                <div className='text-muted-foreground text-sm'>{t('system.preferPassThrough.helpText')}</div>
              </div>
              <Switch
                id='prefer-pass-through'
                checked={preferPassThroughEnabled}
                onCheckedChange={handlePreferPassThroughChange}
                disabled={isLoadingPTSettings || updatePTSettings.isPending}
              />
            </div>

            <div className={cn('bg-muted/30 space-y-3 rounded-lg border p-4', !preferPassThroughEnabled && 'opacity-60')}>
              <div className='flex items-center justify-between gap-4'>
                <div className='space-y-0.5'>
                  <Label htmlFor='prefer-pass-through-exceptions'>{t('system.preferPassThroughExceptions.label')}</Label>
                  <div className='text-muted-foreground text-sm'>{t('system.preferPassThroughExceptions.helpText')}</div>
                </div>
                <Switch
                  id='prefer-pass-through-exceptions'
                  checked={preferPassThroughExceptionsEnabled}
                  onCheckedChange={handlePreferPassThroughExceptionsChange}
                  disabled={!preferPassThroughEnabled || isLoadingPTSettings || updatePTSettings.isPending}
                />
              </div>

              {preferPassThroughExceptionsEnabled && (
                <div className='flex flex-col gap-2 sm:flex-row sm:items-center'>
                  <Popover>
                    <PopoverTrigger asChild>
                      <Button
                        variant='outline'
                        role='combobox'
                        className='w-full justify-between sm:max-w-md'
                        disabled={!preferPassThroughEnabled || isLoadingPTSettings || updatePTSettings.isPending}
                      >
                        {preferPassThroughExceptionConversions.length > 0
                          ? t('system.preferPassThroughExceptions.selected', {
                              count: preferPassThroughExceptionConversions.length,
                            })
                          : t('system.preferPassThroughExceptions.noneSelected')}
                        <ChevronsUpDown className='ml-2 h-4 w-4 shrink-0 opacity-50' />
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className='w-[min(34rem,calc(100vw-2rem))] p-0' align='start'>
                      <Command>
                        <CommandInput placeholder={t('system.preferPassThroughExceptions.searchPlaceholder')} />
                        <CommandList className='max-h-80'>
                          <CommandEmpty>{t('common.noResultsFound')}</CommandEmpty>
                          {conversionGroups.map(([requestType, conversions]) => (
                            <CommandGroup
                              key={requestType}
                              heading={t(`system.preferPassThroughExceptions.requestTypes.${requestType}`, {
                                defaultValue: requestType,
                              })}
                            >
                              {conversions.map((conversion) => {
                                const selected = preferPassThroughExceptionConversions.includes(conversion.key);
                                const source = formatLabel(conversion.sourceFormat);
                                const target = formatLabel(conversion.targetFormat);
                                return (
                                  <CommandItem
                                    key={conversion.key}
                                    value={`${source} ${target} ${conversion.sourceFormat} ${conversion.targetFormat}`}
                                    onSelect={() => handleConversionExceptionToggle(conversion.key)}
                                  >
                                    <Check className={cn('mr-2 h-4 w-4', selected ? 'opacity-100' : 'opacity-0')} />
                                    <span className='min-w-0 truncate'>{source}</span>
                                    <span className='text-muted-foreground shrink-0'>→</span>
                                    <span className='min-w-0 truncate'>{target}</span>
                                  </CommandItem>
                                );
                              })}
                            </CommandGroup>
                          ))}
                        </CommandList>
                      </Command>
                    </PopoverContent>
                  </Popover>
                  <Button
                    onClick={handleSaveConversionExceptions}
                    disabled={!conversionExceptionsHaveChanges || updatePTSettings.isPending}
                    className='sm:shrink-0'
                  >
                    {updatePTSettings.isPending && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
                    {t('system.preferPassThroughExceptions.save')}
                  </Button>
                </div>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {hasChanges && (
        <div className='flex justify-end'>
          <Button onClick={handleSave} disabled={isLoading || updateSettings.isPending} className='min-w-[100px]'>
            {isLoading || updateSettings.isPending ? (
              <>
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                {t('system.buttons.saving')}
              </>
            ) : (
              <>
                <Save className='mr-2 h-4 w-4' />
                {t('system.buttons.save')}
              </>
            )}
          </Button>
        </div>
      )}
    </div>
  );
}
