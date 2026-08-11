import { useEffect, useMemo, useState } from 'react';
import { SIDEBAR_NAVIGATION_GROUPS } from '@/sidebar-navigation';
import { Loader2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { useSidebarNavigationSettings, useUpdateSidebarNavigationSettings } from '../data/system';

export function SidebarNavigationSettings() {
  const { t } = useTranslation();
  const { data: settings, isLoading } = useSidebarNavigationSettings();
  const updateSettings = useUpdateSidebarNavigationSettings();
  const [hiddenItems, setHiddenItems] = useState<string[]>([]);
  const hiddenItemSet = useMemo(() => new Set(hiddenItems), [hiddenItems]);

  useEffect(() => {
    if (settings) {
      setHiddenItems(settings.hiddenItems);
    }
  }, [settings]);

  const handleVisibilityChange = async (itemID: string, visible: boolean) => {
    const previousItems = hiddenItems;
    const nextItems = visible
      ? hiddenItems.filter((hiddenItem) => hiddenItem !== itemID)
      : [...hiddenItems.filter((hiddenItem) => hiddenItem !== itemID), itemID];

    setHiddenItems(nextItems);
    try {
      await updateSettings.mutateAsync({ hiddenItems: nextItems });
    } catch {
      setHiddenItems(previousItems);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('system.sidebarNavigation.title')}</CardTitle>
        <CardDescription>{t('system.sidebarNavigation.description')}</CardDescription>
      </CardHeader>
      <CardContent className='space-y-5'>
        {isLoading ? (
          <div className='flex h-20 items-center justify-center'>
            <Loader2 className='h-5 w-5 animate-spin' />
          </div>
        ) : (
          SIDEBAR_NAVIGATION_GROUPS.map((group) => (
            <section key={group.id} className='space-y-2'>
              <h4 className='text-muted-foreground text-xs font-medium tracking-wide uppercase'>{t(group.titleKey)}</h4>
              <div className='grid gap-2 sm:grid-cols-2 xl:grid-cols-3'>
                {group.items.map((item) => {
                  const visible = !hiddenItemSet.has(item.id);
                  const switchID = `sidebar-navigation-${item.id}`;
                  const ItemIcon = item.icon;

                  return (
                    <div key={item.id} className='flex min-w-0 items-center justify-between gap-3 rounded-xl border px-3 py-2.5'>
                      <Label htmlFor={switchID} className='flex min-w-0 cursor-pointer items-center gap-2 font-normal'>
                        <ItemIcon className='text-muted-foreground size-4 shrink-0' />
                        <span className='truncate'>{t(item.titleKey)}</span>
                      </Label>
                      <Switch
                        id={switchID}
                        checked={visible}
                        onCheckedChange={(checked) => handleVisibilityChange(item.id, checked)}
                        disabled={isLoading || updateSettings.isPending}
                        aria-label={t('system.sidebarNavigation.itemVisibility', { item: t(item.titleKey) })}
                      />
                    </div>
                  );
                })}
              </div>
            </section>
          ))
        )}
        <p className='text-muted-foreground text-xs'>{t('system.sidebarNavigation.directAccessHint')}</p>
      </CardContent>
    </Card>
  );
}
