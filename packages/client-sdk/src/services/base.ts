/**
 * Base service class for all API services
 */

import type { HttpClient } from '../core/http';

/** Base service that all API services extend */
export abstract class BaseService {
  protected readonly http: HttpClient;

  constructor(http: HttpClient) {
    this.http = http;
  }
}
