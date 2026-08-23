'use client';

import { useEffect, useMemo, useState } from 'react';
import { IconCircleOff, IconFlask, IconPlayerPlay, IconSearch } from '@tabler/icons-react';
import { useTranslation } from 'react-i18next';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { ChannelTestFormatSelector } from './channel-test-format-selector';
import { useTestChannel, useUpdateChannel } from '../data/channels';
import {
  ChannelTestAPIFormat,
  getChannelTestAPIFormat,
  getDefaultAvailableChannelTestAPIFormats,
} from '../data/channel-test-api-formats';
import { Channel } from '../data/schema';
import { ErrorDisplay } from '../utils/error-formatter';
import { mergeChannelSettingsForUpdate } from '../utils/merge';
import { disableModelAPIFormat, isModelAPIFormatDisabled } from '../utils/model-api-format-disables';

type TestStatus = 'not_started' | 'testing' | 'success' | 'failed' | 'skipped' | 'disabled';

interface ModelTestResult {
  status: TestStatus;
  latency?: number;
  error?: string;
}

type ModelTestResults = Record<string, Partial<Record<ChannelTestAPIFormat, ModelTestResult>>>;

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  channel: Channel;
}

const MAX_CONCURRENT_TESTS = 4;

function makeInitialResults(models: string[]): ModelTestResults {
  return Object.fromEntries(models.map((model) => [model, {}]));
}

