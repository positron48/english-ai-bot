// Shared definition of the 5 public nav tabs (prototype: SideNav/BottomNav)
export interface NavTab {
  id: string
  labelKey: string
  icon: string
  to: string
}

export const NAV_TABS: NavTab[] = [
  { id: 'home', labelKey: 'lg.nav.home', icon: 'home', to: '/dashboard' },
  { id: 'city', labelKey: 'lg.nav.city', icon: 'map-pin', to: '/city' },
  { id: 'practice', labelKey: 'lg.nav.practice', icon: 'dumbbell', to: '/learning' },
  { id: 'progress', labelKey: 'lg.nav.progress', icon: 'bar', to: '/progress' },
  { id: 'profile', labelKey: 'lg.nav.profile', icon: 'user', to: '/settings' },
]
