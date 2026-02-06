
import React, { useState } from 'react';
import { ArrowLeft, Search, AlertCircle, CheckCircle, Clock } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useLanguage } from '../services/i18n';
import { formatDateTime } from '../lib/date';

const TokenInspector: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [token, setToken] = useState('');
  const [decoded, setDecoded] = useState<{ header: any; payload: any } | null>(null);
  const [error, setError] = useState('');

  const handleDecode = (input: string) => {
    setToken(input);
    setError('');
    
    if (!input) {
      setDecoded(null);
      return;
    }

    try {
      const parts = input.split('.');
      if (parts.length !== 3) {
        throw new Error(t('inspector.invalid'));
      }

      const decodePart = (part: string) => {
        const base64 = part.replace(/-/g, '+').replace(/_/g, '/');
        const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
            return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
        }).join(''));
        return JSON.parse(jsonPayload);
      };

      const header = decodePart(parts[0]);
      const payload = decodePart(parts[1]);

      setDecoded({ header, payload });
    } catch (e) {
      setDecoded(null);
      setError(t('inspector.invalid'));
    }
  };

  const isExpired = decoded?.payload?.exp 
    ? (decoded.payload.exp * 1000) < Date.now() 
    : false;

  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/')}
          className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t('inspector.title')}</h1>
          <p className="text-muted-foreground mt-1">{t('inspector.desc')}</p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Input Column */}
        <div className="space-y-6">
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <label className="block text-sm font-medium text-muted-foreground mb-2">{t('inspector.paste')}</label>
            <textarea
              value={token}
              onChange={(e) => handleDecode(e.target.value)}
              className="w-full h-64 p-4 font-mono text-sm border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none resize-none"
              placeholder="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
            />
          </div>
          
          {error && (
            <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 flex items-center gap-3 text-destructive">
              <AlertCircle size={20} />
              {error}
            </div>
          )}

          {decoded && (
             <div className={`rounded-lg p-4 flex items-center gap-3 border ${isExpired ? 'bg-warning/10 border-warning/20 text-warning' : 'bg-success/10 border-success/20 text-success'}`}>
               {isExpired ? <Clock size={20} /> : <CheckCircle size={20} />}
               <div>
                  <p className="font-medium">{isExpired ? 'Token Expired' : 'Token Active'}</p>
                  {decoded.payload.exp && (
                    <p className="text-xs mt-1">Expires: {formatDateTime(new Date(decoded.payload.exp * 1000).toISOString())}</p>
                  )}
               </div>
             </div>
          )}
        </div>

        {/* Output Column */}
        <div className="space-y-6">
          <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
            <div className="bg-muted px-4 py-2 border-b border-border font-semibold text-muted-foreground text-sm">
              {t('inspector.header')}
            </div>
            <pre className="p-4 overflow-x-auto text-sm font-mono text-foreground bg-card">
              {decoded ? JSON.stringify(decoded.header, null, 2) : <span className="text-muted-foreground">...</span>}
            </pre>
          </div>

          <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
            <div className="bg-muted px-4 py-2 border-b border-border font-semibold text-muted-foreground text-sm">
              {t('inspector.payload')}
            </div>
            <pre className="p-4 overflow-x-auto text-sm font-mono text-foreground bg-card">
              {decoded ? JSON.stringify(decoded.payload, null, 2) : <span className="text-muted-foreground">...</span>}
            </pre>
          </div>
        </div>
      </div>
    </div>
  );
};

export default TokenInspector;
