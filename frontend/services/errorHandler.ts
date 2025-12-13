export interface UserFriendlyError {
  title: string;
  message: string;
  type: 'error' | 'warning' | 'info';
}

export function getUserFriendlyError(error: unknown): UserFriendlyError {
  // Check error name or constructor for SDK error types
  const errorName = error && typeof error === 'object' && 'name' in error ? (error as any).name : '';
  const errorMessage = error && typeof error === 'object' && 'message' in error ? (error as any).message : '';

  // Handle SDK errors by checking error names
  if (errorName === 'TwoFactorRequiredError') {
    return {
      title: '2FA Required',
      message: 'Two-factor authentication is required for this account.',
      type: 'info',
    };
  }

  if (errorName === 'AuthenticationError') {
    return {
      title: 'Authentication Failed',
      message: 'Invalid email or password. Please try again.',
      type: 'error',
    };
  }

  if (errorName === 'AuthorizationError') {
    return {
      title: 'Access Denied',
      message: 'You do not have permission to perform this action.',
      type: 'error',
    };
  }

  if (errorName === 'ValidationError') {
    return {
      title: 'Validation Error',
      message: errorMessage || 'The provided data is invalid.',
      type: 'warning',
    };
  }

  if (errorName === 'ConflictError') {
    return {
      title: 'Conflict',
      message: errorMessage || 'This resource already exists.',
      type: 'warning',
    };
  }

  if (errorName === 'RateLimitError') {
    const retryAfter = error && typeof error === 'object' && 'retryAfter' in error ? (error as any).retryAfter : 60;
    return {
      title: 'Rate Limit Exceeded',
      message: `Too many requests. Please try again in ${retryAfter || 60} seconds.`,
      type: 'warning',
    };
  }

  if (errorName === 'NetworkError') {
    return {
      title: 'Network Error',
      message: 'Unable to connect to the server. Please check your internet connection.',
      type: 'error',
    };
  }

  if (errorName === 'TimeoutError') {
    return {
      title: 'Request Timeout',
      message: 'The request took too long. Please try again.',
      type: 'error',
    };
  }

  if (errorName && errorName.includes('Error')) {
    return {
      title: 'Error',
      message: errorMessage,
      type: 'error',
    };
  }

  // Generic error
  return {
    title: 'Unknown Error',
    message: error instanceof Error ? error.message : 'An unexpected error occurred.',
    type: 'error',
  };
}
