/**
 * Shadow table sync utilities
 * Keeps a local copy of auth-gateway users in the product's database
 */

/** Database adapter interface for shadow sync */
export interface ShadowSyncDB {
  upsert(table: string, data: Record<string, any>, conflictKey: string): Promise<void>;
  deactivate(table: string, id: string): Promise<void>;
  getLastSyncTime(table: string): Promise<Date | null>;
}

/** Token validation result (matches middleware/types.ts AuthData) */
export interface SyncTokenValidation {
  userId: string;
  email: string;
  username: string;
  roles?: string[];
  appRoles?: string[];
}

/** Sync client interface (gRPC or REST) */
export interface SyncClient {
  syncUsers(opts: { updatedAfter: string; limit: number; offset: number }): Promise<{
    users: Array<{
      id: string;
      email: string;
      username: string;
      fullName?: string;
      isActive: boolean;
      updatedAt?: string;
    }>;
    total: number;
    hasMore: boolean;
    syncTimestamp: string;
  }>;
}

export interface ShadowSyncOptions {
  client: SyncClient;
  db: ShadowSyncDB;
  tableName?: string;
  syncInterval?: number;
}

export interface SyncStats {
  synced: number;
  total: number;
  duration: number;
}

export class ShadowSync {
  private client: SyncClient;
  private db: ShadowSyncDB;
  private tableName: string;
  private syncInterval: number;
  private timer: ReturnType<typeof setInterval> | null = null;

  constructor(options: ShadowSyncOptions) {
    this.client = options.client;
    this.db = options.db;
    this.tableName = options.tableName || 'users';
    this.syncInterval = options.syncInterval || 5 * 60 * 1000;
  }

  /** Sync user on login from token validation result */
  async syncOnLogin(validation: SyncTokenValidation): Promise<void> {
    await this.db.upsert(this.tableName, {
      id: validation.userId,
      email: validation.email,
      username: validation.username,
      is_active: true,
      synced_at: new Date().toISOString(),
    }, 'id');
  }

  /** Run one cycle of periodic sync */
  async periodicSync(): Promise<SyncStats> {
    const start = Date.now();
    let synced = 0;
    let total = 0;

    const lastSync = await this.db.getLastSyncTime(this.tableName);
    const updatedAfter = lastSync ? lastSync.toISOString() : '1970-01-01T00:00:00Z';

    let offset = 0;
    const limit = 100;
    let hasMore = true;

    while (hasMore) {
      const result = await this.client.syncUsers({ updatedAfter, limit, offset });
      total = result.total;

      for (const user of result.users) {
        await this.db.upsert(this.tableName, {
          id: user.id,
          email: user.email,
          username: user.username,
          full_name: user.fullName || '',
          is_active: user.isActive,
          synced_at: new Date().toISOString(),
        }, 'id');
        synced++;
      }

      hasMore = result.hasMore;
      offset += result.users.length;
    }

    return { synced, total, duration: Date.now() - start };
  }

  /** Start periodic sync */
  startPeriodicSync(): void {
    if (this.timer) return;
    this.timer = setInterval(() => {
      this.periodicSync().catch(err => console.error('[ShadowSync] periodic sync error:', err));
    }, this.syncInterval);
  }

  /** Stop periodic sync */
  stopPeriodicSync(): void {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }
}

/** Raw SQL adapter (PostgreSQL) */
export class RawSQLShadowDB implements ShadowSyncDB {
  constructor(private query: (sql: string, params: any[]) => Promise<any>) {}

  async upsert(table: string, data: Record<string, any>, conflictKey: string): Promise<void> {
    const keys = Object.keys(data);
    const values = Object.values(data);
    const placeholders = keys.map((_, i) => `$${i + 1}`);
    const updates = keys.filter(k => k !== conflictKey).map(k => `${k} = EXCLUDED.${k}`);

    const sql = `INSERT INTO ${table} (${keys.join(', ')}) VALUES (${placeholders.join(', ')}) ON CONFLICT (${conflictKey}) DO UPDATE SET ${updates.join(', ')}`;
    await this.query(sql, values);
  }

  async deactivate(table: string, id: string): Promise<void> {
    await this.query(`UPDATE ${table} SET is_active = false, synced_at = NOW() WHERE id = $1`, [id]);
  }

  async getLastSyncTime(table: string): Promise<Date | null> {
    const result = await this.query(`SELECT MAX(synced_at) as last_sync FROM ${table}`, []);
    const row = result?.rows?.[0];
    return row?.last_sync ? new Date(row.last_sync) : null;
  }
}
