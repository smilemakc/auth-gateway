/**
 * Webhook handler for shadow table sync
 */
import { createHmac, timingSafeEqual } from 'crypto';
import type { ShadowSyncDB } from './shadow-sync';

export interface WebhookHandlerOptions {
  secret: string;
  db: ShadowSyncDB;
  tableName?: string;
  onUserCreated?: (data: any) => Promise<void>;
  onUserUpdated?: (data: any) => Promise<void>;
  onUserDeactivated?: (data: any) => Promise<void>;
}

export interface WebhookEvent {
  id: string;
  type: string;
  timestamp: string;
  data: Record<string, any>;
}

function verifySignature(body: string | Buffer, secret: string, signature: string): boolean {
  if (!signature || !signature.startsWith('sha256=')) return false;
  const expected = createHmac('sha256', secret).update(body).digest('hex');
  const received = signature.slice(7);
  if (expected.length !== received.length) return false;
  return timingSafeEqual(Buffer.from(expected), Buffer.from(received));
}

/** Create Express-compatible webhook handler */
export function createWebhookHandler(options: WebhookHandlerOptions) {
  const table = options.tableName || 'users';

  return async (req: any, res: any) => {
    try {
      const signature = req.headers['x-webhook-signature'] as string;
      const rawBody = typeof req.body === 'string' ? req.body : JSON.stringify(req.body);

      if (!verifySignature(rawBody, options.secret, signature)) {
        return res.status(401).json({ error: 'Invalid signature' });
      }

      const event: WebhookEvent = typeof req.body === 'string' ? JSON.parse(req.body) : req.body;

      switch (event.type) {
        case 'user.created':
          await options.db.upsert(table, {
            id: event.data.user_id || event.data.id,
            email: event.data.email || '',
            username: event.data.username || '',
            full_name: event.data.full_name || '',
            is_active: true,
            synced_at: new Date().toISOString(),
          }, 'id');
          await options.onUserCreated?.(event.data);
          break;

        case 'user.updated':
          await options.db.upsert(table, {
            id: event.data.user_id || event.data.id,
            email: event.data.email || '',
            username: event.data.username || '',
            full_name: event.data.full_name || '',
            is_active: event.data.is_active !== false,
            synced_at: new Date().toISOString(),
          }, 'id');
          await options.onUserUpdated?.(event.data);
          break;

        case 'user.blocked':
        case 'user.deleted':
          await options.db.deactivate(table, event.data.user_id || event.data.id);
          await options.onUserDeactivated?.(event.data);
          break;
      }

      res.json({ status: 'ok' });
    } catch (err) {
      console.error('[WebhookHandler] error:', err);
      res.status(500).json({ error: 'Internal error' });
    }
  };
}
