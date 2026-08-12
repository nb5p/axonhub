export function getTooltipOpenStateAfterPress(pointerType: string, wasOpen: boolean): boolean | undefined {
  if (!pointerType || pointerType === 'mouse') {
    return undefined;
  }

  return !wasOpen;
}

export function scheduleTooltipOpenStateAfterPress(pointerType: string, wasOpen: boolean, setOpen: (open: boolean) => void): boolean {
  const nextOpen = getTooltipOpenStateAfterPress(pointerType, wasOpen);
  if (nextOpen === undefined) return false;

  // Radix closes a tooltip after the trigger's click handler. Deferring the
  // touch state change ensures the requested state is applied last.
  queueMicrotask(() => setOpen(nextOpen));
  return true;
}
