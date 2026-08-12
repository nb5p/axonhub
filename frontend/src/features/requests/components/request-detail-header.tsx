import { format } from 'date-fns';
import { ArrowLeft, Copy, FileText } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { extractNumberID } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { Header } from '@/components/layout/header';
import type { Request } from '../data';

interface RequestDetailHeaderProps {
  request?: Request | null;
  requestId: string;
  onBack: () => void;
}

export function RequestDetailHeader({ request, requestId, onBack }: RequestDetailHeaderProps) {
  const { t } = useTranslation();
  const visibleRequestId = request ? extractNumberID(request.id) || request.id : extractNumberID(requestId) || requestId;

  const copyRequestID = async () => {
    try {
      if (!navigator.clipboard) throw new Error('Clipboard unavailable');
      await navigator.clipboard.writeText(request?.id ?? requestId);
      toast.success(t('requests.actions.copied'));
    } catch {
      toast.error(t('common.errors.copyFailed'));
    }
  };

  return (
    <Header className='bg-background/95 supports-[backdrop-filter]:bg-background/60 h-auto min-h-16 items-stretch border-b p-0 backdrop-blur sm:h-16 sm:items-center sm:p-4'>
      <div className='flex w-full min-w-0 flex-col sm:flex-row sm:items-center sm:gap-4'>
        <div className='flex h-12 shrink-0 items-center border-b px-2 sm:h-auto sm:border-0 sm:px-0'>
          <Button variant='ghost' size='sm' onClick={onBack} className='hover:bg-accent min-h-12 px-3 sm:min-h-8'>
            <ArrowLeft className='mr-2 h-4 w-4' />
            {t('common.back')}
          </Button>
          <Separator orientation='vertical' className='ml-2 hidden h-6 sm:block' />
        </div>

        <div className='flex min-w-0 items-start gap-3 px-4 py-3 sm:items-center sm:p-0'>
          <div className='bg-primary/10 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg sm:h-8 sm:w-8'>
            <FileText className='text-primary h-4 w-4' />
          </div>
          <div className='min-w-0 flex-1'>
            <div className='flex min-w-0 items-center gap-1'>
              <h1 className='min-w-0 truncate text-base leading-tight font-semibold sm:text-lg'>
                {t('requests.detail.title')} #{visibleRequestId}
              </h1>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant='ghost'
                    size='icon-sm'
                    className='h-12 w-12 shrink-0 sm:h-7 sm:w-7'
                    onClick={() => void copyRequestID()}
                    aria-label={t('requests.actions.copyRequestId')}
                  >
                    <Copy className='h-4 w-4 sm:h-3.5 sm:w-3.5' />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{t('requests.actions.copyRequestId')}</TooltipContent>
              </Tooltip>
            </div>
            {request && (
              <div className='mt-1 flex min-w-0 flex-col gap-0.5 sm:flex-row sm:items-center sm:gap-2'>
                <p className='text-muted-foreground min-w-0 text-sm break-all'>{request.modelID || t('requests.columns.unknown')}</p>
                <span className='text-muted-foreground hidden text-xs sm:inline'>•</span>
                <p className='text-muted-foreground text-xs whitespace-nowrap'>
                  {format(new Date(request.createdAt), 'yyyy-MM-dd HH:mm:ss')}
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </Header>
  );
}
