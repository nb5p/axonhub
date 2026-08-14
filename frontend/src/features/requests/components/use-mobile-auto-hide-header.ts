import { useEffect, useRef, useState, type RefObject } from 'react';

const MOBILE_QUERY = '(max-width: 767px)';
const DIRECTION_THRESHOLD_PX = 12;
const TOP_REVEAL_THRESHOLD_PX = 16;

export function useMobileAutoHideHeader(scrollContainerRef: RefObject<HTMLElement | null>) {
  const [hidden, setHidden] = useState(false);
  const lastHandledScrollTopRef = useRef(0);

  useEffect(() => {
    const scrollContainer = scrollContainerRef.current;
    if (!scrollContainer) return;

    const mobileQuery = window.matchMedia(MOBILE_QUERY);

    const reset = () => {
      lastHandledScrollTopRef.current = scrollContainer.scrollTop;
      if (!mobileQuery.matches) setHidden(false);
    };

    const handleScroll = () => {
      const nextScrollTop = scrollContainer.scrollTop;

      if (!mobileQuery.matches || nextScrollTop <= TOP_REVEAL_THRESHOLD_PX) {
        setHidden(false);
        lastHandledScrollTopRef.current = nextScrollTop;
        return;
      }

      const delta = nextScrollTop - lastHandledScrollTopRef.current;
      if (Math.abs(delta) < DIRECTION_THRESHOLD_PX) return;

      setHidden(delta > 0);
      lastHandledScrollTopRef.current = nextScrollTop;
    };

    reset();
    scrollContainer.addEventListener('scroll', handleScroll, { passive: true });
    mobileQuery.addEventListener('change', reset);

    return () => {
      scrollContainer.removeEventListener('scroll', handleScroll);
      mobileQuery.removeEventListener('change', reset);
    };
  }, [scrollContainerRef]);

  return hidden;
}
