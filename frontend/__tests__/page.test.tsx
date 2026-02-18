import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Home from '../app/page';

// Mock next/navigation
const mockPush = jest.fn();
const mockSearchParams = new URLSearchParams();
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
  useSearchParams: () => mockSearchParams,
}));

// Mock fetch
global.fetch = jest.fn();

describe('Home Page', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockSearchParams.delete('stage');
    mockSearchParams.delete('code');
    sessionStorage.clear();
  });

  it('renders the home page with title', () => {
    render(<Home />);
    expect(screen.getByText('No Small Parts')).toBeInTheDocument();
  });

  it('shows "Create New Stage" button when no stage code is entered', () => {
    render(<Home />);
    expect(screen.getByRole('button', { name: /create new stage/i })).toBeInTheDocument();
  });

  it('shows "Enter Stage" button when stage code is entered', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const codeInput = screen.getByPlaceholderText('ABC123');
    await user.type(codeInput, 'TEST12');

    expect(screen.getByRole('button', { name: /enter stage/i })).toBeInTheDocument();
  });

  it('validates stage code must be at least 6 characters', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const codeInput = screen.getByPlaceholderText('ABC123');
    await user.type(codeInput, 'ABC');

    expect(screen.getByText(/stage code must be at least 6 characters/i)).toBeInTheDocument();
  });

  it('extracts stage code from pasted URL', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const codeInput = screen.getByPlaceholderText('ABC123');
    await user.clear(codeInput);
    // Simulate pasting a full URL by using paste
    await user.click(codeInput);
    await user.paste('http://localhost:3000/?stage=MYCODE');

    expect(codeInput).toHaveValue('MYCODE');
  });

  it('autofocuses the name input', () => {
    render(<Home />);
    const nameInput = screen.getByPlaceholderText('Enter your stage name...');
    expect(nameInput).toHaveFocus();
  });

  it('shows error when submitting without a name', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const nameInput = screen.getByPlaceholderText('Enter your stage name...');
    const submitButton = screen.getByRole('button', { name: /create new stage/i });

    // Button should be disabled when name is empty
    expect(submitButton).toBeDisabled();
  });
});
