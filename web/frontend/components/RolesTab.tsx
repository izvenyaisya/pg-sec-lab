'use client';

import { PolicyReport, RoleInfo } from '@/types/report';
import { User, Shield, Lock } from 'lucide-react';

interface Props {
  report: PolicyReport;
}

export default function RolesTab({ report }: Props) {
  return (
    <div className="bg-white rounded-lg shadow overflow-hidden">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Role Name</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Login</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Superuser</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Bypass RLS</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Grants</th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {report.roles.map((role) => (
            <RoleRow key={role.name} role={role} />
          ))}
        </tbody>
      </table>
    </div>
  );
}

function RoleRow({ role }: { role: RoleInfo }) {
  const isDangerous = role.superuser || role.bypassrls;

  return (
    <tr className={isDangerous ? 'bg-red-50' : ''}>
      <td className="px-6 py-4 whitespace-nowrap">
        <div className="flex items-center">
          <User className="mr-2" size={16} />
          <span className="font-medium">{role.name}</span>
          {role.superuser && <Shield className="ml-2 text-red-600" size={16} />}
          {role.bypassrls && <Lock className="ml-2 text-orange-600" size={16} />}
        </div>
      </td>
      <td className="px-6 py-4"><Badge active={role.login} /></td>
      <td className="px-6 py-4"><Badge active={role.superuser} danger /></td>
      <td className="px-6 py-4"><Badge active={role.bypassrls} danger /></td>
      <td className="px-6 py-4 text-sm text-gray-600">{role.grants.length} grants</td>
    </tr>
  );
}

function Badge({ active, danger }: { active: boolean; danger?: boolean }) {
  const className = active
    ? danger ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800'
    : 'bg-gray-100 text-gray-800';
  
  return <span className={`px-2 py-1 text-xs font-semibold rounded-full ${className}`}>{active ? 'Yes' : 'No'}</span>;
}
