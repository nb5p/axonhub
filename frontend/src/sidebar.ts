import { Command } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import { useRoutePermissions } from '@/hooks/useRoutePermissions';
import { formatUserName, isCJKName } from '@/lib/utils';
import { useMe } from '@/features/auth/data/auth';
import { useSidebarNavigationSettings } from '@/features/system/data/system';
import { type SidebarData, type NavGroup, type NavLink } from './components/layout/types';
import { getVisibleSidebarNavigationGroups } from './sidebar-navigation';

export function useSidebarData(): SidebarData {
  const { t } = useTranslation();
  const { user: authUser } = useAuthStore((state) => state.auth);
  const { data: meData } = useMe();
  const { data: sidebarNavigationSettings } = useSidebarNavigationSettings();
  const { filterNavGroups } = useRoutePermissions();

  // Use data from me query if available, otherwise fall back to auth store
  const user = meData || authUser;

  // Generate user initials for avatar
  const getInitials = (firstName?: string, lastName?: string, email?: string) => {
    if (firstName && lastName) {
      const [first, second] = isCJKName(firstName, lastName) ? [lastName, firstName] : [firstName, lastName];
      return `${first.charAt(0)}${second.charAt(0)}`.toUpperCase();
    }
    if (firstName) {
      return firstName.slice(0, 2).toUpperCase();
    }
    if (email) {
      return email.split('@')[0].slice(0, 2).toUpperCase();
    }
    return 'U';
  };

  // Generate user display name
  const getDisplayName = (firstName?: string, lastName?: string, email?: string) => {
    if (firstName && lastName) {
      return formatUserName(firstName, lastName);
    }
    if (firstName) {
      return firstName;
    }
    if (email) {
      const username = email.split('@')[0];
      return username.charAt(0).toUpperCase() + username.slice(1);
    }
    return 'User';
  };

  const rawNavGroups: NavGroup[] = getVisibleSidebarNavigationGroups(sidebarNavigationSettings?.hiddenItems ?? []).map((group) => ({
    title: t(group.titleKey),
    items: group.items.map(
      (item) =>
        ({
          title: t(item.titleKey),
          url: item.url,
          icon: item.icon,
          mobileOnly: item.mobileOnly,
        }) as NavLink
    ),
  }));

  // 使用权限过滤导航组
  const filteredNavGroups = filterNavGroups(rawNavGroups).filter((group) => group.items.length > 0);

  return {
    user: {
      name: getDisplayName(user?.firstName, user?.lastName, user?.email),
      email: user?.email || 'user@example.com',
      avatar: user?.avatar || getInitials(user?.firstName, user?.lastName, user?.email),
    },
    teams: [
      {
        name: t('sidebar.team.name'),
        logo: Command,
        description: '',
        // DO NOT USE THIS
        // plan: t('sidebar.team.plan'),
      },
    ],
    navGroups: filteredNavGroups,
  };
}
