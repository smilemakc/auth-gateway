import React from 'react';
import {Clock, MapPin, Monitor, Trash2, User} from 'lucide-react';
import {useRevokeSession, useSessions} from '../hooks/useSessions';

const Sessions: React.FC = () => {
    const {data, isLoading, error} = useSessions(1, 100);
    const revokeSession = useRevokeSession();

    const sessions = data?.sessions || [];

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <div
                    className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-8 text-center">
                <p className="text-red-600">Error loading sessions: {(error as Error).message}</p>
            </div>
        );
    }

    const handleRevoke = (sessionId: string) => {
        if (confirm('Are you sure you want to revoke this session?')) {
            revokeSession.mutate(sessionId);
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <h1 className="text-2xl font-bold text-gray-900">Sessions</h1>
                <span className="text-sm text-gray-500">{sessions.length} active sessions</span>
            </div>

            <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
                <div className="overflow-x-auto">
                    <table className="min-w-full text-left text-sm whitespace-nowrap">
                        <thead className="uppercase tracking-wider border-b border-gray-200 bg-gray-50">
                        <tr>
                            <th scope="col" className="px-6 py-4 font-semibold text-gray-700">Session name</th>
                            <th scope="col" className="px-6 py-4 font-semibold text-gray-700">User</th>
                            <th scope="col" className="px-6 py-4 font-semibold text-gray-700">Device</th>
                            <th scope="col" className="px-6 py-4 font-semibold text-gray-700">OS</th>
                            <th scope="col" className="px-6 py-4 font-semibold text-gray-700">IP Address</th>
                            <th scope="col" className="px-6 py-4 font-semibold text-gray-700">User Agent</th>
                            <th scope="col" className="px-6 py-4 font-semibold text-gray-700">Last Active</th>
                            <th scope="col" className="px-6 py-4 font-semibold text-gray-700">Created</th>
                            <th scope="col" className="px-6 py-4 font-semibold text-gray-700">Actions</th>
                        </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200">
                        {sessions.map((session: any) => (
                            <tr key={session.id} className="hover:bg-gray-50">
                                <td className="px-6 py-4">
                                    <div className="flex items-center gap-2 text-gray-600">
                                        <User size={16}/>
                                        {session.session_name || 'Unknown'}
                                    </div>
                                </td>
                                <td className="px-6 py-4">
                                    <div className="flex items-center gap-2 text-gray-600">
                                        <User size={16}/>
                                        {session.user_id || 'Unknown'}
                                    </div>
                                </td>
                                <td className="px-6 py-4">
                                    <div className="flex items-center gap-2 text-gray-600">
                                        <Monitor size={16}/>
                                        <span className="truncate max-w-xs" title={session.device_type}>
                                                {(session.device_type || 'Unknown')}
                                            </span>
                                    </div>
                                </td>
                                <td className="px-6 py-4">
                                    <div className="flex items-center gap-2 text-gray-600">
                                        <Monitor size={16}/>
                                        <span className="truncate max-w-xs" title={session.os}>
                                                {(session.os || 'Unknown')}
                                            </span>
                                    </div>
                                </td>
                                <td className="px-6 py-4">
                                    <div className="flex items-center gap-2 text-gray-500 font-mono text-xs">
                                        <MapPin size={14}/>
                                        {session.ip_address || '-'}
                                    </div>
                                </td>
                                <td className="px-6 py-4">
                                    <div className="flex items-center gap-2 text-gray-600">
                                        <Monitor size={16}/>
                                        <span className="truncate max-w-xs" title={session.user_agent}>
                                                {(session.user_agent || 'Unknown').slice(0, 40)}...
                                            </span>
                                    </div>
                                </td>
                                <td className="px-6 py-4 text-gray-500">
                                    <div className="flex items-center gap-2">
                                        <Clock size={14}/>
                                        {new Date(session.last_active_at || session.updated_at).toLocaleString()}
                                    </div>
                                </td>
                                <td className="px-6 py-4 text-gray-500">
                                    {new Date(session.created_at).toLocaleString()}
                                </td>
                                <td className="px-6 py-4">
                                    <button
                                        onClick={() => handleRevoke(session.id)}
                                        className="text-red-600 hover:text-red-800 p-1 rounded"
                                        title="Revoke session"
                                    >
                                        <Trash2 size={16}/>
                                    </button>
                                </td>
                            </tr>
                        ))}
                        </tbody>
                    </table>
                </div>
                <div className="p-4 border-t border-gray-200 flex justify-between items-center text-sm text-gray-500">
                    <span>Showing {sessions.length} sessions</span>
                </div>
            </div>
        </div>
    );
};

export default Sessions;
