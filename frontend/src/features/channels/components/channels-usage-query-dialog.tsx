'use client';

import { useEffect, useState } from 'react';
import { IconPlayerPlay, IconRefresh, IconTrash } from '@tabler/icons-react';
import { Loader2, Save } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import type { Channel } from '../data/schema';
import {
  type ChannelUsageQueryConfigInput,
  type ChannelUsageQueryPreset,
  type ChannelUsageQueryProgressWindow,
  type ChannelUsageQueryTestResult,
  useChannelUsageQuery,
  useRefreshAllUsageQueries,
  useSaveChannelUsageQuery,
  useTestChannelUsageQuery,
} from '../data/usage-query';
import { getUsageQueryPreset } from '../data/usage-query-presets';

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentRow: Channel;
}

export function ChannelsUsageQueryDialog({ open, onOpenChange, currentRow }: Props) {
  const { t, i18n } = useTranslation();
  const { data, isLoading, isError, error } = useChannelUsageQuery(currentRow.id, open);
  const saveUsageQuery = useSaveChannelUsageQuery();
  const testUsageQuery = useTestChannelUsageQuery();
  const refreshAllUsageQueries = useRefreshAllUsageQueries();

  const [enabled, setEnabled] = useState(true);
  const [showInProviderQuota, setShowInProviderQuota] = useState(true);
  const [preset, setPreset] = useState<ChannelUsageQueryPreset>('NEW_API');
  const [baseUrlOverride, setBaseUrlOverride] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [apiKeyConfigured, setApiKeyConfigured] = useState(false);
  const [clearApiKey, setClearApiKey] = useState(false);
  const [userId, setUserId] = useState('');
  const [script, setScript] = useState(getUsageQueryPreset('NEW_API').script);
  const [customScript, setCustomScript] = useState(getUsageQueryPreset('CUSTOM').script);
  const [testResult, setTestResult] = useState<ChannelUsageQueryTestResult | null>(null);

  useEffect(() => {
    if (!open || !data) return;
    const nextPreset = data.script ? data.preset : 'NEW_API';
    const nextScript = data.script || getUsageQueryPreset(nextPreset).script;
    setEnabled(data.enabled);
    setShowInProviderQuota(data.showInProviderQuota);
    setPreset(nextPreset);
    setBaseUrlOverride(data.baseUrlOverride ?? '');
    setApiKey('');
    setApiKeyConfigured(data.apiKeyConfigured);
    setClearApiKey(false);
    setUserId(data.userId ?? '');
    setScript(nextScript);
    setCustomScript(nextPreset === 'CUSTOM' ? nextScript : getUsageQueryPreset('CUSTOM').script);
    setTestResult(null);
  }, [data, open]);

  const handlePresetChange = (value: ChannelUsageQueryPreset) => {
    if (preset === 'CUSTOM') setCustomScript(script);
    setPreset(value);
    setScript(value === 'CUSTOM' ? customScript || getUsageQueryPreset('CUSTOM').script : getUsageQueryPreset(value).script);
    const defaultBaseUrl = getUsageQueryPreset(value).defaultBaseUrl;
    if (defaultBaseUrl) setBaseUrlOverride(defaultBaseUrl);
    setTestResult(null);
  };

  const convertPresetToCustom = () => {
    setCustomScript(script);
    setPreset('CUSTOM');
    setTestResult(null);
  };

  const buildInput = (): ChannelUsageQueryConfigInput => ({
    enabled,
    showInProviderQuota,
    preset,
    baseUrlOverride: baseUrlOverride.trim() || null,
    userId: getUsageQueryPreset(preset).requiresUserId ? userId.trim() || null : null,
    script,
    apiKey: clearApiKey ? null : apiKey.trim() || null,
    clearApiKey,
  });

  const canSubmit = !isError && script.trim() !== '' && (!getUsageQueryPreset(preset).requiresUserId || userId.trim() !== '');

  const formatValue = (value: number | null | undefined, unit?: string | null): string => {
    if (value == null) return '-';
    if (unit && /^[A-Z]{3}$/.test(unit)) {
      return t('currencies.format', {
        val: value,
        currency: unit,
        locale: i18n.language === 'zh' ? 'zh-CN' : 'en-US',
        minimumFractionDigits: unit === 'A$' ? 2 : 0,
        maximumFractionDigits: unit === 'A$' ? 2 : 6,
      });
    }
    const formatted = new Intl.NumberFormat(i18n.language === 'zh' ? 'zh-CN' : 'en-US', {
      minimumFractionDigits: unit === 'A$' ? 2 : 0,
      maximumFractionDigits: unit === 'A$' ? 2 : 6,
    }).format(value);
    return unit === 'A$' ? `A$${formatted}` : unit ? `${formatted} ${unit}` : formatted;
  };

  const getProgressPercent = (window: ChannelUsageQueryProgressWindow): number | null => {
    return window.remainingPercent == null ? null : 100 - window.remainingPercent;
  };

  const handleTest = async () => {
    setTestResult(null);
    try {
      const result = await testUsageQuery.mutateAsync({ channelID: currentRow.id, input: buildInput() });
      setTestResult(result);
      toast.success(t('channels.dialogs.usageQuery.test.success'));
    } catch (error) {
      toast.error(t('channels.dialogs.usageQuery.test.failed'), {
        description: error instanceof Error ? error.message : String(error),
      });
    }
  };

  const handleSave = async () => {
    try {
      const saved = await saveUsageQuery.mutateAsync({ channelID: currentRow.id, input: buildInput() });
      setApiKeyConfigured(saved.apiKeyConfigured);
      setApiKey('');
      setClearApiKey(false);
      toast.success(t('channels.dialogs.usageQuery.saveSuccess'));
      onOpenChange(false);
      if (saved.enabled) {
        refreshAllUsageQueries.mutate();
      }
    } catch (error) {
      toast.error(t('channels.dialogs.usageQuery.saveFailed'), {
        description: error instanceof Error ? error.message : String(error),
      });
    }
  };

  const resetSavedKeyAction = () => {
    setClearApiKey(!clearApiKey);
    setApiKey('');
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[92vh] overflow-y-auto sm:max-w-4xl'>
        <DialogHeader className='text-left'>
          <DialogTitle>{t('channels.dialogs.usageQuery.title')}</DialogTitle>
          <DialogDescription>{t('channels.dialogs.usageQuery.description', { name: currentRow.name })}</DialogDescription>
        </DialogHeader>

        {isLoading ? (
          <div className='flex min-h-80 items-center justify-center'>
            <Loader2 className='text-muted-foreground h-6 w-6 animate-spin' />
          </div>
        ) : (
          <div className='space-y-5'>
            {isError && (
              <Alert variant='destructive'>
                <AlertTitle>{t('channels.dialogs.usageQuery.loadFailed')}</AlertTitle>
                <AlertDescription>{error instanceof Error ? error.message : String(error)}</AlertDescription>
              </Alert>
            )}
            <div className='grid gap-4 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end'>
              <div className='space-y-2'>
                <Label htmlFor='usage-query-preset'>{t('channels.dialogs.usageQuery.preset.label')}</Label>
                <Select value={preset} onValueChange={(value) => handlePresetChange(value as ChannelUsageQueryPreset)}>
                  <SelectTrigger id='usage-query-preset'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value='NEW_API'>{t('channels.dialogs.usageQuery.preset.newApi')}</SelectItem>
                    <SelectItem value='CODEX'>{t('channels.dialogs.usageQuery.preset.codex')}</SelectItem>
                    <SelectItem value='CLAUDE_OAUTH'>{t('channels.dialogs.usageQuery.preset.claudeOauth')}</SelectItem>
                    <SelectItem value='OPENCODE_GO'>{t('channels.dialogs.usageQuery.preset.opencodeGo')}</SelectItem>
                    <SelectItem value='CUSTOM'>{t('channels.dialogs.usageQuery.preset.custom')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className='flex h-10 items-center justify-between gap-4 sm:justify-end'>
                <Label htmlFor='usage-query-enabled'>{t('channels.dialogs.usageQuery.enabled')}</Label>
                <Switch id='usage-query-enabled' checked={enabled} onCheckedChange={setEnabled} />
              </div>
            </div>

            <div className='flex items-center justify-between gap-4 rounded-md border p-3'>
              <div className='space-y-1'>
                <Label htmlFor='usage-query-show-in-provider-quota'>
                  {t('channels.dialogs.usageQuery.showInProviderQuota.label')}
                </Label>
                <p className='text-muted-foreground text-sm'>{t('channels.dialogs.usageQuery.showInProviderQuota.description')}</p>
              </div>
              <Switch
                id='usage-query-show-in-provider-quota'
                checked={showInProviderQuota}
                onCheckedChange={setShowInProviderQuota}
              />
            </div>

            <Separator />

            <div className='grid gap-4 sm:grid-cols-2'>
              <div className='space-y-2'>
                <Label htmlFor='usage-query-base-url'>{t('channels.dialogs.usageQuery.baseUrl.label')}</Label>
                <Input
                  id='usage-query-base-url'
                  value={baseUrlOverride}
                  onChange={(event) => setBaseUrlOverride(event.target.value)}
                  placeholder={currentRow.baseURL}
                />
              </div>
              <div className='space-y-2'>
                <div className='flex items-center justify-between gap-2'>
                  <Label htmlFor='usage-query-api-key'>{t('channels.dialogs.usageQuery.apiKey.label')}</Label>
                  {apiKeyConfigured && !clearApiKey && (
                    <Badge variant='outline'>{t('channels.dialogs.usageQuery.apiKey.configured')}</Badge>
                  )}
                  {apiKeyConfigured && clearApiKey && (
                    <Badge variant='destructive'>{t('channels.dialogs.usageQuery.apiKey.pendingClear')}</Badge>
                  )}
                </div>
                <div className='flex gap-2'>
                  <Input
                    id='usage-query-api-key'
                    type='password'
                    value={apiKey}
                    disabled={clearApiKey}
                    onChange={(event) => {
                      setApiKey(event.target.value);
                      setClearApiKey(false);
                    }}
                    placeholder={t(
                      apiKeyConfigured
                        ? 'channels.dialogs.usageQuery.apiKey.keepPlaceholder'
                        : 'channels.dialogs.usageQuery.apiKey.fallbackPlaceholder'
                    )}
                  />
                  {apiKeyConfigured && (
                    <Button
                      type='button'
                      variant='outline'
                      size='icon'
                      className='shrink-0'
                      onClick={resetSavedKeyAction}
                      title={t(
                        clearApiKey
                          ? 'channels.dialogs.usageQuery.apiKey.undoClear'
                          : 'channels.dialogs.usageQuery.apiKey.clear'
                      )}
                    >
                      {clearApiKey ? <IconRefresh size={16} /> : <IconTrash size={16} />}
                    </Button>
                  )}
                </div>
              </div>
            </div>

            {getUsageQueryPreset(preset).requiresUserId && (
              <div className='space-y-2'>
                <Label htmlFor='usage-query-user-id'>{t('channels.dialogs.usageQuery.userId.label')}</Label>
                <Input
                  id='usage-query-user-id'
                  value={userId}
                  onChange={(event) => setUserId(event.target.value)}
                  placeholder={t('channels.dialogs.usageQuery.userId.placeholder')}
                />
              </div>
            )}

            <div className='space-y-2'>
              <div className='flex items-center justify-between gap-3'>
                <Label htmlFor='usage-query-script'>{t('channels.dialogs.usageQuery.script.label')}</Label>
                {preset !== 'CUSTOM' && (
                  <Button type='button' size='sm' variant='outline' onClick={convertPresetToCustom}>
                    {t('channels.dialogs.usageQuery.actions.editPreset')}
                  </Button>
                )}
              </div>
              <Textarea
                id='usage-query-script'
                value={script}
                onChange={(event) => {
                  setScript(event.target.value);
                  if (preset === 'CUSTOM') setCustomScript(event.target.value);
                  setTestResult(null);
                }}
                spellCheck={false}
                readOnly={preset !== 'CUSTOM'}
                className='min-h-80 resize-y font-mono text-xs leading-5'
              />
            </div>

            {testResult && (
              <Alert>
                <AlertTitle className='flex items-center gap-2'>
                  {t('channels.dialogs.usageQuery.test.result')}
                  <Badge variant='outline'>{t(`quota.status.${testResult.status}`)}</Badge>
                </AlertTitle>
                <AlertDescription>
                  <div className='mt-2 grid gap-x-6 gap-y-2 text-sm sm:grid-cols-2'>
                      {testResult.balance && (
                        <div>
                          <span className='text-muted-foreground'>{t('channels.dialogs.usageQuery.result.remaining')}</span>{' '}
                          {formatValue(testResult.balance.remaining, testResult.balance.unit)}
                        </div>
                      )}
                      {(testResult.tags ?? []).length > 0 && (
                        <div className='flex flex-wrap gap-1 sm:col-span-2'>
                          {testResult.tags?.map((tag, index) => (
                            <Badge key={`${tag}-${index}`} variant='secondary'>
                              {tag}
                            </Badge>
                          ))}
                        </div>
                      )}
                      {testResult.text && <div className='whitespace-pre-line sm:col-span-2'>{testResult.text}</div>}
                      {(testResult.progress?.windows ?? []).map((window, index) => {
                        const percent = getProgressPercent(window);
                        const label = window.id || `#${index + 1}`;
                        const value = percent == null ? '-' : `${Math.round(100 - percent)}%`;
                        return (
                          <div key={`${label}-${index}`} className='space-y-1 sm:col-span-2'>
                            <div className='flex justify-between gap-3'>
                              <span className='text-muted-foreground'>{label}</span>
                              <span>{value}</span>
                            </div>
                            {percent != null && (
                              <div className='bg-muted h-1.5 overflow-hidden rounded-full'>
                                <div className='h-full bg-primary' style={{ width: `${Math.max(0, Math.min(100, percent))}%` }} />
                              </div>
                            )}
                          </div>
                        );
                      })}
                  </div>
                </AlertDescription>
              </Alert>
            )}
          </div>
        )}

        <DialogFooter>
          <Button type='button' variant='outline' onClick={() => onOpenChange(false)}>
            {t('common.buttons.cancel')}
          </Button>
          <Button type='button' variant='secondary' disabled={!canSubmit || testUsageQuery.isPending || isLoading} onClick={handleTest}>
            {testUsageQuery.isPending ? <Loader2 className='h-4 w-4 animate-spin' /> : <IconPlayerPlay size={16} />}
            {t('channels.dialogs.usageQuery.actions.test')}
          </Button>
          <Button type='button' disabled={!canSubmit || saveUsageQuery.isPending || isLoading} onClick={handleSave}>
            {saveUsageQuery.isPending ? <Loader2 className='h-4 w-4 animate-spin' /> : <Save className='h-4 w-4' />}
            {t('channels.dialogs.usageQuery.actions.save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
