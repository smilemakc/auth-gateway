import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { Loader2 } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { LoadingSpinner, PageHeader } from '../ui';
import {
  useApplicationDetail,
  useCreateApplication,
  useUpdateApplication,
  useDeleteApplication,
} from '../../hooks/useApplications';
import { confirm } from '../../services/confirm';
import { logger } from '@/lib/logger';
import ApplicationEditBasicInfoSection from './ApplicationEditBasicInfoSection';
import ApplicationEditCallbackUrlsSection from './ApplicationEditCallbackUrlsSection';
import ApplicationEditAuthMethodsSection from './ApplicationEditAuthMethodsSection';

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
      logger.error('Failed to save application:', error);
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
        logger.error('Failed to delete application:', error);
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
    return <LoadingSpinner />;
  }

  const isPending = createApplication.isPending || updateApplication.isPending;

  return (
    <div className="space-y-6">
      <PageHeader
        title={isEditMode ? t('apps.edit') : t('apps.create')}
        subtitle={isEditMode ? t('apps.edit_desc') : t('apps.create_desc')}
        onBack={() => navigate(isEditMode ? `/applications/${id}` : '/applications')}
      />

      <form onSubmit={handleSubmit} className="space-y-6">
        <ApplicationEditBasicInfoSection
          name={formData.name}
          displayName={formData.display_name}
          description={formData.description}
          homepageUrl={formData.homepage_url}
          isEditMode={!!isEditMode}
          errors={errors}
          onNameChange={value => setFormData(prev => ({ ...prev, name: value }))}
          onDisplayNameChange={value => setFormData(prev => ({ ...prev, display_name: value }))}
          onDescriptionChange={value => setFormData(prev => ({ ...prev, description: value }))}
          onHomepageUrlChange={value => setFormData(prev => ({ ...prev, homepage_url: value }))}
        />

        <ApplicationEditCallbackUrlsSection
          callbackUrls={formData.callback_urls}
          onAdd={addCallbackUrl}
          onRemove={removeCallbackUrl}
          onUpdate={updateCallbackUrl}
        />

        <ApplicationEditAuthMethodsSection
          selectedMethods={formData.allowed_auth_methods}
          isActive={formData.is_active}
          isEditMode={!!isEditMode}
          error={errors.allowed_auth_methods}
          onToggleMethod={toggleAuthMethod}
          onToggleActive={() => setFormData(prev => ({ ...prev, is_active: !prev.is_active }))}
        />

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
