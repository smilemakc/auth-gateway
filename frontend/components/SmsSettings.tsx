
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
          className="p-2 hover:bg-white rounded-lg transition-colors text-gray-500"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{t('sms.title')}</h1>
          <p className="text-gray-500 mt-1">{t('sms.desc')}</p>
        </div>
      </div>

      <form onSubmit={handleSave} className="space-y-6">
        
        {/* Provider Selection */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">{t('sms.provider')}</h2>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            {['aws', 'twilio', 'mock'].map((p) => (
              <label 
                key={p}
                className={`cursor-pointer border rounded-xl p-4 flex flex-col items-center justify-center transition-all ${
                  config.provider === p ? 'border-blue-500 bg-blue-50 ring-2 ring-blue-200' : 'border-gray-200 hover:bg-gray-50'
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
                <span className="font-bold text-gray-900 uppercase">{p}</span>
                <span className="text-xs text-gray-500 mt-1 capitalize">
                   {p === 'mock' ? 'Testing only' : `${p} integration`}
                </span>
              </label>
            ))}
          </div>
        </div>

        {/* AWS Config */}
        {config.provider === 'aws' && (
           <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 animate-in fade-in slide-in-from-top-4 duration-300">
             <div className="flex items-center gap-2 mb-6 text-yellow-600 bg-yellow-50 p-3 rounded-lg text-sm">
                <AlertTriangle size={18} />
                Requires Amazon SNS access with SMS capabilities.
             </div>
             <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="md:col-span-2">
                   <label className="block text-sm font-medium text-gray-700 mb-1">AWS Region</label>
                   <input type="text" name="awsRegion" value={config.awsRegion || ''} onChange={handleChange} placeholder="us-east-1" className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none" />
                </div>
                <div>
                   <label className="block text-sm font-medium text-gray-700 mb-1">Access Key ID</label>
                   <input type="text" name="awsAccessKeyId" value={config.awsAccessKeyId || ''} onChange={handleChange} placeholder="AKIA..." className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none font-mono" />
                </div>
                <div>
                   <label className="block text-sm font-medium text-gray-700 mb-1">Secret Access Key</label>
                   <input type="password" name="awsSecretAccessKey" value={config.awsSecretAccessKey || ''} onChange={handleChange} placeholder="wJalrX..." className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none font-mono" />
                </div>
             </div>
           </div>
        )}

        {/* Twilio Config */}
        {config.provider === 'twilio' && (
           <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 animate-in fade-in slide-in-from-top-4 duration-300">
             <div className="grid grid-cols-1 gap-6">
                <div>
                   <label className="block text-sm font-medium text-gray-700 mb-1">Account SID</label>
                   <input type="text" name="twilioAccountSid" value={config.twilioAccountSid || ''} onChange={handleChange} placeholder="AC..." className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none font-mono" />
                </div>
                <div>
                   <label className="block text-sm font-medium text-gray-700 mb-1">Auth Token</label>
                   <input type="password" name="twilioAuthToken" value={config.twilioAuthToken || ''} onChange={handleChange} placeholder="0a1b2c..." className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none font-mono" />
                </div>
                <div>
                   <label className="block text-sm font-medium text-gray-700 mb-1">From Phone Number</label>
                   <input type="text" name="twilioPhoneNumber" value={config.twilioPhoneNumber || ''} onChange={handleChange} placeholder="+1234567890" className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none font-mono" />
                </div>
             </div>
           </div>
        )}

        {/* Test Section */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
           <h3 className="font-semibold text-gray-900 mb-4">{t('sms.test')}</h3>
           <div className="flex gap-4">
              <input 
                 type="tel" 
                 value={testPhone} 
                 onChange={(e) => setTestPhone(e.target.value)} 
                 placeholder="+1 (555) 000-0000" 
                 className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
              />
              <button 
                type="button" 
                onClick={handleTestSend}
                disabled={!testPhone}
                className="bg-gray-100 text-gray-700 px-6 py-2 rounded-lg font-medium hover:bg-gray-200 transition-colors flex items-center gap-2"
              >
                 <Send size={16} /> {testSent ? 'Sent!' : 'Send Test'}
              </button>
           </div>
        </div>

        <div className="flex justify-end pt-4">
           <button
             type="submit"
             disabled={loading}
             className={`flex items-center gap-2 px-8 py-3 rounded-lg font-medium text-white transition-colors
               ${saved ? 'bg-green-600' : 'bg-blue-600 hover:bg-blue-700'}
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