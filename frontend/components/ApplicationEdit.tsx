import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import {
  ArrowLeft, Loader2, Plus, X, Boxes, ToggleLeft, ToggleRight,
  KeyRound, Mail, Smartphone, Chrome, Github, Globe, Send, ShieldCheck, Key,
} from 'lucide-react';
import { useLanguage } from '../services/i18n';
import {
  useApplicationDetail,
  useCreateApplication,
  useUpdateApplication,
  useDeleteApplication,
} from '../hooks/useApplications';
import { confirm } from '../services/confirm';

const AUTH_METHODS = [
  { value: 'password', label: 'apps.auth_methods.password', icon: KeyRound },
  { value: 'otp_email', label: 'apps.auth_methods.otp_email', icon: Mail },
  { value: 'otp_sms', label: 'apps.auth_methods.otp_sms', icon: Smartphone },
  { value: 'oauth_google', label: 'apps.auth_methods.oauth_google', icon: Chrome },
  { value: 'oauth_github', label: 'apps.auth_methods.oauth_github', icon: Github },
  { value: 'oauth_yandex', label: 'apps.auth_methods.oauth_yandex', icon: Globe },
  { value: 'oauth_telegram', label: 'apps.auth_methods.oauth_telegram', icon: Send },
  { value: 'totp', label: 'apps.auth_methods.totp', icon: ShieldCheck },
  { value: 'api_key', label: 'apps.auth_methods.api_key', icon: Key },
] as const;

