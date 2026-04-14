import type { MiddlewareHandler } from 'hono';
import type { AppVariables } from '../app.js';

export const requireAuth: MiddlewareHandler<{ Variables: AppVariables }> = async (c, next) => {
  const user = c.get('user');
  if (!user) {
    return c.json({ error: 'unauthorized' }, 401);
  }
  return next();
};
