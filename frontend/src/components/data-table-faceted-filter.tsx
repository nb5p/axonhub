import * as React from 'react';
import { CheckIcon, PlusCircledIcon } from '@radix-ui/react-icons';
import { Column } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList, CommandSeparator } from '@/components/ui/command';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Separator } from '@/components/ui/separator';

interface DataTableFacetedFilterProps<TData, TValue> {
  column?: Column<TData, TValue>;
  title?: string;
  options?: {
    label: string;
    value: string;
    icon?: React.ComponentType<{ className?: string }>;
  }[];
  singleSelect?: boolean;
  selectedFirst?: boolean;
  selectionSummaryThreshold?: number;
  selectionCountLabel?: (count: number) => string;
  selectionControl?: React.ReactNode;
  triggerClassName?: string;
  contentClassName?: string;
  footer?: React.ReactNode;
}

/** Renders the searchable faceted filter backed by a TanStack Table column. */
export function DataTableFacetedFilter<TData, TValue>({
  column,
  title,
  options = [],
  singleSelect = false,
  selectedFirst = false,
  selectionSummaryThreshold = 2,
  selectionCountLabel,
  selectionControl,
  triggerClassName,
  contentClassName,
  footer,
}: DataTableFacetedFilterProps<TData, TValue>) {
  const { t } = useTranslation();

  const facets = column?.getFacetedUniqueValues() || new Map();
  const filterValue = column?.getFilterValue();
  const selectedValues = singleSelect ? new Set(filterValue ? [filterValue as string] : []) : new Set((filterValue || []) as string[]);
  const orderedOptions = selectedFirst
    ? [...options].sort((a, b) => Number(selectedValues.has(b.value)) - Number(selectedValues.has(a.value)))
    : options;
  const selectedCountText = selectionCountLabel?.(selectedValues.size) ?? t('common.selectedItems', { count: selectedValues.size });

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant='outline' size='sm' className={cn('h-8 border-dashed', triggerClassName)}>
          <PlusCircledIcon className='h-4 w-4' />
          {title}
          {selectedValues?.size > 0 && (
            <>
              <Separator orientation='vertical' className='mx-2 h-4' />
              {selectionControl && (
                <>
                  {selectionControl}
                  <Separator orientation='vertical' className='mx-2 h-4' />
                </>
              )}
              <Badge variant='secondary' className='rounded-sm px-1 font-normal lg:hidden'>
                {selectionCountLabel ? selectedCountText : selectedValues.size}
              </Badge>
              <div className='hidden space-x-1 lg:flex'>
                {selectedValues.size > selectionSummaryThreshold ? (
                  <Badge variant='secondary' className='rounded-sm px-1 font-normal'>
                    {selectedCountText}
                  </Badge>
                ) : (
                  options
                    ?.filter((option) => selectedValues.has(option.value))
                    .map((option) => (
                      <Badge variant='secondary' key={option.value} className='rounded-sm px-1 font-normal'>
                        {option.label}
                      </Badge>
                    ))
                )}
              </div>
            </>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className={cn('w-[200px] p-0', contentClassName)} align='start'>
        <Command>
          <CommandInput placeholder={title} />
          <CommandList>
            <CommandEmpty>{t('common.noResultsFound')}</CommandEmpty>
            <CommandGroup>
              {orderedOptions.map((option) => {
                const isSelected = selectedValues.has(option.value);
                return (
                  <CommandItem
                    key={option.value}
                    onSelect={() => {
                      if (singleSelect) {
                        // Single select mode: set value directly or clear if already selected
                        column?.setFilterValue(isSelected ? undefined : option.value);
                      } else {
                        // Multi select mode: toggle selection
                        if (isSelected) {
                          selectedValues.delete(option.value);
                        } else {
                          selectedValues.add(option.value);
                        }
                        const filterValues = Array.from(selectedValues);
                        column?.setFilterValue(filterValues?.length ? filterValues : undefined);
                      }
                    }}
                  >
                    <div
                      className={cn(
                        'border-primary flex h-4 w-4 items-center justify-center rounded-sm border',
                        isSelected ? 'bg-primary text-primary-foreground' : 'opacity-50 [&_svg]:invisible'
                      )}
                    >
                      <CheckIcon className={cn('h-4 w-4')} />
                    </div>
                    {option.icon && <option.icon className='text-muted-foreground h-4 w-4' />}
                    <span>{option.label}</span>
                    {facets?.has(option.value) && (
                      <span className='ml-auto flex h-4 w-4 items-center justify-center font-mono text-xs'>{facets.get(option.value)}</span>
                    )}
                  </CommandItem>
                );
              })}
            </CommandGroup>
            {footer && (
              <>
                <CommandSeparator />
                <CommandGroup>{footer}</CommandGroup>
              </>
            )}
            {selectedValues.size > 0 && (
              <>
                <CommandSeparator />
                <CommandGroup>
                  <CommandItem onSelect={() => column?.setFilterValue(undefined)} className='justify-center text-center'>
                    {t('common.clearFilters')}
                  </CommandItem>
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
