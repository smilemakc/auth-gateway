import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { ArrowLeft, Loader2, Plus, X, Boxes } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import {
  useApplicationDetail,
  useCreateApplication,
  useUpdateApplication,
  useDeleteApplication,
} from '../hooks/useApplications';

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
        });
        navigate(`/applications/${result.id}`);
      }
    } catch (error) {
      console.error('Failed to save application:', error);
    }
  };

  const handleDelete = async () => {
    if (window.confirm(t('common.confirm_delete'))) {
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

        {/* Status */}
        {isEditMode && (
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">{t('apps.status')}</h2>
            <label className="flex items-center gap-3">
              <input
                type="checkbox"
                checked={formData.is_active}
                onChange={e => setFormData(prev => ({ ...prev, is_active: e.target.checked }))}
                className="w-4 h-4 rounded border-input text-primary focus:ring-primary"
              />
              <span className="text-sm text-foreground">{t('apps.active')}</span>
            </label>
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
