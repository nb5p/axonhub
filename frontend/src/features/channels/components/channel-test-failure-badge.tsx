import { type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';

interface ChannelTestFailureBadgeProps {
  children: ReactNode;
  error?: string;
  requestID?: string;
}

export function ChannelTestFailureBadge({ children, error, requestID }: ChannelTestFailureBadgeProps) {
  const { t } = useTranslation();

  if (!error) {
    return children;
  }

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className='inline-flex cursor-help touch-manipulation' tabIndex={0}>
          {children}
        </span>
      </TooltipTrigger>
      <TooltipContent className='max-w-md whitespace-pre-wrap break-words'>
        <p>{error}</p>
        {requestID && (
          <a
            href={`/requests/${encodeURIComponent(requestID)}`}
            className='mt-2 inline-flex text-primary underline underline-offset-2'
          >
            {t('channels.dialogs.test.viewRequestDetails')}
          </a>
        )}
      </TooltipContent>
    </Tooltip>
  );
}
