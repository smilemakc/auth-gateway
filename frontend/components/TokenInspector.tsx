
import React, { useState } from 'react';
import { ArrowLeft, Search, AlertCircle, CheckCircle, Clock } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useLanguage } from '../services/i18n';

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
          className="p-2 hover:bg-white rounded-lg transition-colors text-gray-500"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{t('inspector.title')}</h1>
          <p className="text-gray-500 mt-1">{t('inspector.desc')}</p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Input Column */}
        <div className="space-y-6">
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">{t('inspector.paste')}</label>
            <textarea
              value={token}
              onChange={(e) => handleDecode(e.target.value)}
              className="w-full h-64 p-4 font-mono text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none resize-none"
              placeholder="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
            />
          </div>
          
          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-center gap-3 text-red-700">
              <AlertCircle size={20} />
              {error}
            </div>
          )}

          {decoded && (
             <div className={`rounded-lg p-4 flex items-center gap-3 border ${isExpired ? 'bg-orange-50 border-orange-200 text-orange-700' : 'bg-green-50 border-green-200 text-green-700'}`}>
               {isExpired ? <Clock size={20} /> : <CheckCircle size={20} />}
               <div>
                  <p className="font-medium">{isExpired ? 'Token Expired' : 'Token Active'}</p>
                  {decoded.payload.exp && (
                    <p className="text-xs mt-1">Expires: {new Date(decoded.payload.exp * 1000).toLocaleString()}</p>
                  )}
               </div>
             </div>
          )}
        </div>

        {/* Output Column */}
        <div className="space-y-6">
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div className="bg-gray-50 px-4 py-2 border-b border-gray-100 font-semibold text-gray-700 text-sm">
              {t('inspector.header')}
            </div>
            <pre className="p-4 overflow-x-auto text-sm font-mono text-gray-800 bg-white">
              {decoded ? JSON.stringify(decoded.header, null, 2) : <span className="text-gray-400">...</span>}
            </pre>
          </div>

          <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div className="bg-gray-50 px-4 py-2 border-b border-gray-100 font-semibold text-gray-700 text-sm">
              {t('inspector.payload')}
            </div>
            <pre className="p-4 overflow-x-auto text-sm font-mono text-gray-800 bg-white">
              {decoded ? JSON.stringify(decoded.payload, null, 2) : <span className="text-gray-400">...</span>}
            </pre>
          </div>
        </div>
      </div>
    </div>
  );
};

export default TokenInspector;
