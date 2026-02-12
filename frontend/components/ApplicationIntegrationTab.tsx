import React, { useState } from 'react';
import { Copy, Check } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { Application } from '../types';

interface ApplicationIntegrationTabProps {
  application: Application;
}

const ApplicationIntegrationTab: React.FC<ApplicationIntegrationTabProps> = ({ application }) => {
  const { t } = useLanguage();
  const [copiedField, setCopiedField] = useState<string | null>(null);
  const [activeCodeTab, setActiveCodeTab] = useState<'curl' | 'go' | 'typescript'>('curl');

  const copyToClipboard = (text: string, field: string) => {
    navigator.clipboard.writeText(text);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };

  const authConfigJson = JSON.stringify(
    {
      application_id: application.id,
      name: application.name,
      display_name: application.display_name,
      is_active: application.is_active,
      callback_urls: application.callback_urls,
    },
    null,
    2
  );

  const curlExample = `# Get auth config (authenticate with app secret)
curl -X GET 'https://your-auth-gateway.com/api/applications/config' \\
  -H 'Authorization: Bearer YOUR_APP_SECRET'

# Sign in a user within this application
curl -X POST 'https://your-auth-gateway.com/api/auth/signin' \\
  -H 'Content-Type: application/json' \\
  -H 'X-Application-ID: ${application.id}' \\
  -d '{"email": "user@example.com", "password": "secret"}'`;

  const goExample = `package main

import (
    "context"
    "fmt"
    authgateway "github.com/smilemakc/auth-gateway/packages/go-sdk"
)

func main() {
    client := authgateway.NewClient(authgateway.Config{
        BaseURL: "https://your-auth-gateway.com",
        Headers: map[string]string{
            "X-Application-ID": "${application.id}",
        },
    })

    // Sign in a user
    resp, err := client.Auth.SignIn(context.Background(), &authgateway.SignInRequest{
        Email:    "user@example.com",
        Password: "secret",
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("User: %s\\n", resp.User.Email)
}`;

  const typescriptExample = `import { createClient } from '@auth-gateway/client-sdk';

const client = createClient({
  baseUrl: 'https://your-auth-gateway.com',
  headers: {
    'X-Application-ID': '${application.id}',
  },
});

// Sign in a user
const { user, accessToken } = await client.auth.signIn({
  email: 'user@example.com',
  password: 'secret',
});

console.log('User:', user.email);`;

  return (
    <div className="space-y-6">
      {/* Section 1: Credentials */}
      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('apps.integration.credentials')}</h2>
        <div className="space-y-4">
          <div>
            <dt className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">
              Application ID
            </dt>
            <dd className="flex items-center gap-2">
              <code className="flex-1 bg-muted rounded px-3 py-2 text-sm text-foreground font-mono border border-border truncate">
                {application.id}
              </code>
              <button
                onClick={() => copyToClipboard(application.id, 'app-id')}
                className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-accent rounded transition-colors"
                title={copiedField === 'app-id' ? t('apps.integration.copied') : 'Copy'}
              >
                {copiedField === 'app-id' ? <Check size={14} /> : <Copy size={14} />}
              </button>
            </dd>
          </div>

          <div>
            <dt className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">
              {t('apps.integration.secret_prefix')}
            </dt>
            <dd className="flex items-center gap-2">
              <code className="flex-1 bg-muted rounded px-3 py-2 text-sm text-muted-foreground font-mono border border-border">
                {application.secret_prefix ? `${application.secret_prefix}••••••••` : '—'}
              </code>
            </dd>
          </div>
        </div>
      </div>

      {/* Section 2: Auth Config (API Response Preview) */}
      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <h2 className="text-lg font-semibold text-foreground mb-2">{t('apps.integration.auth_config')}</h2>
        <p className="text-sm text-muted-foreground mb-4">{t('apps.integration.auth_config_desc')}</p>
        <div className="bg-muted rounded-lg p-4 overflow-x-auto">
          <pre className="text-sm font-mono text-foreground">
            <code>{authConfigJson}</code>
          </pre>
        </div>
      </div>

      {/* Section 3: Quick Start Code Examples */}
      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('apps.integration.quick_start')}</h2>

        {/* Tab selector */}
        <div className="flex gap-2 mb-4 border-b border-border">
          <button
            onClick={() => setActiveCodeTab('curl')}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeCodeTab === 'curl'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            cURL
          </button>
          <button
            onClick={() => setActiveCodeTab('go')}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeCodeTab === 'go'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            Go
          </button>
          <button
            onClick={() => setActiveCodeTab('typescript')}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeCodeTab === 'typescript'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            TypeScript
          </button>
        </div>

        {/* Code blocks */}
        <div className="relative">
          <button
            onClick={() => {
              const code = activeCodeTab === 'curl' ? curlExample : activeCodeTab === 'go' ? goExample : typescriptExample;
              copyToClipboard(code, 'code');
            }}
            className="absolute top-2 right-2 p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded transition-colors z-10"
            title={copiedField === 'code' ? t('apps.integration.copied') : 'Copy code'}
          >
            {copiedField === 'code' ? <Check size={16} /> : <Copy size={16} />}
          </button>

          {activeCodeTab === 'curl' && (
            <div className="bg-muted rounded-lg p-4 text-sm font-mono overflow-x-auto">
              <pre className="text-foreground">{curlExample}</pre>
            </div>
          )}

          {activeCodeTab === 'go' && (
            <div className="bg-muted rounded-lg p-4 text-sm font-mono overflow-x-auto">
              <pre className="text-foreground">{goExample}</pre>
            </div>
          )}

          {activeCodeTab === 'typescript' && (
            <div className="bg-muted rounded-lg p-4 text-sm font-mono overflow-x-auto">
              <pre className="text-foreground">{typescriptExample}</pre>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ApplicationIntegrationTab;
