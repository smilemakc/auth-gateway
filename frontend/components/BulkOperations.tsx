import React from 'react';
import { Link } from 'react-router-dom';
import { Users, UserPlus, UserMinus, UserCheck, FileSpreadsheet } from 'lucide-react';

const BulkOperations: React.FC = () => {
  const operations = [
    {
      id: 'create',
      title: 'Bulk Create Users',
      description: 'Create multiple users at once from CSV or JSON file',
      icon: UserPlus,
      path: '/bulk/create',
      color: 'bg-blue-500',
    },
    {
      id: 'update',
      title: 'Bulk Update Users',
      description: 'Update multiple users simultaneously',
      icon: Users,
      path: '/bulk/update',
      color: 'bg-green-500',
    },
    {
      id: 'delete',
      title: 'Bulk Delete Users',
      description: 'Delete multiple users at once',
      icon: UserMinus,
      path: '/bulk/delete',
      color: 'bg-red-500',
    },
    {
      id: 'assign-roles',
      title: 'Bulk Assign Roles',
      description: 'Assign roles to multiple users',
      icon: UserCheck,
      path: '/bulk/assign-roles',
      color: 'bg-purple-500',
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">Bulk Operations</h1>
        <p className="text-muted-foreground mt-1">Perform operations on multiple users at once</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {operations.map((op) => {
          const Icon = op.icon;
          return (
            <Link
              key={op.id}
              to={op.path}
              className="bg-card rounded-xl shadow-sm border border-border p-6 hover:shadow-md transition-shadow"
            >
              <div className="flex items-start gap-4">
                <div className={`${op.color} p-3 rounded-lg`}>
                  <Icon className="text-white" size={24} />
                </div>
                <div className="flex-1">
                  <h3 className="text-lg font-semibold text-foreground mb-1">{op.title}</h3>
                  <p className="text-sm text-muted-foreground">{op.description}</p>
                </div>
              </div>
            </Link>
          );
        })}
      </div>

      <div className="bg-primary/10 border border-border rounded-lg p-4">
        <h3 className="font-semibold text-primary mb-2 flex items-center gap-2">
          <FileSpreadsheet size={20} />
          CSV/JSON Format
        </h3>
        <p className="text-sm text-primary">
          For bulk create operations, you can upload a CSV or JSON file. The CSV should have columns: email, username,
          full_name, password, is_active, email_verified
        </p>
      </div>
    </div>
  );
};

export default BulkOperations;

