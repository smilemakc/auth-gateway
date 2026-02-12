/**
 * Token Exchange service for cross-application SSO
 */

import type { HttpClient } from '../core/http';
import { BaseService } from './base';

export interface ExchangeResponse {
  exchange_code: string;
  expires_at: string;
  redirect_url?: string;
}

export interface ExchangeRedeemResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: {
    id: string;
    email: string;
    username: string;
  };
  application_id: string;
}

export class TokenExchangeService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  async createExchange(accessToken: string, targetAppId: string): Promise<ExchangeResponse> {
    const response = await this.http.post<ExchangeResponse>('/api/auth/token/exchange', {
      access_token: accessToken,
      target_application_id: targetAppId,
    });
    return response.data;
  }

  async redeemExchange(code: string): Promise<ExchangeRedeemResponse> {
    const response = await this.http.post<ExchangeRedeemResponse>('/api/auth/token/exchange/redeem', {
      exchange_code: code,
    });
    return response.data;
  }
}
