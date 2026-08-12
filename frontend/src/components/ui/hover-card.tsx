'use client';

import * as React from 'react';
import * as HoverCardPrimitive from '@radix-ui/react-hover-card';
import { useControllableState } from '@radix-ui/react-use-controllable-state';
import { cn } from '@/lib/utils';
import { getTooltipOpenStateAfterPress } from './tap-tooltip-state';

interface HoverCardTouchContextValue {
  open: boolean;
  setOpen: (open: boolean) => void;
}

const HoverCardTouchContext = React.createContext<HoverCardTouchContextValue | null>(null);

function HoverCard({ open: openProp, defaultOpen, onOpenChange, ...props }: React.ComponentProps<typeof HoverCardPrimitive.Root>) {
  const [open = false, setOpen] = useControllableState({
    prop: openProp,
    defaultProp: defaultOpen ?? false,
    onChange: onOpenChange,
    caller: 'HoverCard',
  });

  return (
    <HoverCardTouchContext.Provider value={{ open, setOpen }}>
      <HoverCardPrimitive.Root data-slot='hover-card' open={open} onOpenChange={setOpen} {...props} />
    </HoverCardTouchContext.Provider>
  );
}

function HoverCardTrigger({ onPointerDown, ...props }: React.ComponentProps<typeof HoverCardPrimitive.Trigger>) {
  const touchContext = React.useContext(HoverCardTouchContext);

  const handlePointerDown: React.PointerEventHandler<HTMLAnchorElement> = (event) => {
    onPointerDown?.(event);
    if (!touchContext) return;

    const nextOpen = getTooltipOpenStateAfterPress(event.pointerType, touchContext.open);
    if (nextOpen !== undefined) touchContext.setOpen(nextOpen);
  };

  return <HoverCardPrimitive.Trigger data-slot='hover-card-trigger' onPointerDown={handlePointerDown} {...props} />;
}

function HoverCardContent({
  className,
  align = 'center',
  sideOffset = 4,
  ...props
}: React.ComponentProps<typeof HoverCardPrimitive.Content>) {
  return (
    <HoverCardPrimitive.Portal data-slot='hover-card-portal'>
      <HoverCardPrimitive.Content
        data-slot='hover-card-content'
        align={align}
        sideOffset={sideOffset}
        className={cn(
          'bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 w-64 origin-(--radix-hover-card-content-transform-origin) rounded-md border p-4 shadow-md outline-hidden',
          className
        )}
        {...props}
      />
    </HoverCardPrimitive.Portal>
  );
}

export { HoverCard, HoverCardTrigger, HoverCardContent };
