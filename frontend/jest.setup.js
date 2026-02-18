import '@testing-library/jest-dom';

// Mock IntersectionObserver which is used by Next.js Link component internally.
// Without this, Link triggers async state updates via requestIdleCallback that
// fire outside of act() boundaries, producing spurious warnings in every test
// that renders a page containing <Link>.
class MockIntersectionObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
Object.defineProperty(window, 'IntersectionObserver', {
  writable: true,
  value: MockIntersectionObserver,
});

// Suppress the "not wrapped in act(...)" console.error warnings that are
// false positives when using @testing-library/user-event with setup().
// React 18's development mode warns about state updates during DOM event
// dispatch even though userEvent.setup() properly wraps calls in act().
// See: https://github.com/testing-library/react-testing-library/issues/1061
const originalConsoleError = console.error;
console.error = (...args) => {
  if (typeof args[0] === 'string' && args[0].includes('was not wrapped in act')) {
    return;
  }
  originalConsoleError(...args);
};