const ApplicationEdit: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();

  const isEditMode = id && id !== 'new';

  const { data: application, isLoading: isLoadingApp } = useApplicationDetail(id || '');
  const createApplication = useCreateApplication();
  const updateApplication = useUpdateApplication();
  const deleteApplication = useDeleteApplication();

  const [formData, setFormData] = useState({
    name: '',
    display_name: '',
    description: '',
    homepage_url: '',
    callback_urls: [''],
    allowed_auth_methods: ['password'] as string[],
    is_active: true,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (application) {
      setFormData({
        name: application.name || '',
        display_name: application.display_name || '',
        description: application.description || '',
        homepage_url: application.homepage_url || '',
        callback_urls: application.callback_urls?.length > 0 ? application.callback_urls : [''],
        allowed_auth_methods: application.allowed_auth_methods?.length
          ? application.allowed_auth_methods
          : ['password'],
        is_active: application.is_active ?? true,
      });
    }
  }, [application]);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = t('apps.error.name_required');
    } else if (!/^[a-z0-9-]+$/.test(formData.name)) {
      newErrors.name = t('apps.error.name_format');
    }

    if (!formData.display_name.trim()) {
      newErrors.display_name = t('apps.error.display_name_required');
    }

    if (formData.allowed_auth_methods.length === 0) {
      newErrors.allowed_auth_methods = t('apps.auth_methods.error_empty');
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) return;

    const filteredCallbackUrls = formData.callback_urls.filter(url => url.trim() !== '');

    try {
      if (isEditMode) {
        await updateApplication.mutateAsync({
          id: id!,
          data: {
            display_name: formData.display_name,
            description: formData.description || undefined,
            homepage_url: formData.homepage_url || undefined,
            callback_urls: filteredCallbackUrls,
            allowed_auth_methods: formData.allowed_auth_methods,
            is_active: formData.is_active,
          },
        });
        navigate(`/applications/${id}`);
      } else {
        const result = await createApplication.mutateAsync({
          name: formData.name,
          display_name: formData.display_name,
          description: formData.description || undefined,
          homepage_url: formData.homepage_url || undefined,
          callback_urls: filteredCallbackUrls,
          allowed_auth_methods: formData.allowed_auth_methods,
        });
        navigate(`/applications/${result.application.id}`, {
          state: { secret: result.secret },
        });
      }
    } catch (error) {
      console.error('Failed to save application:', error);
    }
  };

  const handleDelete = async () => {
    const ok = await confirm({
      title: t('confirm.delete_title'),
      description: t('common.confirm_delete'),
      variant: 'danger'
    });
    if (ok) {
      try {
        await deleteApplication.mutateAsync(id!);
        navigate('/applications');
      } catch (error) {
        console.error('Failed to delete application:', error);
      }
    }
  };

  const addCallbackUrl = () => {
    setFormData(prev => ({
      ...prev,
      callback_urls: [...prev.callback_urls, ''],
    }));
  };

  const removeCallbackUrl = (index: number) => {
    setFormData(prev => ({
      ...prev,
      callback_urls: prev.callback_urls.filter((_, i) => i !== index),
    }));
  };

  const updateCallbackUrl = (index: number, value: string) => {
    setFormData(prev => ({
      ...prev,
      callback_urls: prev.callback_urls.map((url, i) => (i === index ? value : url)),
    }));
  };

  const toggleAuthMethod = (method: string) => {
    setFormData(prev => ({
      ...prev,
      allowed_auth_methods: prev.allowed_auth_methods.includes(method)
        ? prev.allowed_auth_methods.filter(m => m !== method)
        : [...prev.allowed_auth_methods, method],
    }));
  };

  if (isEditMode && isLoadingApp) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  const isPending = createApplication.isPending || updateApplication.isPending;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          to={isEditMode ? `/applications/${id}` : '/applications'}
          className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
        >
          <ArrowLeft size={20} />
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-foreground">
            {isEditMode ? t('apps.edit') : t('apps.create')}
          </h1>
          <p className="text-muted-foreground mt-1">
            {isEditMode
              ? t('apps.edit_desc')
              : t('apps.create_desc')}
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
              <Boxes className="text-primary" size={20} />
            </div>
            <h2 className="text-lg font-semibold text-foreground">{t('apps.basic_info')}</h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('apps.name')} <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={e => setFormData(prev => ({ ...prev, name: e.target.value.toLowerCase() }))}
                disabled={isEditMode}
                className={`w-full px-3 py-2 bg-background border rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary ${
                  errors.name ? 'border-destructive' : 'border-input'
                } ${isEditMode ? 'opacity-50 cursor-not-allowed' : ''}`}
                placeholder="my-application"
              />
              {errors.name && <p className="text-destructive text-sm mt-1">{errors.name}</p>}
              <p className="text-xs text-muted-foreground mt-1">
                {t('apps.name_hint')}
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('apps.display_name')} <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                value={formData.display_name}
                onChange={e => setFormData(prev => ({ ...prev, display_name: e.target.value }))}
                className={`w-full px-3 py-2 bg-background border rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary ${
                  errors.display_name ? 'border-destructive' : 'border-input'
                }`}
                placeholder="My Application"
              />
              {errors.display_name && <p className="text-destructive text-sm mt-1">{errors.display_name}</p>}
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('apps.description')}
              </label>
              <textarea
                value={formData.description}
                onChange={e => setFormData(prev => ({ ...prev, description: e.target.value }))}
                rows={3}
                className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder={t('apps.description_placeholder')}
              />
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('apps.homepage')}
              </label>
              <input
                type="url"
                value={formData.homepage_url}
                onChange={e => setFormData(prev => ({ ...prev, homepage_url: e.target.value }))}
                className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="https://myapp.example.com"
              />
            </div>
          </div>
        </div>

        {/* Callback URLs */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-foreground">{t('apps.callback_urls')}</h2>
            <button
              type="button"
              onClick={addCallbackUrl}
              className="flex items-center gap-1 text-sm text-primary hover:text-primary-600"
            >
              <Plus size={16} />
              {t('apps.add_url')}
            </button>
          </div>
          <p className="text-sm text-muted-foreground mb-4">
            {t('apps.callback_hint')}
          </p>
          <div className="space-y-3">
            {formData.callback_urls.map((url, index) => (
              <div key={index} className="flex items-center gap-2">
                <input
                  type="url"
                  value={url}
                  onChange={e => updateCallbackUrl(index, e.target.value)}
                  className="flex-1 px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder="https://myapp.example.com/callback"
                />
                {formData.callback_urls.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeCallbackUrl(index)}
                    className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
                  >
                    <X size={18} />
                  </button>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Auth Methods */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
              <ShieldCheck className="text-primary" size={20} />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-foreground">{t('apps.auth_methods.title')}</h2>
              <p className="text-sm text-muted-foreground">{t('apps.auth_methods.description')}</p>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 mt-4">
            {AUTH_METHODS.map(method => {
              const Icon = method.icon;
              const isSelected = formData.allowed_auth_methods.includes(method.value);
              return (
                <button
                  type="button"
                  key={method.value}
                  onClick={() => toggleAuthMethod(method.value)}
                  className={`flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-colors text-left ${
                    isSelected
                      ? 'border-primary bg-primary/5'
                      : 'border-border hover:border-muted-foreground/30'
                  }`}
                >
                  <Icon size={18} className={isSelected ? 'text-primary' : 'text-muted-foreground'} />
                  <span className={`text-sm font-medium ${isSelected ? 'text-foreground' : 'text-muted-foreground'}`}>
                    {t(method.label)}
                  </span>
                </button>
              );
            })}
          </div>

          {errors.allowed_auth_methods && (
            <p className="text-destructive text-sm mt-2">{errors.allowed_auth_methods}</p>
          )}
        </div>

        {/* Status */}
        {isEditMode && (
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">{t('apps.status')}</h2>
            <div className="flex items-center gap-3">
              <button type="button" onClick={() => setFormData(prev => ({ ...prev, is_active: !prev.is_active }))}
                className={`transition-colors ${formData.is_active ? 'text-success' : 'text-muted-foreground'}`}>
                {formData.is_active ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
              </button>
              <span className="text-sm text-foreground">{t('apps.active')}</span>
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              {t('apps.active_hint')}
            </p>
          </div>
        )}

        {/* Actions */}
        <div className="flex items-center justify-between">
          <div>
            {isEditMode && !application?.is_system && (
              <button
                type="button"
                onClick={handleDelete}
                disabled={deleteApplication.isPending}
                className="px-4 py-2 text-sm font-medium text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
              >
                {deleteApplication.isPending ? t('common.deleting') : t('common.delete')}
              </button>
            )}
          </div>
          <div className="flex items-center gap-3">
            <Link
              to={isEditMode ? `/applications/${id}` : '/applications'}
              className="px-4 py-2 text-sm font-medium text-foreground bg-card border border-input rounded-lg hover:bg-accent transition-colors"
            >
              {t('common.cancel')}
            </Link>
            <button
              type="submit"
              disabled={isPending}
              className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
            >
              {isPending && <Loader2 className="w-4 h-4 animate-spin" />}
              {isPending
                ? t('common.saving')
                : isEditMode
                ? t('common.save')
                : t('apps.create_app')}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
};

export default ApplicationEdit;
