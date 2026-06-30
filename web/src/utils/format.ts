export function formatDateShort(dateStr: string): string {
  const d = new Date(dateStr)
  const now = new Date()
  if (d.toDateString() === now.toDateString()) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  const diff = now.getTime() - d.getTime()
  if (diff < 7 * 86400000) {
    return d.toLocaleDateString([], { weekday: 'short' })
  }
  return d.toLocaleDateString([], { month: 'short', day: 'numeric' })
}

export function formatDateTime(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleString([], {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function folderIcon(name: string): string {
  if (!name) return 'mdi-folder'
  const n = name.toLowerCase()
  if (n === 'inbox') return 'mdi-inbox'
  if (n === 'sent' || n === 'sent mail' || n.includes('sent')) return 'mdi-send'
  if (n === 'drafts' || n.includes('draft')) return 'mdi-file-document-edit'
  if (n === 'trash' || n.includes('trash')) return 'mdi-delete'
  if (n === 'spam' || n === 'junk' || n.includes('spam')) return 'mdi-alert-circle'
  if (n === 'starred' || n === 'important' || n.includes('star')) return 'mdi-star'
  if (n === 'invoices' || n.includes('invoice')) return 'mdi-file-document'
  if (n === 'newsletters' || n.includes('newsletter')) return 'mdi-email-newsletter'
  if (n === 'notifications' || n.includes('notifications') || n.includes('notification')) return 'mdi-bell'
  if (n === 'promotions' || n.includes('promotion')) return 'mdi-tag'
  if (n === 'all mail' || n.includes('all mail')) return 'mdi-archive'
  if (n === 'archive' || n.includes('archive')) return 'mdi-archive'
  if (n === 'scheduled' || n.includes('scheduled')) return 'mdi-clock'
  if (n.includes('lindy')) return 'mdi-briefcase'
  return 'mdi-folder'
}

export function avatarColor(email: string): string {
  const hash = email.split('').reduce((a, c) => ((a << 5) - a + c.charCodeAt(0)) | 0, 0)
  const colors = [
    '#e57373', '#f06292', '#ba68c8', '#9575cd', '#7986cb',
    '#64b5f6', '#4fc3f7', '#4dd0e1', '#4db6ac', '#81c784',
    '#aed581', '#dce775', '#fff176', '#ffd54f', '#ffb74d', '#ff8a65'
  ]
  return colors[Math.abs(hash) % colors.length]
}
