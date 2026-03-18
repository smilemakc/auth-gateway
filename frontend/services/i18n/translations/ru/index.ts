import common from './common';
import auth from './auth';
import nav from './nav';
import users from './users';
import applications from './applications';
import accessControl from './access-control';
import oauth from './oauth';
import email from './email';
import security from './security';
import integrations from './integrations';
import dashboard from './dashboard';

const ru: Record<string, string> = {
  ...common,
  ...auth,
  ...nav,
  ...users,
  ...applications,
  ...accessControl,
  ...oauth,
  ...email,
  ...security,
  ...integrations,
  ...dashboard,
};

export default ru;
