
import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, MessageSquare, AlertTriangle, Send } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { getSmsConfig, updateSmsConfig } from '../services/mockData';
import { SmsConfig, SmsProviderType } from '../types';

const SmsSettings: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [config, setConfig] = useState<SmsConfig>({ provider: 'mock' });
  const [loading, setLoading] = useState(false);
  const [saved, setSaved] = useState(false);
  const [testPhone, setTestPhone] = useState('');
  const [testSent, setTestSent] = useState(false);

  useEffect(() => {
    setConfig(getSmsConfig());
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setConfig(prev => ({ ...prev, [name]: value }));
  };

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setTimeout(() => {
      updateSmsConfig(config);
      setLoading(false);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    }, 800);
  };

  const handleTestSend = () => {
    if (!testPhone) return;
    setTestSent(true);
    setTimeout(() => setTestSent(false), 3000);
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/settings')}
          className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t('sms.title')}</h1>
          <p className="text-muted-foreground mt-1">{t('sms.desc')}</p>
        </div>
      </div>

      <form onSubmit={handleSave} className="space-y-6">
        
        {/* Provider Selection */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">{t('sms.provider')}</h2>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            {['aws', 'twilio', 'mock'].map((p) => (
              <label
                key={p}
                className={`cursor-pointer border rounded-xl p-4 flex flex-col items-center justify-center transition-all ${
                  config.provider === p ? 'border-primary bg-primary/10 ring-2 ring-ring' : 'border-border hover:bg-accent'
                }`}
              >
                <input
                  type="radio"
                  name="provider"
                  value={p}
                  checked={config.provider === p}
                  onChange={handleChange}
                  className="sr-only"
                />
                <span className="font-bold text-foreground uppercase">{p}</span>
                <span className="text-xs text-muted-foreground mt-1 capitalize">
                   {p === 'mock' ? 'Testing only' : `${p} integration`}
                </span>
              </label>
            ))}
          </div>
        </div>

        {/* AWS Config */}
        {config.provider === 'aws' && (
           <div className="bg-card rounded-xl shadow-sm border border-border p-6 animate-in fade-in slide-in-from-top-4 duration-300">
             <div className="flex items-center gap-2 mb-6 text-warning bg-warning/10 p-3 rounded-lg text-sm">
                <AlertTriangle size={18} />
                Requires Amazon SNS access with SMS capabilities.
             </div>
             <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="md:col-span-2">
                   <label className="block text-sm font-medium text-foreground mb-1">AWS Region</label>
                   <input type="text" name="awsRegion" value={config.awsRegion || ''} onChange={handleChange} placeholder="us-east-1" className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none" />
                </div>
                <div>
                   <label className="block text-sm font-medium text-foreground mb-1">Access Key ID</label>
                   <input type="text" name="awsAccessKeyId" value={config.awsAccessKeyId || ''} onChange={handleChange} placeholder="AKIA..." className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none font-mono" />
                </div>
                <div>
                   <label className="block text-sm font-medium text-foreground mb-1">Secret Access Key</label>
                   <input type="password" name="awsSecretAccessKey" value={config.awsSecretAccessKey || ''} onChange={handleChange} placeholder="wJalrX..." className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none font-mono" />
                </div>
             </div>
           </div>
        )}

        {/* Twilio Config */}
        {config.provider === 'twilio' && (
           <div className="bg-card rounded-xl shadow-sm border border-border p-6 animate-in fade-in slide-in-from-top-4 duration-300">
             <div className="grid grid-cols-1 gap-6">
                <div>
                   <label className="block text-sm font-medium text-foreground mb-1">Account SID</label>
                   <input type="text" name="twilioAccountSid" value={config.twilioAccountSid || ''} onChange={handleChange} placeholder="AC..." className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none font-mono" />
                </div>
                <div>
                   <label className="block text-sm font-medium text-foreground mb-1">Auth Token</label>
                   <input type="password" name="twilioAuthToken" value={config.twilioAuthToken || ''} onChange={handleChange} placeholder="0a1b2c..." className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none font-mono" />
                </div>
                <div>
                   <label className="block text-sm font-medium text-foreground mb-1">From Phone Number</label>
                   <input type="text" name="twilioPhoneNumber" value={config.twilioPhoneNumber || ''} onChange={handleChange} placeholder="+1234567890" className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none font-mono" />
                </div>
             </div>
           </div>
        )}

        {/* Test Section */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
           <h3 className="font-semibold text-foreground mb-4">{t('sms.test')}</h3>
           <div className="flex gap-4">
              <input
                 type="tel"
                 value={testPhone}
                 onChange={(e) => setTestPhone(e.target.value)}
                 placeholder="+1 (555) 000-0000"
                 className="flex-1 px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
              />
              <button
                type="button"
                onClick={handleTestSend}
                disabled={!testPhone}
                className="bg-muted text-foreground px-6 py-2 rounded-lg font-medium hover:bg-accent transition-colors flex items-center gap-2"
              >
                 <Send size={16} /> {testSent ? 'Sent!' : 'Send Test'}
              </button>
           </div>
        </div>

        <div className="flex justify-end pt-4">
           <button
             type="submit"
             disabled={loading}
             className={`flex items-center gap-2 px-8 py-3 rounded-lg font-medium text-primary-foreground transition-colors
               ${saved ? 'bg-success' : 'bg-primary hover:bg-primary-600'}
             `}
           >
             {loading ? t('common.saving') : saved ? t('common.saved') : (
               <>
                 <Save size={18} /> {t('common.save')}
               </>
             )}
           </button>
        </div>

      </form>
    </div>
  );
};

export default SmsSettings;