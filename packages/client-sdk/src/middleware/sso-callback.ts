/**
 * SSO callback handler for token exchange flows
 */

import type { ExchangeRedeemResponse } from '../services/token-exchange';

/**
 * Create an Express handler for SSO callback endpoints
 *
 * @param redeemFn Function to redeem the exchange code
 * @param onSuccess Callback to handle successful token exchange
 * @returns Express request handler
 *
 * @example
 * ```typescript
 * import express from 'express';
 * import { createSSOCallbackHandler } from '@auth-gateway/client-sdk/middleware';
 *
 * const app = express();
 *
 * app.get('/auth/sso-callback', createSSOCallbackHandler(
 *   (code) => client.tokenExchange.redeemExchange(code),
 *   (req, res, tokens) => {
 *     res.cookie('access_token', tokens.access_token, { httpOnly: true });
 *     res.cookie('refresh_token', tokens.refresh_token, { httpOnly: true });
 *     res.redirect('/dashboard');
 *   }
 * ));
 * ```
 */
export function createSSOCallbackHandler(
  redeemFn: (code: string) => Promise<ExchangeRedeemResponse>,
  onSuccess: (req: any, res: any, tokens: ExchangeRedeemResponse) => void
) {
  return async (req: any, res: any) => {
    const code = req.query?.code as string;
    if (!code) {
      return res.status(400).json({ error: 'Missing exchange code' });
    }
    try {
      const tokens = await redeemFn(code);
      onSuccess(req, res, tokens);
    } catch (error) {
      res.status(401).json({ error: 'Invalid or expired exchange code' });
    }
  };
}
