import type { ElementType } from 'react';
import type { LinkProps } from '@tanstack/react-router';
import {
  IconAB2,
  IconActivity,
  IconAi,
  IconBaselineDensityMedium,
  IconChartBar,
  IconDatabase,
  IconKey,
  IconLayoutDashboard,
  IconNote,
  IconPackages,
  IconRobot,
  IconSettings,
  IconShield,
  IconUsers,
} from '@tabler/icons-react';

export interface SidebarNavigationItemDefinition {
  id: string;
  titleKey: string;
  url: LinkProps['to'];
  icon: ElementType;
  mobileOnly?: boolean;
}

export interface SidebarNavigationGroupDefinition {
  id: string;
  titleKey: string;
  items: SidebarNavigationItemDefinition[];
}

export const SIDEBAR_NAVIGATION_GROUPS: SidebarNavigationGroupDefinition[] = [
  {
    id: 'admin',
    titleKey: 'sidebar.groups.admin',
    items: [
      { id: 'admin.dashboard', titleKey: 'sidebar.items.dashboard', url: '/', icon: IconLayoutDashboard },
      { id: 'admin.projects', titleKey: 'sidebar.items.projects', url: '/projects', icon: IconPackages },
      { id: 'admin.channels', titleKey: 'sidebar.items.channels', url: '/channels', icon: IconAi },
      { id: 'admin.models', titleKey: 'sidebar.items.models', url: '/models', icon: IconRobot },
      {
        id: 'admin.prompt-protection-rules',
        titleKey: 'sidebar.items.promptProtectionRules',
        url: '/prompt-protection-rules',
        icon: IconShield,
      },
      { id: 'admin.data-storages', titleKey: 'sidebar.items.dataStorages', url: '/data-storages', icon: IconDatabase },
      { id: 'admin.users', titleKey: 'sidebar.items.users', url: '/users', icon: IconUsers },
      { id: 'admin.roles', titleKey: 'sidebar.items.roles', url: '/roles', icon: IconShield },
    ],
  },
  {
    id: 'project',
    titleKey: 'sidebar.groups.project',
    items: [
      { id: 'project.api-keys', titleKey: 'sidebar.items.apiKeys', url: '/project/api-keys', icon: IconKey },
      { id: 'project.prompts', titleKey: 'sidebar.items.prompts', url: '/project/prompts', icon: IconNote },
      { id: 'project.requests', titleKey: 'sidebar.items.requests', url: '/project/requests', icon: IconActivity },
      {
        id: 'project.usage-stats',
        titleKey: 'sidebar.items.usageStats',
        url: '/project/usage-stats',
        icon: IconChartBar,
      },
      { id: 'project.traces', titleKey: 'sidebar.items.traces', url: '/project/traces', icon: IconAB2 },
      {
        id: 'project.threads',
        titleKey: 'sidebar.items.threads',
        url: '/project/threads',
        icon: IconBaselineDensityMedium,
      },
      { id: 'project.users', titleKey: 'sidebar.items.users', url: '/project/users', icon: IconUsers },
      { id: 'project.roles', titleKey: 'sidebar.items.roles', url: '/project/roles', icon: IconShield },
      { id: 'project.playground', titleKey: 'sidebar.items.playground', url: '/project/playground', icon: IconRobot },
    ],
  },
  {
    id: 'settings',
    titleKey: 'sidebar.groups.settings',
    items: [
      {
        id: 'settings.system',
        titleKey: 'sidebar.items.system',
        url: '/system',
        icon: IconSettings,
        mobileOnly: true,
      },
    ],
  },
];

export function getVisibleSidebarNavigationGroups(hiddenItems: readonly string[]): SidebarNavigationGroupDefinition[] {
  const hiddenItemSet = new Set(hiddenItems);
  return SIDEBAR_NAVIGATION_GROUPS.map((group) => ({
    ...group,
    items: group.items.filter((item) => !hiddenItemSet.has(item.id)),
  })).filter((group) => group.items.length > 0);
}
