import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Search, Plus, Edit, Trash2, Users, Eye } from 'lucide-react';
import type { Group } from '@auth-gateway/client-sdk';
import { useLanguage } from '../services/i18n';
import { useGroups, useDeleteGroup } from '../hooks/useGroups';

const Groups: React.FC = () => {
  const [page, setPage] = useState(1);
  const [searchTerm, setSearchTerm] = useState('');
  const pageSize = 20;
  const navigate = useNavigate();
  const { t } = useLanguage();

  const { data, isLoading, error } = useGroups(page, pageSize);
  const deleteGroup = useDeleteGroup();

  const handleDelete = async (id: string, name: string) => {
    if (window.confirm(`Are you sure you want to delete group "${name}"?`)) {
      try {
        await deleteGroup.mutateAsync(id);
      } catch (error) {
        console.error('Failed to delete group:', error);
        alert('Failed to delete group');
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
        <p className="text-red-600">Error loading groups: {(error as Error).message}</p>
      </div>
    );
  }

  const groups = data?.groups || [];
  const filteredGroups = searchTerm
    ? groups.filter(
        (g) =>
          g.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
          g.display_name.toLowerCase().includes(searchTerm.toLowerCase())
      )
    : groups;

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <h1 className="text-2xl font-bold text-gray-900">Groups</h1>
        <button
          type="button"
          onClick={(e) => {
            e.preventDefault();
            navigate('/groups/new');
          }}
          className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2"
        >
          <Plus size={18} />
          Create Group
        </button>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        {/* Search */}
        <div className="p-4 border-b border-gray-100">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={20} />
            <input
              type="text"
              placeholder="Search groups..."
              className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>

        {/* Table */}
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Name
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Display Name
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Description
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Members
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
              {filteredGroups.map((group) => (
                <tr key={group.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-gray-900">{group.name}</div>
                    {group.is_system_group && (
                      <span className="text-xs text-gray-500">System Group</span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">{group.display_name}</div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-gray-500 max-w-xs truncate">
                      {group.description || '-'}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2 text-sm text-gray-500">
                      <Users size={16} />
                      {group.member_count || 0}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(group.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div className="flex justify-end gap-2">
                      <Link
                        to={`/groups/${group.id}`}
                        className="p-1 text-gray-400 hover:text-blue-600 rounded-md hover:bg-gray-100"
                        title="View Details"
                      >
                        <Eye size={18} />
                      </Link>
                      <Link
                        to={`/groups/${group.id}/edit`}
                        className="p-1 text-gray-400 hover:text-blue-600 rounded-md hover:bg-gray-100"
                        title="Edit"
                      >
                        <Edit size={18} />
                      </Link>
                      {!group.is_system_group && (
                        <button
                          onClick={() => handleDelete(group.id, group.display_name)}
                          className="p-1 text-gray-400 hover:text-red-600 rounded-md hover:bg-gray-100"
                          title="Delete"
                          disabled={deleteGroup.isPending}
                        >
                          <Trash2 size={18} />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {filteredGroups.length === 0 && (
            <div className="p-12 text-center text-gray-500">No groups found.</div>
          )}
        </div>

        {/* Pagination */}
        {data && data.total > pageSize && (
          <div className="px-6 py-4 border-t border-gray-100 flex items-center justify-between">
            <div className="text-sm text-gray-500">
              Showing {(page - 1) * pageSize + 1} to {Math.min(page * pageSize, data.total)} of {data.total} groups
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

export default Groups;

