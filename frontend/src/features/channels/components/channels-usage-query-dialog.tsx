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
import { checkProviderQuotas } from '@/features/system/data/quotas';
import type { Channel } from '../data/schema';
import {
  type ChannelUsageQueryConfigInput,
  type ChannelUsageQueryPreset,
  type ChannelUsageQueryTestResult,
  useChannelUsageQuery,
  useSaveChannelUsageQuery,
  useTestChannelUsageQuery,
} from '../data/usage-query';

const NEW_API_SCRIPT = `({
  request: {
    url: "{{baseUrl}}/api/user/self",
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      "Authorization": "Bearer {{accessToken}}",
      "User-Agent": "cc-switch/1.0",
      "New-Api-User": "{{userId}}"
    },
  },
  extractor: function (response) {
    if (response.success && response.data) {
      return {
        planName: response.data.group || "默认套餐",
        remaining: response.data.quota / 500000,
        used: response.data.used_quota / 500000,
        total: (response.data.quota + response.data.used_quota) / 500000,
        unit: "USD",
      };
    }
    return {
      isValid: false,
      invalidMessage: response.message || "查询失败"
    };
  },
})`;

const CUSTOM_SCRIPT = `({
  request: {
    url: "{{baseUrl}}/user/balance",
    method: "GET",
    headers: {
      "Authorization": "Bearer {{apiKey}}",
      "User-Agent": "cc-switch/1.0"
    }
  },
  extractor: function(response) {
    return {
      isValid: response.is_active || true,
      remaining: response.balance,
      unit: "USD"
    };
  }
})`;

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

  const [enabled, setEnabled] = useState(true);
  const [preset, setPreset] = useState<ChannelUsageQueryPreset>('NEW_API');
  const [baseUrlOverride, setBaseUrlOverride] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [apiKeyConfigured, setApiKeyConfigured] = useState(false);
  const [clearApiKey, setClearApiKey] = useState(false);
  const [userId, setUserId] = useState('');
  const [script, setScript] = useState(NEW_API_SCRIPT);
  const [customScript, setCustomScript] = useState(CUSTOM_SCRIPT);
  const [testResult, setTestResult] = useState<ChannelUsageQueryTestResult | null>(null);

  useEffect(() => {
    if (!open || !data) return;
    const nextPreset = data.script ? data.preset : 'NEW_API';
    const nextScript = data.script || (nextPreset === 'NEW_API' ? NEW_API_SCRIPT : CUSTOM_SCRIPT);
    setEnabled(data.script ? data.enabled : true);
    setPreset(nextPreset);
    setBaseUrlOverride(data.baseUrlOverride ?? '');
    setApiKey('');
    setApiKeyConfigured(data.apiKeyConfigured);
    setClearApiKey(false);
    setUserId(data.userId ?? '');
    setScript(nextScript);
    setCustomScript(nextPreset === 'CUSTOM' ? nextScript : CUSTOM_SCRIPT);
    setTestResult(null);
  }, [data, open]);

  const handlePresetChange = (value: ChannelUsageQueryPreset) => {
    if (preset === 'CUSTOM') setCustomScript(script);
    setPreset(value);
    setScript(value === 'NEW_API' ? NEW_API_SCRIPT : customScript || CUSTOM_SCRIPT);
    setTestResult(null);
  };

  const buildInput = (): ChannelUsageQueryConfigInput => ({
    enabled,
    preset,
    baseUrlOverride: baseUrlOverride.trim() || null,
    userId: preset === 'NEW_API' ? userId.trim() || null : null,
    script,
    apiKey: clearApiKey ? null : apiKey.trim() || null,
    clearApiKey,
  });

  const canSubmit = !isError && script.trim() !== '' && (preset !== 'NEW_API' || userId.trim() !== '');

  const formatValue = (value: number | null | undefined, unit?: string | null): string => {
    if (value == null) return '-';
    if (unit && /^[A-Z]{3}$/.test(unit)) {
      return t('currencies.format', {
        val: value,
        currency: unit,
        locale: i18n.language === 'zh' ? 'zh-CN' : 'en-US',
        minimumFractionDigits: 0,
        maximumFractionDigits: 6,
      });
    }
    const formatted = new Intl.NumberFormat(i18n.language === 'zh' ? 'zh-CN' : 'en-US', {
      maximumFractionDigits: 6,
    }).format(value);
    return unit ? `${formatted} ${unit}` : formatted;
  };

  const handleTest = async () => {
    setTestResult(null);
    try {
      const result = await testUsageQuery.mutateAsync({ channelID: currentRow.id, input: buildInput() });
      setTestResult(result);
      if (result.isValid === false) {
        toast.error(result.invalidMessage || t('channels.dialogs.usageQuery.test.invalid'));
      } else {
        toast.success(t('channels.dialogs.usageQuery.test.success'));
      }
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
        void checkProviderQuotas();
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
                    <SelectItem value='CUSTOM'>{t('channels.dialogs.usageQuery.preset.custom')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className='flex h-10 items-center justify-between gap-4 sm:justify-end'>
                <Label htmlFor='usage-query-enabled'>{t('channels.dialogs.usageQuery.enabled')}</Label>
                <Switch id='usage-query-enabled' checked={enabled} onCheckedChange={setEnabled} />
              </div>
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

            {preset === 'NEW_API' && (
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
              <Label htmlFor='usage-query-script'>{t('channels.dialogs.usageQuery.script.label')}</Label>
              <Textarea
                id='usage-query-script'
                value={script}
                onChange={(event) => {
                  setScript(event.target.value);
                  if (preset === 'CUSTOM') setCustomScript(event.target.value);
                  setTestResult(null);
                }}
                spellCheck={false}
                className='min-h-80 resize-y font-mono text-xs leading-5'
              />
            </div>

            {testResult && (
              <Alert variant={testResult.isValid === false ? 'destructive' : 'default'}>
                <AlertTitle className='flex items-center gap-2'>
                  {t('channels.dialogs.usageQuery.test.result')}
                  <Badge variant='outline'>{t(`quota.status.${testResult.status}`)}</Badge>
                </AlertTitle>
                <AlertDescription>
                  {testResult.isValid === false ? (
                    testResult.invalidMessage || t('channels.dialogs.usageQuery.test.invalid')
                  ) : (
                    <div className='mt-2 grid gap-x-6 gap-y-2 text-sm sm:grid-cols-2'>
                      {testResult.planName && (
                        <div>
                          <span className='text-muted-foreground'>{t('channels.dialogs.usageQuery.result.plan')}</span>{' '}
                          {testResult.planName}
                        </div>
                      )}
                      {testResult.remaining != null && (
                        <div>
                          <span className='text-muted-foreground'>{t('channels.dialogs.usageQuery.result.remaining')}</span>{' '}
                          {formatValue(testResult.remaining, testResult.unit)}
                        </div>
                      )}
                      {testResult.used != null && (
                        <div>
                          <span className='text-muted-foreground'>{t('channels.dialogs.usageQuery.result.used')}</span>{' '}
                          {formatValue(testResult.used, testResult.unit)}
                        </div>
                      )}
                      {testResult.total != null && (
                        <div>
                          <span className='text-muted-foreground'>{t('channels.dialogs.usageQuery.result.total')}</span>{' '}
                          {formatValue(testResult.total, testResult.unit)}
                        </div>
                      )}
                      {testResult.extra && <div className='sm:col-span-2'>{testResult.extra}</div>}
                    </div>
                  )}
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
