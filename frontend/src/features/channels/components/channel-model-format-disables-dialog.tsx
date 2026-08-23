'use client';

import { useEffect, useMemo, useState } from 'react';
import { IconCircleOff, IconPlayerPlay } from '@tabler/icons-react';
import { useTranslation } from 'react-i18next';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { channelTestAPIFormats } from '../data/channel-test-api-formats';
import { useUpdateChannel } from '../data/channels';
import { Channel } from '../data/schema';
import { mergeChannelSettingsForUpdate } from '../utils/merge';
import { enableModelAPIFormat } from '../utils/model-api-format-disables';

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  channel: Channel;
}

export function ChannelModelFormatDisablesDialog({ open, onOpenChange, channel }: Props) {
  const { t } = useTranslation();
  const updateChannel = useUpdateChannel();
  const [disabledModelAPIFormats, setDisabledModelAPIFormats] = useState(channel.settings?.disabledModelApiFormats ?? []);

  useEffect(() => {
    if (open) {
      setDisabledModelAPIFormats(channel.settings?.disabledModelApiFormats ?? []);
    }
  }, [channel.settings?.disabledModelApiFormats, open]);

  const formatLabels = useMemo(() => new Map(channelTestAPIFormats.map((format) => [format.endpointFormat, t(format.labelKey)])), [t]);

  const handleEnable = async (model: string, endpointFormat: string, clients?: string[] | null) => {
    const nextDisabledModelAPIFormats = enableModelAPIFormat(disabledModelAPIFormats, model, endpointFormat, clients ?? undefined);

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
    } catch {
      // Errors are handled by useUpdateChannel toast.
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-3xl'>
        <DialogHeader>
          <DialogTitle>{t('channels.dialogs.modelFormatDisables.title')}</DialogTitle>
          <DialogDescription>{t('channels.dialogs.modelFormatDisables.description', { name: channel.name })}</DialogDescription>
        </DialogHeader>

        {disabledModelAPIFormats.length === 0 ? (
          <div className='flex flex-col items-center gap-2 rounded-lg border border-dashed px-6 py-12 text-center'>
            <IconCircleOff className='text-muted-foreground h-8 w-8' />
            <p className='text-muted-foreground text-sm'>{t('channels.dialogs.modelFormatDisables.empty')}</p>
          </div>
        ) : (
          <div className='max-h-[55vh] overflow-auto rounded-lg border'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('channels.dialogs.modelFormatDisables.model')}</TableHead>
                  <TableHead>{t('channels.dialogs.modelFormatDisables.clients')}</TableHead>
                  <TableHead>{t('channels.dialogs.modelFormatDisables.apiFormats')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {disabledModelAPIFormats.map(({ model, apiFormats, clients }) => (
                  <TableRow key={`${model}:${(clients ?? []).join(',')}`}>
                    <TableCell className='align-top font-medium'>{model}</TableCell>
                    <TableCell className='align-top'>
                      <Badge variant='secondary'>
                        {(clients?.length ?? 0) === 0
                          ? t('channels.dialogs.modelFormatDisables.allClients')
                          : clients?.map((client) => t(`channels.dialogs.modelFormatDisables.client.${client}`)).join(', ')}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className='flex flex-wrap gap-2'>
                        {apiFormats.map((endpointFormat) => (
                          <Badge key={endpointFormat} variant='outline' className='h-8 gap-1.5 pr-1 pl-2.5'>
                            {formatLabels.get(endpointFormat) ?? endpointFormat}
                            <Button
                              type='button'
                              variant='ghost'
                              size='icon'
                              className='h-6 w-6'
                              onClick={() => handleEnable(model, endpointFormat, clients)}
                              disabled={updateChannel.isPending}
                              aria-label={t('channels.dialogs.modelFormatDisables.enable')}
                            >
                              <IconPlayerPlay className='h-3.5 w-3.5' />
                            </Button>
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
