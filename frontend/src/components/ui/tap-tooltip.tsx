import * as React from 'react';
import { getTooltipOpenStateAfterPress } from './tap-tooltip-state';
import { Tooltip, TooltipContent, TooltipTrigger } from './tooltip';

interface TapTooltipProps {
  children: React.ReactElement;
  content: React.ReactNode;
  contentProps?: React.ComponentProps<typeof TooltipContent>;
  onActivate?: React.MouseEventHandler;
}

function TapTooltip({ children, content, contentProps, onActivate }: TapTooltipProps) {
  const [open, setOpen] = React.useState(false);
  const pointerTypeRef = React.useRef('');
  const wasOpenOnPointerDownRef = React.useRef(false);

  const handlePointerDown = (event: React.PointerEvent) => {
    pointerTypeRef.current = event.pointerType;
    wasOpenOnPointerDownRef.current = open;
  };

  const handlePointerCancel = () => {
    pointerTypeRef.current = '';
  };

  const handleClick = (event: React.MouseEvent) => {
    const nextOpen = getTooltipOpenStateAfterPress(pointerTypeRef.current, wasOpenOnPointerDownRef.current);
    pointerTypeRef.current = '';

    if (nextOpen !== undefined) {
      setOpen(nextOpen);
      return;
    }

    onActivate?.(event);
  };

  return (
    <Tooltip open={open} onOpenChange={setOpen}>
      <TooltipTrigger
        asChild
        onPointerDown={handlePointerDown}
        onPointerCancel={handlePointerCancel}
        onClick={handleClick}
      >
        {children}
      </TooltipTrigger>
      <TooltipContent {...contentProps}>{content}</TooltipContent>
    </Tooltip>
  );
}

export { TapTooltip };
