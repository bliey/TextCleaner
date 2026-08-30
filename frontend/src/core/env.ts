export function isDesktop(): boolean {
  if (typeof window === 'undefined') return false
  const { hostname, protocol } = window.location
  return hostname === 'wails.localhost' || protocol === 'wails:'
}