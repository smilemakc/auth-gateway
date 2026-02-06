
import React, { useState, useEffect } from 'react';
import { Save, AlertTriangle, Send, Loader2 } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useSmsSettingsActive, useUpdateSmsSettings, useCreateSmsSettings } from '../hooks/useSettings';

const SmsSettings: React.FC = () => {
  const { t } = useLanguage();

  const { data: activeSettings, isLoading: loadingSettings } = useSmsSettingsActive();
  const updateMutation = useUpdateSmsSettings();
  const createMutation = useCreateSmsSettings();

  const [config, setConfig] = useState({
    provider: 'mock' as string,
    awsRegion: '',
    awsAccessKeyId: '',
    awsSecretAccessKey: '',
    twilioAccountSid: '',
    twilioAuthToken: '',
    twilioPhoneNumber: '',
  });
  const [saved, setSaved] = useState(false);
  const [testPhone, setTestPhone] = useState('');
  const [testSent, setTestSent] = useState(false);

  useEffect(() => {
    if (activeSettings) {
      setConfig({
        provider: activeSettings.provider || 'mock',
        awsRegion: activeSettings.aws_region || '',
        awsAccessKeyId: activeSettings.aws_access_key_id || '',
        awsSecretAccessKey: activeSettings.aws_secret_access_key || '',
        twilioAccountSid: activeSettings.twilio_account_sid || '',
        twilioAuthToken: activeSettings.twilio_auth_token || '',
        twilioPhoneNumber: activeSettings.twilio_phone_number || '',
      });
    }
  }, [activeSettings]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setConfig(prev => ({ ...prev, [name]: value }));
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();

    const data = {
      provider: config.provider,
      aws_region: config.awsRegion || undefined,
      aws_access_key_id: config.awsAccessKeyId || undefined,
      aws_secret_access_key: config.awsSecretAccessKey || undefined,
      twilio_account_sid: config.twilioAccountSid || undefined,
      twilio_auth_token: config.twilioAuthToken || undefined,
      twilio_phone_number: config.twilioPhoneNumber || undefined,
      is_active: true,
    };

    try {
      if (activeSettings?.id) {
        await updateMutation.mutateAsync({ id: activeSettings.id, data });
      } else {
        await createMutation.mutateAsync(data);
      }
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      console.error('Failed to save SMS settings:', err);
    }
  };

  const isLoading = updateMutation.isPending || createMutation.isPending;

  const handleTestSend = () => {
    if (!testPhone) return;
    setTestSent(true);
    setTimeout(() => setTestSent(false), 3000);
  };

  if (loadingSettings) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">{t('sms.title')}</h1>
        <p className="text-muted-foreground mt-1">{t('sms.desc')}</p>
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
             disabled={isLoading}
             className={`flex items-center gap-2 px-8 py-3 rounded-lg font-medium text-primary-foreground transition-colors
               ${saved ? 'bg-success' : 'bg-primary hover:bg-primary-600'}
             `}
           >
             {isLoading ? (
               <Loader2 size={18} className="animate-spin" />
             ) : saved ? t('common.saved') : (
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