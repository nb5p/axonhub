import { IconChevronDown } from '@tabler/icons-react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { channelTestAPIFormats, ChannelTestAPIFormat, orderChannelTestAPIFormats } from '../data/channel-test-api-formats';

interface ChannelTestFormatSelectorProps {
  value: ChannelTestAPIFormat[];
  onChange: (value: ChannelTestAPIFormat[]) => void;
  availableEndpointFormats?: ReadonlySet<string>;
  portalContainer?: HTMLDivElement | null;
}

export function ChannelTestFormatSelector({ value, onChange, availableEndpointFormats, portalContainer }: ChannelTestFormatSelectorProps) {
  const { t } = useTranslation();
  const formats = availableEndpointFormats
    ? channelTestAPIFormats.filter((format) => availableEndpointFormats.has(format.endpointFormat))
    : channelTestAPIFormats;

  const toggleFormat = (format: ChannelTestAPIFormat, checked: boolean) => {
    if (!checked && value.length === 1) {
      return;
    }

    onChange(orderChannelTestAPIFormats(checked ? [...value, format] : value.filter((item) => item !== format)));
  };

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button type='button' variant='outline' className='w-full justify-between sm:w-72'>
          {t('channels.dialogs.test.apiFormatSelection', { count: value.length })}
          <IconChevronDown className='h-4 w-4' />
        </Button>
      </PopoverTrigger>
      <PopoverContent container={portalContainer} align='end' className='w-80 space-y-1 p-2'>
        <p className='text-muted-foreground px-2 py-1 text-xs'>{t('channels.dialogs.test.apiFormatSelectionHint')}</p>
        {formats.map((format) => {
          const selected = value.includes(format.value);
          const id = `channel-test-format-${format.value}`;

          return (
            <label key={format.value} htmlFor={id} className='hover:bg-accent flex cursor-pointer items-center gap-3 rounded-sm px-2 py-2 text-sm'>
              <Checkbox id={id} checked={selected} onCheckedChange={(checked) => toggleFormat(format.value, !!checked)} disabled={selected && value.length === 1} />
              <span className='min-w-0 flex-1'>{t(format.labelKey)}</span>
            </label>
          );
        })}
      </PopoverContent>
    </Popover>
  );
}
