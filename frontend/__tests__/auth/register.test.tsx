import '@testing-library/jest-dom';
import { render, screen, waitFor, cleanup } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import RegisterPage from '../../app/auth/register/page';

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

describe('Registration Page', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockSearchParams.delete('invitation');
    (global.fetch as jest.Mock).mockReset();
  });

  afterEach(() => {
    cleanup();
  });

  it('renders the registration form', () => {
    render(<RegisterPage />);
    expect(screen.getByText('Create Your Account')).toBeInTheDocument();
  });

  describe('Username validation', () => {
    it('shows error when username is empty after blur', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const usernameInput = screen.getByPlaceholderText('Choose a unique username');
      await user.click(usernameInput);
      await user.tab(); // blur

      expect(screen.getByText('Username is required')).toBeInTheDocument();
    });

    it('shows error when username is less than 3 characters', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const usernameInput = screen.getByPlaceholderText('Choose a unique username');
      await user.type(usernameInput, 'ab');
      await user.tab();

      expect(screen.getByText('Must be at least 3 characters')).toBeInTheDocument();
    });

    it('does not show error for valid username', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const usernameInput = screen.getByPlaceholderText('Choose a unique username');
      await user.type(usernameInput, 'validuser');
      await user.tab();

      expect(screen.queryByText('Username is required')).not.toBeInTheDocument();
      expect(screen.queryByText('Must be at least 3 characters')).not.toBeInTheDocument();
    });
  });

  describe('Email validation', () => {
    it('shows error when email is empty after blur', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const emailInput = screen.getByPlaceholderText('you@example.com');
      await user.click(emailInput);
      await user.tab();

      expect(screen.getByText('Email is required')).toBeInTheDocument();
    });

    it('shows error for invalid email format', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const emailInput = screen.getByPlaceholderText('you@example.com');
      await user.type(emailInput, 'notanemail');
      await user.tab();

      expect(screen.getByText('Invalid email format')).toBeInTheDocument();
    });

    it('does not show error for valid email', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const emailInput = screen.getByPlaceholderText('you@example.com');
      await user.type(emailInput, 'test@example.com');
      await user.tab();

      expect(screen.queryByText('Email is required')).not.toBeInTheDocument();
      expect(screen.queryByText('Invalid email format')).not.toBeInTheDocument();
    });
  });

  describe('Password validation', () => {
    it('shows error when password is empty after blur', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const passwordInput = screen.getByPlaceholderText('Strong password required');
      await user.click(passwordInput);
      await user.tab();

      expect(screen.getByText('Password is required')).toBeInTheDocument();
    });

    it('shows error when password is less than 8 characters', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const passwordInput = screen.getByPlaceholderText('Strong password required');
      await user.type(passwordInput, 'Sh0rt!');
      await user.tab();

      expect(screen.getByText('Must be at least 8 characters')).toBeInTheDocument();
    });

    it('shows error when password lacks uppercase letter', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const passwordInput = screen.getByPlaceholderText('Strong password required');
      await user.type(passwordInput, 'password123!');
      await user.tab();

      expect(screen.getByText('Must contain at least one uppercase letter')).toBeInTheDocument();
    });

    it('shows error when password lacks number', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const passwordInput = screen.getByPlaceholderText('Strong password required');
      await user.type(passwordInput, 'Password!');
      await user.tab();

      expect(screen.getByText('Must contain at least one number')).toBeInTheDocument();
    });

    it('shows error when password lacks special character', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const passwordInput = screen.getByPlaceholderText('Strong password required');
      await user.type(passwordInput, 'Password123');
      await user.tab();

      expect(screen.getByText('Must contain at least one special character')).toBeInTheDocument();
    });

    it('does not show error for valid strong password', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const passwordInput = screen.getByPlaceholderText('Strong password required');
      await user.type(passwordInput, 'Password123!');
      await user.tab();

      expect(screen.queryByText('Password is required')).not.toBeInTheDocument();
      expect(screen.queryByText('Must be at least 8 characters')).not.toBeInTheDocument();
      expect(
        screen.queryByText('Must contain at least one uppercase letter')
      ).not.toBeInTheDocument();
      expect(screen.queryByText('Must contain at least one number')).not.toBeInTheDocument();
      expect(
        screen.queryByText('Must contain at least one special character')
      ).not.toBeInTheDocument();
    });
  });

  describe('Confirm password validation', () => {
    it('shows error when confirm password is empty after blur', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const confirmInput = screen.getByPlaceholderText('Re-enter your password');
      await user.click(confirmInput);
      await user.tab();

      expect(screen.getByText('Please confirm your password')).toBeInTheDocument();
    });

    it('shows error when passwords do not match', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const passwordInput = screen.getByPlaceholderText('Strong password required');
      const confirmInput = screen.getByPlaceholderText('Re-enter your password');

      await user.type(passwordInput, 'Password123!');
      await user.type(confirmInput, 'DifferentPass1!');
      await user.tab();

      expect(screen.getByText('Passwords do not match')).toBeInTheDocument();
    });

    it('shows success message when passwords match', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const passwordInput = screen.getByPlaceholderText('Strong password required');
      const confirmInput = screen.getByPlaceholderText('Re-enter your password');

      await user.type(passwordInput, 'Password123!');
      await user.type(confirmInput, 'Password123!');
      await user.tab();

      expect(screen.getByText('Passwords match')).toBeInTheDocument();
      expect(screen.queryByText('Passwords do not match')).not.toBeInTheDocument();
    });

    it('re-validates when password changes after confirm password was entered', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const passwordInput = screen.getByPlaceholderText('Strong password required');
      const confirmInput = screen.getByPlaceholderText('Re-enter your password');

      // First enter matching passwords
      await user.type(passwordInput, 'Password123!');
      await user.type(confirmInput, 'Password123!');
      await user.tab();

      expect(screen.getByText('Passwords match')).toBeInTheDocument();

      // Now change the password
      await user.clear(passwordInput);
      await user.type(passwordInput, 'NewPassword456!');

      // Should now show mismatch
      expect(screen.getByText('Passwords do not match')).toBeInTheDocument();
      expect(screen.queryByText('Passwords match')).not.toBeInTheDocument();
    });
  });

  describe('Name validation', () => {
    it('shows error when first name is empty after blur', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const firstNameInput = screen.getByPlaceholderText('Jane');
      await user.click(firstNameInput);
      await user.tab();

      expect(screen.getByText('First name is required')).toBeInTheDocument();
    });

    it('shows error when last name is empty after blur', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const lastNameInput = screen.getByPlaceholderText('Doe');
      await user.click(lastNameInput);
      await user.tab();

      expect(screen.getByText('Last name is required')).toBeInTheDocument();
    });
  });

  describe('Form submission', () => {
    it('shows validation error on submit with empty form', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      const submitButton = screen.getByRole('button', { name: /create account/i });
      await user.click(submitButton);

      // Should show the first validation error
      expect(screen.getByText('Username is required')).toBeInTheDocument();
    });

    it('shows error when Terms and Privacy is not accepted', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      // Fill in all required fields
      await user.type(screen.getByPlaceholderText('Choose a unique username'), 'testuser');
      await user.type(screen.getByPlaceholderText('Strong password required'), 'Password123!');
      await user.type(screen.getByPlaceholderText('Re-enter your password'), 'Password123!');
      await user.type(screen.getByPlaceholderText('you@example.com'), 'test@example.com');
      await user.type(screen.getByPlaceholderText('Jane'), 'John');
      await user.type(screen.getByPlaceholderText('Doe'), 'Doe');

      const submitButton = screen.getByRole('button', { name: /create account/i });
      await user.click(submitButton);

      expect(
        screen.getByText('You must accept the Terms of Service and Privacy Policy')
      ).toBeInTheDocument();
    });

    it('shows error when age 18 is not confirmed', async () => {
      const user = userEvent.setup();
      render(<RegisterPage />);

      // Fill in all required fields and accept Terms/Privacy but not age
      await user.type(screen.getByPlaceholderText('Choose a unique username'), 'testuser');
      await user.type(screen.getByPlaceholderText('Strong password required'), 'Password123!');
      await user.type(screen.getByPlaceholderText('Re-enter your password'), 'Password123!');
      await user.type(screen.getByPlaceholderText('you@example.com'), 'test@example.com');
      await user.type(screen.getByPlaceholderText('Jane'), 'John');
      await user.type(screen.getByPlaceholderText('Doe'), 'Doe');

      const termsCheckbox = screen.getByRole('checkbox', { name: /terms of service/i });
      await user.click(termsCheckbox);

      const submitButton = screen.getByRole('button', { name: /create account/i });
      await user.click(submitButton);

      expect(screen.getByText("You must confirm you're at least 18 years old")).toBeInTheDocument();
    });

    it('submits form when all fields are valid', async () => {
      const user = userEvent.setup();
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true }),
      });

      render(<RegisterPage />);

      // Fill in all required fields
      await user.type(screen.getByPlaceholderText('Choose a unique username'), 'testuser');
      await user.type(screen.getByPlaceholderText('Strong password required'), 'Password123!');
      await user.type(screen.getByPlaceholderText('Re-enter your password'), 'Password123!');
      await user.type(screen.getByPlaceholderText('you@example.com'), 'test@example.com');
      await user.type(screen.getByPlaceholderText('Jane'), 'John');
      await user.type(screen.getByPlaceholderText('Doe'), 'Doe');

      // Accept Terms/Privacy and age checkboxes
      const termsCheckbox = screen.getByRole('checkbox', { name: /terms of service/i });
      const ageCheckbox = screen.getByRole('checkbox', { name: /18 years old/i });
      await user.click(termsCheckbox);
      await user.click(ageCheckbox);

      const submitButton = screen.getByRole('button', { name: /create account/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(global.fetch).toHaveBeenCalledWith(
          '/api/auth/register',
          expect.objectContaining({
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
          })
        );
      });
    });

    it('redirects to login after successful registration', async () => {
      const user = userEvent.setup();
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true }),
      });

      render(<RegisterPage />);

      // Fill in all required fields
      await user.type(screen.getByPlaceholderText('Choose a unique username'), 'testuser');
      await user.type(screen.getByPlaceholderText('Strong password required'), 'Password123!');
      await user.type(screen.getByPlaceholderText('Re-enter your password'), 'Password123!');
      await user.type(screen.getByPlaceholderText('you@example.com'), 'test@example.com');
      await user.type(screen.getByPlaceholderText('Jane'), 'John');
      await user.type(screen.getByPlaceholderText('Doe'), 'Doe');

      // Accept Terms/Privacy and age checkboxes
      const termsCheckbox = screen.getByRole('checkbox', { name: /terms of service/i });
      const ageCheckbox = screen.getByRole('checkbox', { name: /18 years old/i });
      await user.click(termsCheckbox);
      await user.click(ageCheckbox);

      const submitButton = screen.getByRole('button', { name: /create account/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/auth/login');
      });
    });

    it('shows error message on failed registration', async () => {
      const user = userEvent.setup();
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        json: () => Promise.resolve({ success: false, error: 'Username already taken' }),
      });

      render(<RegisterPage />);

      // Fill in all required fields
      await user.type(screen.getByPlaceholderText('Choose a unique username'), 'testuser');
      await user.type(screen.getByPlaceholderText('Strong password required'), 'Password123!');
      await user.type(screen.getByPlaceholderText('Re-enter your password'), 'Password123!');
      await user.type(screen.getByPlaceholderText('you@example.com'), 'test@example.com');
      await user.type(screen.getByPlaceholderText('Jane'), 'John');
      await user.type(screen.getByPlaceholderText('Doe'), 'Doe');

      // Accept Terms/Privacy and age checkboxes
      const termsCheckbox = screen.getByRole('checkbox', { name: /terms of service/i });
      const ageCheckbox = screen.getByRole('checkbox', { name: /18 years old/i });
      await user.click(termsCheckbox);
      await user.click(ageCheckbox);

      const submitButton = screen.getByRole('button', { name: /create account/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('Username already taken')).toBeInTheDocument();
      });
    });
  });
});
