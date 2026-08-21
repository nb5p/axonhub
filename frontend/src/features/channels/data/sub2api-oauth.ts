import { apiRequest } from '@/lib/api-client'

export type Sub2APIOAuthProvider = 'codex' | 'claudecode' | 'gemini' | 'grok'

export async function importSub2APIOAuthCredentials(input: {
  provider: Sub2APIOAuthProvider
  accountJSON: string
}): Promise<{ credentials: string; base_url?: string }> {
  return apiRequest('/admin/oauth/import/sub2api', {
    method: 'POST',
    body: {
      provider: input.provider,
      account_json: input.accountJSON,
    },
    requireAuth: true,
  })
}
