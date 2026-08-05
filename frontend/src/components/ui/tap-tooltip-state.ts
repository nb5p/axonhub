export function getTooltipOpenStateAfterPress(pointerType: string, wasOpen: boolean): boolean | undefined {
  if (!pointerType || pointerType === 'mouse') {
    return undefined;
  }

  return !wasOpen;
}
