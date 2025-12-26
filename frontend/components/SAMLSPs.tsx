import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Plus, Edit, Trash2, FileText, CheckCircle, XCircle } from 'lucide-react';
import type { SAMLServiceProvider } from '@auth-gateway/client-sdk';
import { useSAMLSPs, useDeleteSAMLSP } from '../hooks/useSAML';

const SAMLSPs: React.FC = () => {
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const navigate = useNavigate();

  const { data, isLoading, error } = useSAMLSPs(page, pageSize);
  const deleteSP = useDeleteSAMLSP();

  const handleDelete = async (id: string, name: string) => {
    if (window.confirm(`Are you sure you want to delete SAML Service Provider "${name}"?`)) {
      try {
        await deleteSP.mutateAsync(id);
      } catch (error) {
        console.error('Failed to delete SAML SP:', error);
        alert('Failed to delete SAML Service Provider');
      }
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <p className="text-red-600">Error loading SAML Service Providers: {(error as Error).message}</p>
      </div>
    );
  }

  const sps = data?.sps || [];

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">SAML Service Providers</h1>
          <p className="text-gray-500 mt-1">Manage SAML 2.0 Service Provider configurations</p>
        </div>
        <div className="flex gap-2">
          <Link
            to="/saml/metadata"
            className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors flex items-center gap-2"
          >
            <FileText size={18} />
            View Metadata
          </Link>
          <button
            onClick={() => navigate('/saml/new')}
            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2"
          >
            <Plus size={18} />
            Create SP
          </button>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Name
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Entity ID
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  ACS URL
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Created
                </th>
                <th scope="col" className="relative px-6 py-3">
                  <span className="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {sps.map((sp) => (
                <tr key={sp.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-gray-900">{sp.name}</div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-gray-500 max-w-xs truncate">{sp.entity_id}</div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-gray-500 max-w-xs truncate">{sp.acs_url}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {sp.is_active ? (
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                        <CheckCircle size={12} />
                        Active
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                        <XCircle size={12} />
                        Inactive
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(sp.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div className="flex justify-end gap-2">
                      <Link
                        to={`/saml/${sp.id}`}
                        className="p-1.5 text-gray-400 hover:text-blue-600 rounded-md hover:bg-gray-100"
                        title="Edit"
                      >
                        <Edit size={16} />
                      </Link>
                      <button
                        onClick={() => handleDelete(sp.id, sp.name)}
                        className="p-1.5 text-gray-400 hover:text-red-600 rounded-md hover:bg-gray-100"
                        title="Delete"
                        disabled={deleteSP.isPending}
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {sps.length === 0 && (
            <div className="p-12 text-center text-gray-500">No SAML Service Providers found.</div>
          )}
        </div>

        {data && data.total > pageSize && (
          <div className="px-6 py-4 border-t border-gray-100 flex items-center justify-between">
            <div className="text-sm text-gray-500">
              Showing {(page - 1) * pageSize + 1} to {Math.min(page * pageSize, data.total)} of {data.total} SPs
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
              >
                Previous
              </button>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page * pageSize >= data.total}
                className="px-3 py-1 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default SAMLSPs;

