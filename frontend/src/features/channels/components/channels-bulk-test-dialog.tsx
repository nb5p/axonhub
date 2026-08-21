'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { IconCheck, IconFlask, IconLoader2, IconPlayerPlay, IconRefresh } from '@tabler/icons-react';
import { useTranslation } from 'react-i18next';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { TruncatedText } from '@/components/truncated-text';
import { ChannelTestFormatSelector } from './channel-test-format-selector';
import { useChannels } from '../context/channels-context';
import { useBulkRecoverChannels, useTestChannel } from '../data/channels';
import { ChannelTestAPIFormat, defaultChannelTestAPIFormats, getChannelTestAPIFormat } from '../data/channel-test-api-formats';
import { Channel } from '../data/schema';
import { ErrorDisplay } from '../utils/error-formatter';

type BulkTestStatus = 'idle' | 'testing' | 'success' | 'failed' | 'skipped';

interface BulkTestResult {
  modelID?: string;
  status: BulkTestStatus;
  latency?: number;
  error?: string;
}

type BulkTestResults = Record<string, Partial<Record<ChannelTestAPIFormat, BulkTestResult>>>;

const MAX_CONCURRENT_TESTS = 4;

export function ChannelsBulkTestDialog() {
  const { t } = useTranslation();
  const { open, setOpen, selectedChannels, resetRowSelection, setSelectedChannels } = useChannels();
  const testChannel = useTestChannel({ silent: true });
  const bulkRecoverChannels = useBulkRecoverChannels();
  const [dialogContent, setDialogContent] = useState<HTMLDivElement | null>(null);
  const [results, setResults] = useState<BulkTestResults>({});
  const [selectedAPIFormats, setSelectedAPIFormats] = useState<ChannelTestAPIFormat[]>(defaultChannelTestAPIFormats);
  const [isTesting, setIsTesting] = useState(false);

  const isDialogOpen = open === 'bulkTest';

  const resolveTestModel = useCallback((channel: Channel) => channel.defaultTestModel || channel.supportedModels[0] || '', []);
  const isFormatAvailable = useCallback(
    (channel: Channel, format: ChannelTestAPIFormat) => {
      const endpointFormat = getChannelTestAPIFormat(format).endpointFormat;
      return [...(channel.defaultEndpoints ?? []), ...(channel.endpoints ?? [])].some((endpoint) => endpoint.apiFormat === endpointFormat);
    },
    []
  );

  const initializeResults = useCallback(() => {
    const nextResults = selectedChannels.reduce<BulkTestResults>((accumulator, channel) => {
      const modelID = resolveTestModel(channel);
      accumulator[channel.id] = {};
      selectedAPIFormats.forEach((format) => {
        const available = isFormatAvailable(channel, format);
        accumulator[channel.id][format] = {
          modelID: modelID || undefined,
          status: modelID && available ? 'idle' : 'skipped',
          error: modelID ? (available ? undefined : t('channels.dialogs.test.formatUnavailable')) : t('channels.dialogs.bulkTest.noTestModel'),
        };
      });
      return accumulator;
    }, {});

    setResults(nextResults);
  }, [isFormatAvailable, resolveTestModel, selectedAPIFormats, selectedChannels, t]);

  useEffect(() => {
    if (isDialogOpen && dialogContent) {
      initializeResults();
      setIsTesting(false);
    }
  }, [dialogContent, initializeResults, isDialogOpen]);

  const resultList = useMemo(
    () =>
      selectedChannels.flatMap((channel) =>
        selectedAPIFormats.map((format) => {
          const modelID = resolveTestModel(channel);
          const available = isFormatAvailable(channel, format);
          return (
            results[channel.id]?.[format] ?? {
              modelID: modelID || undefined,
              status: modelID && available ? ('idle' as const) : ('skipped' as const),
              error: modelID ? (available ? undefined : t('channels.dialogs.test.formatUnavailable')) : t('channels.dialogs.bulkTest.noTestModel'),
            }
          );
        })
      ),
    [isFormatAvailable, resolveTestModel, results, selectedAPIFormats, selectedChannels, t]
  );

  const completedCount = useMemo(() => resultList.filter((result) => ['success', 'failed', 'skipped'].includes(result.status)).length, [resultList]);
  const successCount = useMemo(() => resultList.filter((result) => result.status === 'success').length, [resultList]);
  const failedCount = useMemo(() => resultList.filter((result) => result.status === 'failed').length, [resultList]);
  const skippedCount = useMemo(() => resultList.filter((result) => result.status === 'skipped').length, [resultList]);

  const recoverableChannels = useMemo(
    () =>
      selectedChannels.filter((channel) => {
        const channelResults = selectedAPIFormats.map((format) => results[channel.id]?.[format]);
        const hasSuccess = channelResults.some((result) => result?.status === 'success');
        const hasFailure = channelResults.some((result) => result?.status === 'failed');
        return hasSuccess && !hasFailure && (channel.status === 'disabled' || !!channel.errorMessage);
      }),
    [results, selectedAPIFormats, selectedChannels]
  );

  const failedTasks = useMemo(
    () =>
      selectedChannels.flatMap((channel) =>
        selectedAPIFormats
          .filter((format) => results[channel.id]?.[format]?.status === 'failed')
          .map((format) => ({ channel, format }))
      ),
    [results, selectedAPIFormats, selectedChannels]
  );

  const setResultStatus = useCallback(
    (channel: Channel, format: ChannelTestAPIFormat, status: BulkTestStatus, extra?: Partial<BulkTestResult>) => {
      setResults((previous) => ({
        ...previous,
        [channel.id]: {
          ...previous[channel.id],
          [format]: {
            ...previous[channel.id]?.[format],
            modelID: previous[channel.id]?.[format]?.modelID || resolveTestModel(channel) || undefined,
            status,
            ...extra,
          },
        },
      }));
    },
    [resolveTestModel]
  );

  const runSingleTest = useCallback(
    async (channel: Channel, format: ChannelTestAPIFormat) => {
      const modelID = resolveTestModel(channel);
      if (!modelID) {
        setResultStatus(channel, format, 'skipped', { error: t('channels.dialogs.bulkTest.noTestModel') });
        return;
      }
      if (!isFormatAvailable(channel, format)) {
        setResultStatus(channel, format, 'skipped', { error: t('channels.dialogs.test.formatUnavailable') });
        return;
      }

      setResultStatus(channel, format, 'testing', { error: undefined, latency: undefined, modelID });
      try {
        const result = await testChannel.mutateAsync({ channelID: channel.id, modelID, apiFormat: format });
        setResultStatus(channel, format, result.success ? 'success' : 'failed', {
          modelID,
          latency: result.success ? result.latency : undefined,
          error: result.success ? undefined : result.error || t('common.errors.internalServerError'),
        });
      } catch (error) {
        setResultStatus(channel, format, 'failed', {
          modelID,
          error: error instanceof Error ? error.message : t('common.errors.internalServerError'),
        });
      }
    },
    [isFormatAvailable, resolveTestModel, setResultStatus, t, testChannel]
  );

  const runBatch = useCallback(
    async (tasks: Array<{ channel: Channel; format: ChannelTestAPIFormat }>) => {
      const queue = tasks.filter(({ channel, format }) => !!resolveTestModel(channel) && isFormatAvailable(channel, format));
      const workerCount = Math.min(MAX_CONCURRENT_TESTS, queue.length);
      await Promise.all(
        Array.from({ length: workerCount }, async () => {
          while (queue.length > 0) {
            const task = queue.shift();
            if (!task) return;
            await runSingleTest(task.channel, task.format);
          }
        })
      );
    },
    [isFormatAvailable, resolveTestModel, runSingleTest]
  );

  const allTasks = useMemo(
    () => selectedChannels.flatMap((channel) => selectedAPIFormats.map((format) => ({ channel, format }))),
    [selectedAPIFormats, selectedChannels]
  );
  const runnableTaskCount = useMemo(
    () => allTasks.filter(({ channel, format }) => !!resolveTestModel(channel) && isFormatAvailable(channel, format)).length,
    [allTasks, isFormatAvailable, resolveTestModel]
  );

  const handleRunAll = useCallback(async () => {
    if (runnableTaskCount === 0 || isTesting) return;

    initializeResults();
    setIsTesting(true);
    try {
      await runBatch(allTasks);
    } finally {
      setIsTesting(false);
    }
  }, [allTasks, initializeResults, isTesting, runBatch, runnableTaskCount]);

  const handleRetryFailed = useCallback(async () => {
    if (failedTasks.length === 0 || isTesting) return;

    setIsTesting(true);
    try {
      await runBatch(failedTasks);
    } finally {
      setIsTesting(false);
    }
  }, [failedTasks, isTesting, runBatch]);

  const handleRecoverChannels = useCallback(async () => {
    const ids = recoverableChannels.map((channel) => channel.id);
    if (ids.length === 0) return;

    try {
      await bulkRecoverChannels.mutateAsync(ids);
      resetRowSelection();
      setSelectedChannels([]);
      setOpen(null);
    } catch (_error) {
      // Errors are surfaced by the mutation toast.
    }
  }, [bulkRecoverChannels, recoverableChannels, resetRowSelection, setOpen, setSelectedChannels]);

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      if (!nextOpen && isTesting) return;
      setOpen(nextOpen ? 'bulkTest' : null);
    },
    [isTesting, setOpen]
  );

  const getStatusBadge = useCallback(
    (status: BulkTestStatus) => {
      switch (status) {
        case 'testing':
          return <Badge variant='secondary'>{t('channels.dialogs.bulkTest.testing')}</Badge>;
        case 'success':
          return <Badge className='border-green-200 bg-green-100 text-green-800'>{t('channels.dialogs.bulkTest.success')}</Badge>;
        case 'failed':
          return <Badge variant='destructive'>{t('channels.dialogs.bulkTest.failed')}</Badge>;
        case 'skipped':
          return <Badge variant='outline'>{t('channels.dialogs.bulkTest.skipped')}</Badge>;
        default:
          return <Badge variant='outline'>{t('channels.dialogs.bulkTest.idle')}</Badge>;
      }
    },
    [t]
  );

  if (selectedChannels.length === 0 && !isDialogOpen) return null;

  return (
    <Dialog open={isDialogOpen} onOpenChange={handleOpenChange}>
      <DialogContent ref={setDialogContent} className='flex max-h-[90vh] flex-col sm:max-w-7xl'>
        <DialogHeader className='shrink-0'>
          <DialogTitle>{t('channels.dialogs.bulkTest.title')}</DialogTitle>
          <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
            <DialogDescription>{t('channels.dialogs.bulkTest.description', { count: selectedChannels.length })}</DialogDescription>
            <ChannelTestFormatSelector value={selectedAPIFormats} onChange={setSelectedAPIFormats} portalContainer={dialogContent} />
          </div>
        </DialogHeader>

        <div className='flex min-h-0 flex-1 flex-col'>
          <div className='grid shrink-0 gap-3 border-b pb-4 md:grid-cols-4'>
            <div className='rounded-lg border bg-slate-50 p-3'>
              <div className='text-muted-foreground text-xs'>{t('channels.dialogs.bulkTest.progress', { completed: completedCount, total: resultList.length })}</div>
              <div className='mt-1 text-lg font-semibold'>{completedCount}/{resultList.length}</div>
            </div>
            <div className='rounded-lg border bg-green-50 p-3'>
              <div className='text-xs text-green-700'>{t('channels.dialogs.bulkTest.summary.success', { count: successCount })}</div>
              <div className='mt-1 text-lg font-semibold text-green-800'>{successCount}</div>
            </div>
            <div className='rounded-lg border bg-red-50 p-3'>
              <div className='text-xs text-red-700'>{t('channels.dialogs.bulkTest.summary.failed', { count: failedCount })}</div>
              <div className='mt-1 text-lg font-semibold text-red-800'>{failedCount}</div>
            </div>
            <div className='rounded-lg border bg-amber-50 p-3'>
              <div className='text-xs text-amber-700'>{t('channels.dialogs.bulkTest.summary.skipped', { count: skippedCount })}</div>
              <div className='mt-1 text-lg font-semibold text-amber-800'>{skippedCount}</div>
            </div>
          </div>

          <div className='min-h-0 flex-1 overflow-y-auto py-4'>
            <div className='overflow-x-auto rounded-lg border'>
              <Table className='min-w-max'>
                <TableHeader>
                  <TableRow>
                    <TableHead className='min-w-48'>{t('channels.dialogs.bulkTest.channelColumn')}</TableHead>
                    <TableHead className='min-w-28'>{t('channels.dialogs.bulkTest.currentStatusColumn')}</TableHead>
                    <TableHead className='min-w-48'>{t('channels.dialogs.bulkTest.testModelColumn')}</TableHead>
                    {selectedAPIFormats.map((format) => (
                      <TableHead key={format} className='min-w-52'>
                        {t(getChannelTestAPIFormat(format).labelKey)}
                      </TableHead>
                    ))}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {selectedChannels.map((channel) => (
                    <TableRow key={channel.id} className='align-top'>
                      <TableCell><TruncatedText className='block font-medium'>{channel.name}</TruncatedText></TableCell>
                      <TableCell><div className='truncate'>{t(`channels.status.${channel.status}`)}</div></TableCell>
                      <TableCell><TruncatedText className='block'>{resolveTestModel(channel) || '-'}</TruncatedText></TableCell>
                      {selectedAPIFormats.map((format) => {
                        const result = results[channel.id]?.[format];
                        const available = isFormatAvailable(channel, format);
                        const status = available && resolveTestModel(channel) ? result?.status || 'idle' : 'skipped';
                        return (
                          <TableCell key={format} className='min-w-52'>
                            <div className='space-y-2'>
                              {getStatusBadge(status)}
                              {typeof result?.latency === 'number' && <div className='text-muted-foreground text-xs'>{result.latency.toFixed(2)}s</div>}
                              {result?.status === 'testing' && <IconLoader2 className='text-muted-foreground h-4 w-4 animate-spin' />}
                              {result?.error && <ErrorDisplay error={result.error} messageClassName='block max-w-48 break-all text-xs font-medium text-red-600 whitespace-pre-wrap' />}
                              <Button
                                size='sm'
                                variant='outline'
                                onClick={() => runSingleTest(channel, format)}
                                disabled={!available || !resolveTestModel(channel) || status === 'testing' || isTesting}
                              >
                                <IconPlayerPlay className='h-3 w-3' />
                                {status === 'testing' ? t('channels.dialogs.bulkTest.testing') : t('channels.dialogs.test.testModel')}
                              </Button>
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

        <DialogFooter className='shrink-0 !flex-col gap-2 border-t pt-4 sm:!flex-row sm:items-center sm:justify-between'>
          <div className='text-muted-foreground min-w-0 text-sm'>
            {recoverableChannels.length > 0
              ? t('channels.dialogs.bulkTest.recoverButton', { count: recoverableChannels.length })
              : t('channels.dialogs.bulkTest.noRecoverableChannels')}
          </div>
          <div className='flex flex-wrap justify-end gap-2'>
            <Button variant='outline' onClick={() => handleOpenChange(false)} disabled={isTesting || bulkRecoverChannels.isPending}>
              {t('common.buttons.cancel')}
            </Button>
            <Button variant='outline' onClick={handleRetryFailed} disabled={failedTasks.length === 0 || isTesting || bulkRecoverChannels.isPending}>
              <IconRefresh className='mr-2 h-4 w-4' />
              {t('channels.dialogs.bulkTest.retryFailedButton', { count: failedTasks.length })}
            </Button>
            <Button variant='outline' onClick={handleRunAll} disabled={isTesting || bulkRecoverChannels.isPending || runnableTaskCount === 0}>
              {isTesting ? <IconLoader2 className='mr-2 h-4 w-4 animate-spin' /> : <IconFlask className='mr-2 h-4 w-4' />}
              {completedCount > 0
                ? t('channels.dialogs.bulkTest.runAgainButton')
                : t('channels.dialogs.bulkTest.runButton', { count: runnableTaskCount })}
            </Button>
            <Button onClick={handleRecoverChannels} disabled={recoverableChannels.length === 0 || isTesting || bulkRecoverChannels.isPending}>
              {bulkRecoverChannels.isPending ? <IconLoader2 className='mr-2 h-4 w-4 animate-spin' /> : <IconCheck className='mr-2 h-4 w-4' />}
              {t('channels.dialogs.bulkTest.recoverButton', { count: recoverableChannels.length })}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
