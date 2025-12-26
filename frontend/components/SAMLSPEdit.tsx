import React, { useState, useEffect } from 'react';
import { useNavigate, useParams, Link } from 'react-router-dom';
import { Save, X, Loader, Download, FileText } from 'lucide-react';
import type { CreateSAMLSPRequest, UpdateSAMLSPRequest, SAMLServiceProvider } from '@auth-gateway/client-sdk';
import { useSAMLSP, useCreateSAMLSP, useUpdateSAMLSP } from '../hooks/useSAML';

const SAMLSPEdit: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isNew = !id;

  const { data: sp, isLoading: isLoadingSP } = useSAMLSP(id || '');
  const createSP = useCreateSAMLSP();
  const updateSP = useUpdateSAMLSP();

  const [formData, setFormData] = useState<CreateSAMLSPRequest>({
    name: '',
    entity_id: '',
    acs_url: '',
    slo_url: '',
    x509_cert: '',
    metadata_url: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (sp && !isNew) {
      setFormData({
        name: sp.name,
        entity_id: sp.entity_id,
        acs_url: sp.acs_url,
        slo_url: sp.slo_url || '',
        x509_cert: sp.x509_cert || '',
        metadata_url: sp.metadata_url || '',
      });
    }
  }, [sp, isNew]);

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    }
    if (!formData.entity_id.trim()) {
      newErrors.entity_id = 'Entity ID is required';
    } else if (!formData.entity_id.startsWith('http://') && !formData.entity_id.startsWith('https://')) {
      newErrors.entity_id = 'Entity ID must be a valid URL';
    }
    if (!formData.acs_url.trim()) {
      newErrors.acs_url = 'ACS URL is required';
    } else if (!formData.acs_url.startsWith('http://') && !formData.acs_url.startsWith('https://')) {
      newErrors.acs_url = 'ACS URL must be a valid URL';
    }
    if (formData.slo_url && !formData.slo_url.startsWith('http://') && !formData.slo_url.startsWith('https://')) {
      newErrors.slo_url = 'SLO URL must be a valid URL';
    }
    if (formData.metadata_url && !formData.metadata_url.startsWith('http://') && !formData.metadata_url.startsWith('https://')) {
      newErrors.metadata_url = 'Metadata URL must be a valid URL';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    try {
      if (isNew) {
        await createSP.mutateAsync(formData);
      } else {
        const updateData: UpdateSAMLSPRequest = {
          name: formData.name,
          entity_id: formData.entity_id,
          acs_url: formData.acs_url,
          slo_url: formData.slo_url || undefined,
          x509_cert: formData.x509_cert || undefined,
          metadata_url: formData.metadata_url || undefined,
        };
        await updateSP.mutateAsync({ id: id!, data: updateData });
      }
      navigate('/saml');
    } catch (error) {
      console.error('Failed to save SAML SP:', error);
      alert('Failed to save SAML Service Provider');
    }
  };

  if (isLoadingSP && !isNew) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{isNew ? 'Create SAML Service Provider' : 'Edit SAML Service Provider'}</h1>
        <div className="flex gap-2">
          <Link
            to="/saml/metadata"
            className="px-3 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors flex items-center gap-2 text-sm"
          >
            <Download size={16} />
            Download Metadata
          </Link>
          <button onClick={() => navigate('/saml')} className="text-gray-500 hover:text-gray-700 flex items-center gap-2">
            <X size={20} />
            Cancel
          </button>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 space-y-6">
        {/* Basic Information */}
        <div className="border-b border-gray-200 pb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Basic Information</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Name <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                  errors.name ? 'border-red-300' : 'border-gray-300'
                }`}
                placeholder="Salesforce"
              />
              {errors.name && <p className="mt-1 text-sm text-red-600">{errors.name}</p>}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Entity ID <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={formData.entity_id}
                onChange={(e) => setFormData({ ...formData, entity_id: e.target.value })}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                  errors.entity_id ? 'border-red-300' : 'border-gray-300'
                }`}
                placeholder="https://saml.salesforce.com"
              />
              {errors.entity_id && <p className="mt-1 text-sm text-red-600">{errors.entity_id}</p>}
              <p className="mt-1 text-xs text-gray-500">Unique identifier for this Service Provider</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Assertion Consumer Service (ACS) URL <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={formData.acs_url}
                onChange={(e) => setFormData({ ...formData, acs_url: e.target.value })}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                  errors.acs_url ? 'border-red-300' : 'border-gray-300'
                }`}
                placeholder="https://saml.salesforce.com/sp/ACS"
              />
              {errors.acs_url && <p className="mt-1 text-sm text-red-600">{errors.acs_url}</p>}
              <p className="mt-1 text-xs text-gray-500">Where SAML assertions will be sent</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Single Logout (SLO) URL</label>
              <input
                type="text"
                value={formData.slo_url}
                onChange={(e) => setFormData({ ...formData, slo_url: e.target.value })}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                  errors.slo_url ? 'border-red-300' : 'border-gray-300'
                }`}
                placeholder="https://saml.salesforce.com/sp/SLO"
              />
              {errors.slo_url && <p className="mt-1 text-sm text-red-600">{errors.slo_url}</p>}
              <p className="mt-1 text-xs text-gray-500">Optional: URL for single logout</p>
            </div>
          </div>
        </div>

        {/* Certificate */}
        <div className="border-b border-gray-200 pb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Service Provider Certificate</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">X.509 Certificate (PEM format)</label>
            <textarea
              value={formData.x509_cert}
              onChange={(e) => setFormData({ ...formData, x509_cert: e.target.value })}
              rows={8}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-xs"
              placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
            />
            <p className="mt-1 text-xs text-gray-500">PEM-encoded X.509 certificate from the Service Provider</p>
          </div>
        </div>

        {/* Metadata */}
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Metadata</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Metadata URL</label>
            <input
              type="text"
              value={formData.metadata_url}
              onChange={(e) => setFormData({ ...formData, metadata_url: e.target.value })}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                errors.metadata_url ? 'border-red-300' : 'border-gray-300'
              }`}
              placeholder="https://saml.salesforce.com/metadata"
            />
            {errors.metadata_url && <p className="mt-1 text-sm text-red-600">{errors.metadata_url}</p>}
            <p className="mt-1 text-xs text-gray-500">Optional: URL to fetch SP metadata from</p>
          </div>
        </div>

        <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
          <button
            type="button"
            onClick={() => navigate('/saml')}
            className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={createSP.isPending || updateSP.isPending}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {(createSP.isPending || updateSP.isPending) && <Loader size={16} className="animate-spin" />}
            <Save size={16} />
            {isNew ? 'Create SP' : 'Save Changes'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default SAMLSPEdit;

