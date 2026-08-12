import * as React from 'react';
import * as TooltipPrimitive from '@radix-ui/react-tooltip';
import { useControllableState } from '@radix-ui/react-use-controllable-state';
import { cn } from '@/lib/utils';
import { scheduleTooltipOpenStateAfterPress } from './tap-tooltip-state';

interface TooltipTouchContextValue {
  open: boolean;
  setOpen: (open: boolean) => void;
}

const TooltipTouchContext = React.createContext<TooltipTouchContextValue | null>(null);

function TooltipProvider({ delayDuration = 0, ...props }: React.ComponentProps<typeof TooltipPrimitive.Provider>) {
  return <TooltipPrimitive.Provider data-slot='tooltip-provider' delayDuration={delayDuration} {...props} />;
}

function Tooltip({ open: openProp, defaultOpen, onOpenChange, ...props }: React.ComponentProps<typeof TooltipPrimitive.Root>) {
  const [open = false, setOpen] = useControllableState({
    prop: openProp,
    defaultProp: defaultOpen ?? false,
    onChange: onOpenChange,
    caller: 'Tooltip',
  });

  return (
    <TooltipTouchContext.Provider value={{ open, setOpen }}>
      <TooltipProvider>
        <TooltipPrimitive.Root data-slot='tooltip' open={open} onOpenChange={setOpen} {...props} />
      </TooltipProvider>
    </TooltipTouchContext.Provider>
  );
}

function TooltipTrigger({ onPointerDown, onPointerCancel, onClick, ...props }: React.ComponentProps<typeof TooltipPrimitive.Trigger>) {
  const touchContext = React.useContext(TooltipTouchContext);
  const pointerTypeRef = React.useRef('');
  const wasOpenOnPointerDownRef = React.useRef(false);

  const handlePointerDown: React.PointerEventHandler<HTMLButtonElement> = (event) => {
    onPointerDown?.(event);
    if (!touchContext) return;

    pointerTypeRef.current = event.pointerType;
    wasOpenOnPointerDownRef.current = touchContext.open;
  };

  const handlePointerCancel: React.PointerEventHandler<HTMLButtonElement> = (event) => {
    onPointerCancel?.(event);
    pointerTypeRef.current = '';
  };

  const handleClick: React.MouseEventHandler<HTMLButtonElement> = (event) => {
    onClick?.(event);
    if (!touchContext) {
      pointerTypeRef.current = '';
      return;
    }

    scheduleTooltipOpenStateAfterPress(pointerTypeRef.current, wasOpenOnPointerDownRef.current, touchContext.setOpen);
    pointerTypeRef.current = '';
  };

  return (
    <TooltipPrimitive.Trigger
      data-slot='tooltip-trigger'
      onPointerDown={handlePointerDown}
      onPointerCancel={handlePointerCancel}
      onClick={handleClick}
      {...props}
    />
  );
}

function TooltipContent({ className, sideOffset = 0, children, ...props }: React.ComponentProps<typeof TooltipPrimitive.Content>) {
  return (
    <TooltipPrimitive.Portal>
      <TooltipPrimitive.Content
        data-slot='tooltip-content'
        sideOffset={sideOffset}
        className={cn(
          'bg-foreground text-background animate-in fade-in-0 zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 w-fit origin-(--radix-tooltip-content-transform-origin) rounded-md px-3 py-1.5 text-xs text-balance',
          className
        )}
        {...props}
      >
        {children}
        <TooltipPrimitive.Arrow className='bg-foreground fill-foreground z-50 size-2.5 translate-y-[calc(-50%_-_2px)] rotate-45 rounded-[2px]' />
      </TooltipPrimitive.Content>
    </TooltipPrimitive.Portal>
  );
}

export { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider };
