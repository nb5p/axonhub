import { useParams, useRouter } from '@tanstack/react-router';
import { Main } from '@/components/layout/main';
import { useRequest } from '../data';
import { RequestDetailContent } from './request-detail-content';
import { RequestDetailHeader } from './request-detail-header';

export default function RequestDetailGlobalPage() {
  const { requestId } = useParams({ from: '/_authenticated/requests/$requestId' });
  const router = useRouter();
  const { data: request } = useRequest(requestId, { projectId: null });

  return (
    <div className='flex h-full flex-col'>
      <RequestDetailHeader request={request} requestId={requestId} onBack={() => router.history.back()} />

      <Main className='flex-1 overflow-auto px-0 py-3 sm:px-4 sm:py-6'>
        <div className='container mx-auto max-w-7xl p-3 sm:p-6'>
          <RequestDetailContent requestId={requestId} projectId={null} />
        </div>
      </Main>
    </div>
  );
}