export function ChannelsTestDialog({ open, onOpenChange, channel }: Props) {
  const { t } = useTranslation();
  const [dialogContent, setDialogContent] = useState<HTMLDivElement | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedModels, setSelectedModels] = useState<string[]>([]);
  const [testResults, setTestResults] = useState<ModelTestResults>({});
  const [localSupportedModels, setLocalSupportedModels] = useState<string[]>(channel.supportedModels);
  const [selectedAPIFormats, setSelectedAPIFormats] = useState<ChannelTestAPIFormat[]>([]);
  const [isTesting, setIsTesting] = useState(false);
  const [isRemovePopoverOpen, setIsRemovePopoverOpen] = useState(false);
  const [disabledModelAPIFormats, setDisabledModelAPIFormats] = useState(channel.settings?.disabledModelApiFormats ?? []);
  const testChannel = useTestChannel();
  const updateChannel = useUpdateChannel();

  const availableEndpointFormats = useMemo(
    () => new Set([...(channel.defaultEndpoints ?? []), ...(channel.endpoints ?? [])].map((endpoint) => endpoint.apiFormat)),
    [channel.defaultEndpoints, channel.endpoints]
  );
  const defaultAvailableAPIFormats = useMemo(
    () => getDefaultAvailableChannelTestAPIFormats(availableEndpointFormats),
    [availableEndpointFormats]
  );
  const filteredModels = localSupportedModels.filter((model) => model.toLowerCase().includes(searchQuery.toLowerCase()));

  useEffect(() => {
    if (open) {
      setTestResults(makeInitialResults(channel.supportedModels));
      setLocalSupportedModels(channel.supportedModels);
      setSelectedModels([]);
      setSelectedAPIFormats(defaultAvailableAPIFormats);
      setSearchQuery('');
      setDisabledModelAPIFormats(channel.settings?.disabledModelApiFormats ?? []);
    }
  }, [channel.settings?.disabledModelApiFormats, channel.supportedModels, defaultAvailableAPIFormats, open]);

  const isFormatAvailable = (format: ChannelTestAPIFormat) => availableEndpointFormats.has(getChannelTestAPIFormat(format).endpointFormat);

  const handleAPIFormatsChange = (formats: ChannelTestAPIFormat[]) => {
    setSelectedAPIFormats(formats.filter(isFormatAvailable));
    setTestResults(makeInitialResults(localSupportedModels));
  };

  const handleModelSelect = (modelName: string, checked: boolean) => {
    setSelectedModels((previous) => (checked ? [...previous, modelName] : previous.filter((model) => model !== modelName)));
  };

  const handleSelectAll = (checked: boolean) => {
    setSelectedModels(checked ? filteredModels : []);
  };

  const setTestResult = (modelName: string, format: ChannelTestAPIFormat, result: ModelTestResult) => {
    setTestResults((previous) => ({
      ...previous,
      [modelName]: {
        ...previous[modelName],
        [format]: result,
      },
    }));
  };

  const testModel = async (modelName: string, format: ChannelTestAPIFormat) => {
    const endpointFormat = getChannelTestAPIFormat(format).endpointFormat;
    if (isModelAPIFormatDisabled(disabledModelAPIFormats, modelName, endpointFormat)) {
      setTestResult(modelName, format, { status: 'disabled' });
      return;
    }

    if (!isFormatAvailable(format)) {
      setTestResult(modelName, format, { status: 'skipped', error: t('channels.dialogs.test.formatUnavailable') });
      return;
    }

    setTestResult(modelName, format, { status: 'testing' });

    try {
      const startTime = Date.now();
      const result = await testChannel.mutateAsync({ channelID: channel.id, modelID: modelName, apiFormat: format });
      const latency = (Date.now() - startTime) / 1000;

      setTestResult(modelName, format, {
        status: result.success ? 'success' : 'failed',
        latency: result.success ? result.latency || latency : undefined,
        error: result.success ? undefined : result.error || t('channels.dialogs.test.testFailed'),
      });
    } catch (error) {
      setTestResult(modelName, format, {
        status: 'failed',
        error: error instanceof Error ? error.message : t('common.errors.internalServerError'),
      });
    }
  };

  const runTests = async (tasks: Array<{ modelName: string; format: ChannelTestAPIFormat }>) => {
    const queue = [...tasks];
    const workerCount = Math.min(MAX_CONCURRENT_TESTS, queue.length);
    await Promise.all(
      Array.from({ length: workerCount }, async () => {
        while (queue.length > 0) {
          const task = queue.shift();
          if (!task) return;
          await testModel(task.modelName, task.format);
        }
      })
    );
  };

  const handleTestSelected = async () => {
    const tasks = selectedModels.flatMap((modelName) =>
      selectedAPIFormats
        .filter((format) => isFormatAvailable(format))
        .filter((format) => !isModelAPIFormatDisabled(disabledModelAPIFormats, modelName, getChannelTestAPIFormat(format).endpointFormat))
        .map((format) => ({ modelName, format }))
    );
    if (tasks.length === 0 || isTesting) return;

    setIsTesting(true);
    try {
      await runTests(tasks);
    } finally {
      setIsTesting(false);
    }
  };

  const getStatusBadge = (status: TestStatus) => {
    switch (status) {
      case 'testing':
        return <Badge variant='secondary'>{t('channels.dialogs.test.testingModel')}</Badge>;
      case 'success':
        return <Badge className='border-green-200 bg-green-100 text-green-800'>{t('channels.dialogs.test.testSuccess')}</Badge>;
      case 'failed':
        return <Badge variant='destructive'>{t('channels.dialogs.test.testFailed')}</Badge>;
      case 'skipped':
        return <Badge variant='outline'>{t('channels.dialogs.test.formatUnavailable')}</Badge>;
      case 'disabled':
        return <Badge variant='outline'>{t('channels.dialogs.modelFormatDisables.disabled')}</Badge>;
      default:
        return <Badge variant='outline'>{t('channels.dialogs.test.notStarted')}</Badge>;
    }
  };

  const isAllSelected = filteredModels.length > 0 && filteredModels.every((model) => selectedModels.includes(model));
  const isIndeterminate = selectedModels.length > 0 && !isAllSelected;
  const failedModels = selectedModels.filter((model) => {
    const availableFormats = selectedAPIFormats.filter(isFormatAvailable);
    return availableFormats.length > 0 && availableFormats.every((format) => testResults[model]?.[format]?.status === 'failed');
  });

  const handleRemoveFailed = async () => {
    const failedModelNames = new Set(failedModels);
    const newSupportedModels = localSupportedModels.filter((model) => !failedModelNames.has(model));

    try {
      await updateChannel.mutateAsync({ id: channel.id, input: { supportedModels: newSupportedModels } });
      setLocalSupportedModels(newSupportedModels);
      setSelectedModels((previous) => previous.filter((model) => !failedModelNames.has(model)));
      setTestResults(makeInitialResults(newSupportedModels));
      setIsRemovePopoverOpen(false);
    } catch (_error) {
      // Errors are handled by useUpdateChannel toast.
    }
  };

  const handleDisableModelFormat = async (modelName: string, format: ChannelTestAPIFormat) => {
    const nextDisabledModelAPIFormats = disableModelAPIFormat(
      disabledModelAPIFormats,
      modelName,
      getChannelTestAPIFormat(format).endpointFormat
    );

    try {
      await updateChannel.mutateAsync({
        id: channel.id,
        input: {
          settings: mergeChannelSettingsForUpdate(channel.settings, {
            disabledModelApiFormats: nextDisabledModelAPIFormats,
          }),
        },
      });
      setDisabledModelAPIFormats(nextDisabledModelAPIFormats);
      setTestResult(modelName, format, { status: 'disabled' });
    } catch {
      // Errors are handled by useUpdateChannel toast.
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent ref={setDialogContent} className='flex max-h-[90vh] w-full max-w-full flex-col sm:max-w-6xl'>
        <DialogHeader>
          <DialogTitle className='text-lg sm:text-xl'>{t('channels.dialogs.test.title')}</DialogTitle>
          <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
            <DialogDescription className='text-sm sm:text-base'>{t('channels.dialogs.test.description', { name: channel.name })}</DialogDescription>
            <ChannelTestFormatSelector
              value={selectedAPIFormats}
              onChange={handleAPIFormatsChange}
              availableEndpointFormats={availableEndpointFormats}
              portalContainer={dialogContent}
            />
          </div>
        </DialogHeader>

        <div className='min-h-0 flex-1 space-y-4'>
          <div className='relative'>
            <IconSearch className='text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 transform' />
            <Input
              placeholder={t('channels.dialogs.test.searchPlaceholder')}
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              className='h-10 pl-10 sm:h-9'
            />
          </div>

          <div className='min-h-0 flex-1 overflow-hidden rounded-lg border'>
            <div className='max-h-96 overflow-auto'>
              <Table className='min-w-max'>
                <TableHeader>
                  <TableRow>
                    <TableHead className='w-14 sm:w-12'>
                      <Checkbox
                        checked={isAllSelected}
                        onCheckedChange={handleSelectAll}
                        ref={(element) => {
                          const input = element?.querySelector('input') as HTMLInputElement | null;
                          if (input) input.indeterminate = isIndeterminate;
                        }}
                        className='scale-100 sm:scale-75'
                      />
                    </TableHead>
                    <TableHead className='min-w-48'>{t('channels.dialogs.test.modelNameColumn')}</TableHead>
                    {selectedAPIFormats.map((format) => (
                      <TableHead key={format} className='min-w-52'>
                        {t(getChannelTestAPIFormat(format).labelKey)}
                      </TableHead>
                    ))}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredModels.map((model) => (
                    <TableRow key={model} className='align-top'>
                      <TableCell>
                        <Checkbox checked={selectedModels.includes(model)} onCheckedChange={(checked) => handleModelSelect(model, !!checked)} className='scale-100 sm:scale-75' />
                      </TableCell>
                      <TableCell className='pr-4 font-medium sm:pr-8'>{model}</TableCell>
                      {selectedAPIFormats.map((format) => {
                        const result = testResults[model]?.[format];
                        const available = isFormatAvailable(format);
                        const disabled = isModelAPIFormatDisabled(
                          disabledModelAPIFormats,
                          model,
                          getChannelTestAPIFormat(format).endpointFormat
                        );
                        return (
                          <TableCell key={format} className='min-w-52 align-top'>
                            <div className='space-y-2'>
                              {getStatusBadge(disabled ? 'disabled' : available ? result?.status || 'not_started' : 'skipped')}
                              {typeof result?.latency === 'number' && <div className='text-muted-foreground text-xs'>{result.latency.toFixed(2)}s</div>}
                              {result?.error && <ErrorDisplay error={result.error} messageClassName='text-xs font-medium text-red-600' />}
                              <Button
                                size='sm'
                                variant='outline'
                                onClick={() => testModel(model, format)}
                                disabled={!available || disabled || result?.status === 'testing' || isTesting || updateChannel.isPending}
                              >
                                <IconPlayerPlay className='h-3 w-3' />
                                {result?.status === 'testing' ? t('channels.dialogs.test.testingModel') : t('channels.dialogs.test.testModel')}
                              </Button>
                              {result?.status === 'failed' && !disabled && (
                                <Button
                                  size='sm'
                                  variant='outline'
                                  className='border-destructive/50 text-destructive hover:bg-destructive/10 hover:text-destructive'
                                  onClick={() => handleDisableModelFormat(model, format)}
                                  disabled={isTesting || updateChannel.isPending}
                                >
                                  <IconCircleOff className='h-3 w-3' />
                                  {t('channels.dialogs.modelFormatDisables.disableThisFormat')}
                                </Button>
                              )}
                            </div>
                          </TableCell>
                        );
                      })}
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
        </div>

        <DialogFooter className='flex flex-col items-stretch justify-between gap-3 sm:flex-row sm:items-center sm:gap-2'>
          <div className='flex w-full flex-col gap-2 sm:w-auto sm:flex-row'>
            <Button variant='outline' onClick={() => onOpenChange(false)} className='w-full sm:w-auto'>
              {t('common.buttons.cancel')}
            </Button>
            {failedModels.length > 0 && (
              <Popover open={isRemovePopoverOpen} onOpenChange={setIsRemovePopoverOpen}>
                <PopoverTrigger asChild>
                  <Button variant='destructive' size='sm' className='h-10 sm:h-8'>
                    {t('channels.dialogs.test.removeFailed')} ({failedModels.length})
                  </Button>
                </PopoverTrigger>
                <PopoverContent container={dialogContent} className='w-full sm:w-80'>
                  <div className='grid gap-4'>
                    <p className='text-muted-foreground text-sm'>{t('channels.dialogs.test.removeFailedConfirm')}</p>
                    <div className='flex justify-end gap-2'>
                      <Button size='sm' variant='destructive' onClick={handleRemoveFailed} disabled={updateChannel.isPending}>
                        {updateChannel.isPending ? t('common.buttons.saving') : t('common.buttons.confirm')}
                      </Button>
                    </div>
                  </div>
                </PopoverContent>
              </Popover>
            )}
          </div>
          <Button onClick={handleTestSelected} disabled={selectedModels.length === 0 || isTesting} className='h-10 sm:h-9'>
            <IconFlask className='h-4 w-4' />
            {t('channels.dialogs.test.batchTestButton', { count: selectedModels.length, formats: selectedAPIFormats.length })}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
