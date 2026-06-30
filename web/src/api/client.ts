const API_BASE = '/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(API_BASE + path, options)
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`)
  }
  return res.json()
}

export interface Account {
  id: string
  email: string
  display_name: string
  photo_url: string
  provider: string
  imap_host: string
  imap_port: number
  smtp_host: string
  smtp_port: number
  access_token: string
}

export interface QuotaInfo {
  used: number
  total: number
}

export interface Folder {
  name: string
  full_name: string
  delimiter: string
}

export interface Message {
  uid: number
  subject: string
  from: string
  date: string
  is_read: boolean
  is_starred: boolean
  has_attachments?: boolean
  size: number
  account_id?: string
}

export interface FullMessage {
  uid: number
  subject: string
  from: string
  to: string
  cc?: string
  bcc?: string
  date: string
  text_body: string
  html_body: string
  is_read: boolean
  is_starred: boolean
  size: number
  attachments?: { filename: string; size: number }[]
}

export const api = {
  getProviders: () => request<{ name: string; display_name: string }[]>('/providers'),

  connect: (provider: string) =>
    request<{ redirect_url: string }>('/accounts/connect', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ provider }),
    }),

  connectCustom: (data: any) =>
    request<{ ok: boolean; account_id: string }>('/accounts/custom', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    }),

  getAccounts: () => request<Account[]>('/accounts').catch(() => []),

  deleteAccount: (id: string) =>
    request(`/accounts/${id}`, { method: 'DELETE' }),

  getFolders: (acc: Account) =>
    request<Folder[]>('/folders', { headers: accountHeaders(acc) }),

  getMessages: (acc: Account, folder: string, limit = 50) =>
    request<{ messages: Message[]; total: number }>(
      `/messages?folder=${encodeURIComponent(folder)}&limit=${limit}`,
      { headers: accountHeaders(acc) }
    ),

  getMessage: (acc: Account, folder: string, uid: number) =>
    request<FullMessage>(
      `/message?folder=${encodeURIComponent(folder)}&uid=${uid}`,
      { headers: accountHeaders(acc) }
    ),

  markRead: (acc: Account, folder: string, uid: number) =>
    request<{ ok: boolean }>(
      `/message/mark-read?folder=${encodeURIComponent(folder)}&uid=${uid}`,
      { method: 'POST', headers: accountHeaders(acc) }
    ).catch(() => ({ ok: false })),

  markUnread: (acc: Account, folder: string, uid: number) =>
    request<{ ok: boolean }>(
      `/message/mark-unread?folder=${encodeURIComponent(folder)}&uid=${uid}`,
      { method: 'POST', headers: accountHeaders(acc) }
    ).catch(() => ({ ok: false })),

  deleteMessage: (acc: Account, folder: string, uid: number) =>
    request<{ ok: boolean }>(
      `/message?folder=${encodeURIComponent(folder)}&uid=${uid}`,
      { method: 'DELETE', headers: accountHeaders(acc) }
    ),

  getQuota: (acc: Account) =>
    request<QuotaInfo>('/quota', { headers: accountHeaders(acc) }),
}

function accountHeaders(acc: Account): Record<string, string> {
  return {
    'X-Mailly-Provider': acc.provider,
    'X-Mailly-Email': acc.email,
    'X-Mailly-IMAP-Host': acc.imap_host,
    'X-Mailly-IMAP-Port': String(acc.imap_port),
    'X-Mailly-Access-Token': acc.access_token,
  }
}
